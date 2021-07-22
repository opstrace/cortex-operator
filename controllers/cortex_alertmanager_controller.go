/**
 * Copyright 2021 Opstrace, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cortexv1alpha1 "github.com/opstrace/cortex-operator/api/v1alpha1"
)

const AlertManagerName = "alertmanager"

// CortexAlertManagerReconciler reconciles a Cortex object and ensures the Cortex
// Alert Manager is deployed
type CortexAlertManagerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cortex.opstrace.io,resources=cortices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cortex.opstrace.io,resources=cortices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cortex.opstrace.io,resources=cortices/finalizers,verbs=update

//
// Setup RBAC to create and manage Kubernetes resources required to deploy Cortex.
//

//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete;scope=Cluster
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete;scope=Cluster
//+kubebuilder:rbac:groups="apps",resources=statefulsets,verbs=get;list;watch;create;update;patch;delete;scope=Cluster

//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *CortexAlertManagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("cortex", req.NamespacedName)

	cortex := &cortexv1alpha1.Cortex{}
	if err := r.Get(ctx, req.NamespacedName, cortex); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "unable to fetch Cortex")
		}
		// Ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get
		// them on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	krr := KubernetesResourceReconciler{
		scheme: r.Scheme,
		client: r.Client,
		cortex: cortex,
		log:    log,
	}

	svc := NewAlertManagerService(req)
	err := krr.Reconcile(ctx, svc)
	if err != nil {
		return ctrl.Result{}, err
	}

	deploy := NewAlertManagerDeployment(req, cortex, cortex.Spec.AlertManagerSpec)
	cortex.Status.AlertManagerRef = deploy.ref
	err = krr.Reconcile(ctx, deploy)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CortexAlertManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cortexv1alpha1.Cortex{}).
		Complete(r)
}

func NewAlertManagerService(req ctrl.Request) *KubernetesResource {
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: AlertManagerName, Namespace: req.Namespace}}
	ref := &corev1.LocalObjectReference{Name: AlertManagerName}
	return &KubernetesResource{
		obj: svc,
		ref: ref,
		mutator: func() error {
			svc.Labels = map[string]string{
				"name": AlertManagerName,
				"job":  fmt.Sprintf("%s.%s", req.Namespace, AlertManagerName),
			}
			svc.Spec.Ports = make([]corev1.ServicePort, 0)
			svc.Spec.Ports = []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
				{
					Name:       AlertManagerName,
					Port:       9094,
					TargetPort: intstr.FromInt(9094),
				},
			}
			svc.Spec.Selector = map[string]string{"name": AlertManagerName}

			return nil
		},
	}
}

func NewAlertManagerDeployment(
	req ctrl.Request,
	cortex *cortexv1alpha1.Cortex,
	spec *cortexv1alpha1.DeploymentSpec,
) *KubernetesResource {
	deploy := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: AlertManagerName, Namespace: req.Namespace}}
	labels := map[string]string{
		"name": AlertManagerName,
	}
	annotations := map[string]string{
		CortexConfigShasumAnnotationName: cortex.Spec.ConfigSHA(),
	}
	ref := &corev1.LocalObjectReference{Name: AlertManagerName}

	configMapName := cortex.Name + CortexConfigMapNameSuffix
	runtimeConfigMapName := cortex.Name + CortexRuntimeConfigMapNameSuffix

	return &KubernetesResource{
		obj: deploy,
		ref: ref,
		mutator: func() error {
			deploy.Spec.Replicas = spec.Replicas
			deploy.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: labels,
			}
			deploy.Spec.Template.Labels = labels
			deploy.Spec.Template.Annotations = annotations
			deploy.Spec.Template.Spec.Affinity = WithPodAntiAffinity(AlertManagerName)
			deploy.Spec.Template.Spec.Containers = []corev1.Container{
				{
					Name:            AlertManagerName,
					Image:           cortex.Spec.Image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Args: []string{
						"-target=" + AlertManagerName,
						"-config.file=/etc/cortex/config.yaml",
					},
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: 80,
						},
						{
							Name:          AlertManagerName,
							ContainerPort: 9094,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							MountPath: "/etc/cortex",
							Name:      "cortex",
						},
					},
				},
			}
			deploy.Spec.Template.Spec.ServiceAccountName = ServiceAccountName
			deploy.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name: "cortex",
					VolumeSource: corev1.VolumeSource{
						Projected: &corev1.ProjectedVolumeSource{
							Sources: []corev1.VolumeProjection{
								{
									ConfigMap: &corev1.ConfigMapProjection{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: configMapName,
										},
									},
								},
								{
									ConfigMap: &corev1.ConfigMapProjection{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: runtimeConfigMapName,
										},
									},
								},
							},
						},
					},
				},
			}

			return nil
		},
	}
}

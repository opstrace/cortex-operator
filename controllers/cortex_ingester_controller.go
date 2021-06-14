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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cortexv1alpha1 "github.com/opstrace/cortex-operator/api/v1alpha1"
)

// CortexIngesterReconciler reconciles a Cortex object and ensures the Cortex
// Ingesters are deployed
type CortexIngesterReconciler struct {
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
func (r *CortexIngesterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	svc := NewService(req, "ingester")
	err := krr.Reconcile(ctx, svc)
	if err != nil {
		return ctrl.Result{}, err
	}

	sts := NewIngesterStatefulSet(req, "ingester", cortex)
	cortex.Status.IngesterRef = sts.ref
	err = krr.Reconcile(ctx, sts)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CortexIngesterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cortexv1alpha1.Cortex{}).
		Complete(r)
}

func NewIngesterStatefulSet(
	req ctrl.Request,
	name string,
	cortex *cortexv1alpha1.Cortex,
) *KubernetesResource {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: "ingester", Namespace: req.Namespace},
	}
	labels := map[string]string{
		"name": "ingester",
	}
	annotations := map[string]string{
		CortexConfigShasumAnnotationName: cortex.Spec.ConfigSHA(),
	}
	ref := &corev1.LocalObjectReference{Name: name}

	return &KubernetesResource{
		obj: sts,
		ref: ref,
		mutator: func() error {
			sts.Spec.ServiceName = "ingester"
			sts.Spec.Replicas = cortex.Spec.IngesterSpec.Replicas
			sts.Spec.PodManagementPolicy = appsv1.ParallelPodManagement
			sts.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: map[string]string{"name": "ingester"},
			}
			sts.Spec.Template.Labels = labels
			sts.Spec.Template.Annotations = annotations
			sts.Spec.Template.Spec.Affinity = WithPodAntiAffinity("ingester")
			sts.Spec.Template.Spec.Containers = []corev1.Container{
				{
					Name:  "ingester",
					Image: cortex.Spec.Image,
					Args: []string{
						"-target=ingester",
						"-ingester.chunk-encoding=3",
						"-config.file=/etc/cortex/config.yaml",
					},
					ImagePullPolicy: corev1.PullIfNotPresent,
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: 80,
						},
						{
							Name:          "grpc",
							ContainerPort: 9095,
						},
					},
					ReadinessProbe: &corev1.Probe{
						Handler: corev1.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Path:   "/ready",
								Port:   intstr.FromInt(80),
								Scheme: corev1.URISchemeHTTP,
							},
						},
						InitialDelaySeconds: 45,
						TimeoutSeconds:      1,
						PeriodSeconds:       10,
						SuccessThreshold:    1,
						FailureThreshold:    3,
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							MountPath: "/etc/cortex",
							Name:      "cortex",
						},
						{
							Name:      "datadir",
							MountPath: "/cortex",
						},
					},
				},
			}
			sts.Spec.Template.Spec.ServiceAccountName = ServiceAccountName
			// https://cortexmetrics.io/docs/guides/running-cortex-on-kubernetes/#take-extra-care-with-ingesters
			sts.Spec.Template.Spec.TerminationGracePeriodSeconds = pointer.Int64Ptr(2400)
			sts.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name: "cortex",
					VolumeSource: corev1.VolumeSource{
						Projected: &corev1.ProjectedVolumeSource{
							Sources: []corev1.VolumeProjection{
								{
									ConfigMap: &corev1.ConfigMapProjection{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: CortexConfigMapName,
										},
									},
								},
								{
									ConfigMap: &corev1.ConfigMapProjection{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: CortexRuntimeConfigMapName,
										},
									},
								},
							},
						},
					},
				},
				{
					Name: "datadir",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "datadir",
						},
					},
				},
			}
			sts.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "datadir"},
					Spec: corev1.PersistentVolumeClaimSpec{
						// Uses the default storage class.
						StorageClassName: cortex.Spec.IngesterSpec.StorageClassName,
						AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"storage": *cortex.Spec.IngesterSpec.DatadirSize,
							},
						},
					},
				},
			}

			return nil
		},
	}
}

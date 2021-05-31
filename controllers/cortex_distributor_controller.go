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
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cortexv1alpha1 "github.com/opstrace/cortex-operator/api/v1alpha1"
)

// CortexDistributorReconciler reconciles a Cortex object and ensures the Cortex
// Distributors are deployed
type CortexDistributorReconciler struct {
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
func (r *CortexDistributorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	if cortex.Status.IngesterRef == nil {
		log.Info("waiting for ingester")
		return ctrl.Result{}, nil
	}

	if cortex.Status.DistributorRef == nil {
		cortex.Status.DistributorRef = &cortexv1alpha1.DistributorReference{}
	}

	krr := KubernetesResourceReconciler{
		scheme: r.Scheme,
		client: r.Client,
		cortex: cortex,
		log:    log,
	}

	svc := NewService(req, "distributor")
	cortex.Status.DistributorRef.Svc = svc.ref
	err := krr.Reconcile(ctx, svc)
	if err != nil {
		return ctrl.Result{}, err
	}

	deploy := NewDeployment(req, "distributor", cortex)
	cortex.Status.DistributorRef.Deploy = deploy.ref
	err = krr.Reconcile(ctx, deploy)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CortexDistributorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cortexv1alpha1.Cortex{}).
		Complete(r)
}

func NewDeployment(
	req ctrl.Request,
	name string,
	cortex *cortexv1alpha1.Cortex,
) *KubernetesResource {
	deploy := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: req.Namespace}}
	labels := map[string]string{
		"name":       name,
		"memberlist": "gossip-ring",
	}
	annotations := map[string]string{
		CortexConfigShasumAnnotationName: cortex.Spec.ConfigSHA(),
	}
	ref := &corev1.LocalObjectReference{Name: name}

	return &KubernetesResource{
		obj: deploy,
		ref: ref,
		mutator: func() error {
			deploy.Spec.Replicas = pointer.Int32Ptr(2)
			deploy.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: labels,
			}
			deploy.Spec.Template.Labels = labels
			deploy.Spec.Template.Annotations = annotations
			deploy.Spec.Template.Spec.Affinity = WithPodAntiAffinity(name)
			deploy.Spec.Template.Spec.Containers = []corev1.Container{
				{
					Name:            name,
					Image:           cortex.Spec.Image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Args: []string{
						"-target=" + name,
						"-config.file=/etc/cortex/config.yaml",
					},
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
					VolumeMounts: []corev1.VolumeMount{
						{
							MountPath: "/etc/cortex",
							Name:      "cortex",
						},
					},
				},
			}
			deploy.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name: "cortex",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "cortex",
							},
						},
					},
				},
			}

			return nil
		},
	}
}

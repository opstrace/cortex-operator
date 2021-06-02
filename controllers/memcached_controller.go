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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cortexv1alpha1 "github.com/opstrace/cortex-operator/api/v1alpha1"
)

// MemacachedReconciler reconciles a Cortex object and ensures all Memcached
// resources are deployed
type MemcachedReconciler struct {
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
func (r *MemcachedReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	if cortex.Status.MemcachedRef == nil {
		cortex.Status.MemcachedRef = &cortexv1alpha1.MemcachedReference{}
	}

	krr := KubernetesResourceReconciler{
		scheme: r.Scheme,
		client: r.Client,
		cortex: cortex,
		log:    log,
	}

	svc := NewMemcachedService(req, "memcached-chunks")
	err := krr.Reconcile(ctx, svc)
	if err != nil {
		return ctrl.Result{}, err
	}

	sts := NewMemcachedStatefulSet(
		req,
		"memcached-chunks",
		cortex.Spec.Memcached.Image,
		cortex.Spec.Memcached.ChunksCacheSpec,
	)
	cortex.Status.MemcachedRef.ChunksCacheRef = sts.ref
	err = krr.Reconcile(ctx, sts)
	if err != nil {
		return ctrl.Result{}, err
	}

	svc = NewMemcachedService(req, "memcached-index-queries")
	err = krr.Reconcile(ctx, svc)
	if err != nil {
		return ctrl.Result{}, err
	}

	sts = NewMemcachedStatefulSet(
		req,
		"memcached-index-queries",
		cortex.Spec.Memcached.Image,
		cortex.Spec.Memcached.IndexQueriesCacheSpec,
	)
	cortex.Status.MemcachedRef.IndexQueriesCacheRef = sts.ref
	err = krr.Reconcile(ctx, sts)
	if err != nil {
		return ctrl.Result{}, err
	}

	svc = NewMemcachedService(req, "memcached-index-writes")
	err = krr.Reconcile(ctx, svc)
	if err != nil {
		return ctrl.Result{}, err
	}

	sts = NewMemcachedStatefulSet(
		req,
		"memcached-index-writes",
		cortex.Spec.Memcached.Image,
		cortex.Spec.Memcached.IndexWritesCacheSpec,
	)
	cortex.Status.MemcachedRef.IndexWritesCacheRef = sts.ref
	err = krr.Reconcile(ctx, sts)
	if err != nil {
		return ctrl.Result{}, err
	}

	svc = NewMemcachedService(req, "memcached-results")
	err = krr.Reconcile(ctx, svc)
	if err != nil {
		return ctrl.Result{}, err
	}

	sts = NewMemcachedStatefulSet(
		req,
		"memcached-results",
		cortex.Spec.Memcached.Image,
		cortex.Spec.Memcached.ResultsCacheSpec,
	)
	cortex.Status.MemcachedRef.ResultsCacheRef = svc.ref
	err = krr.Reconcile(ctx, sts)
	if err != nil {
		return ctrl.Result{}, err
	}

	svc = NewMemcachedService(req, "memcached-metadata")
	err = krr.Reconcile(ctx, svc)
	if err != nil {
		return ctrl.Result{}, err
	}

	sts = NewMemcachedStatefulSet(
		req,
		"memcached-metadata",
		cortex.Spec.Memcached.Image,
		cortex.Spec.Memcached.MetadataCacheSpec,
	)
	cortex.Status.MemcachedRef.MetadataCacheRef = sts.ref
	err = krr.Reconcile(ctx, sts)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MemcachedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cortexv1alpha1.Cortex{}).
		Complete(r)
}

func NewMemcachedService(req ctrl.Request, name string) *KubernetesResource {
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: req.Namespace}}
	return &KubernetesResource{
		obj: svc,
		mutator: func() error {
			svc.Labels = map[string]string{
				"name": name,
			}
			svc.Spec.Ports = make([]corev1.ServicePort, 0)
			svc.Spec.ClusterIP = "None"
			svc.Spec.Ports = []corev1.ServicePort{
				{
					Name:       "memcached-client",
					Port:       11211,
					TargetPort: intstr.FromInt(11211),
				},
			}
			svc.Spec.Selector = map[string]string{"name": name}

			return nil
		},
	}
}

func NewMemcachedStatefulSet(
	req ctrl.Request,
	name string,
	image string,
	spec *cortexv1alpha1.MemcachedStatefulSetSpec,
) *KubernetesResource {
	statefulSet := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: req.Namespace}}
	ref := &corev1.LocalObjectReference{Name: name}
	args := append([]string{"-v"}, spec.AsArgs()...)

	return &KubernetesResource{
		obj: statefulSet,
		ref: ref,
		mutator: func() error {
			statefulSet.Spec.ServiceName = name
			statefulSet.Spec.Replicas = spec.Replicas
			statefulSet.Spec.PodManagementPolicy = appsv1.ParallelPodManagement
			statefulSet.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: map[string]string{"name": name},
			}
			statefulSet.Spec.Template.ObjectMeta.Labels = map[string]string{
				"name": name,
			}
			statefulSet.Spec.Template.Spec.Affinity = WithPodAntiAffinity(name)
			statefulSet.Spec.Template.Spec.Containers = []corev1.Container{
				{
					Name:            "memcached",
					Image:           image,
					Args:            args,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Ports: []corev1.ContainerPort{
						{
							Name:          "client",
							ContainerPort: 11211,
						},
					},
				},
			}
			statefulSet.Spec.UpdateStrategy = appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			}

			return nil
		},
	}
}

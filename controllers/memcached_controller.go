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
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	cortexv1alpha1 "github.com/opstrace/cortex-operator/api/v1alpha1"
)

// CortexReconciler reconciles a Cortex object
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

	svc := NewMemcachedService(req, "memcached")
	cortex.Status.MemcachedRef.MemcachedSvc = svc.ref
	err := krr.Reconcile(ctx, svc)
	if err != nil {
		return ctrl.Result{}, err
	}

	// o = makeStatefulSetMemcached(req, "memcached")
	// resources = append(resources, o)

	svc = NewMemcachedService(req, "memcached-index-queries")
	cortex.Status.MemcachedRef.MemcachedIndexQueriesSvc = svc.ref
	err = krr.Reconcile(ctx, svc)
	if err != nil {
		return ctrl.Result{}, err
	}

	// o = makeStatefulSetMemcached(req, "memcached-index-queries")
	// resources = append(resources, o)

	// o = makeHeadlessService(req, "memcached-index-writes", "name", servicePort{"memcached-client", 11211})
	// resources = append(resources, o)
	svc = NewMemcachedService(req, "memcached-index-writes")
	cortex.Status.MemcachedRef.MemcachedIndexWritesSvc = svc.ref
	err = krr.Reconcile(ctx, svc)
	if err != nil {
		return ctrl.Result{}, err
	}

	// o = makeStatefulSetMemcached(req, "memcached-index-writes")
	// resources = append(resources, o)

	// o = makeHeadlessService(req, "memcached-results", "name", servicePort{"memcached-client", 11211})
	// resources = append(resources, o)
	svc = NewMemcachedService(req, "memcached-results")
	cortex.Status.MemcachedRef.MemcachedResultsSvc = svc.ref
	err = krr.Reconcile(ctx, svc)
	if err != nil {
		return ctrl.Result{}, err
	}

	// o = makeStatefulSetMemcached(req, "memcached-results")
	// resources = append(resources, o)

	// o = makeHeadlessService(req, "memcached-metadata", "name", servicePort{"memcached-client", 11211})
	// resources = append(resources, o)
	svc = NewMemcachedService(req, "memcached-metadata")
	cortex.Status.MemcachedRef.MemcachedMetadataSvc = svc.ref
	err = krr.Reconcile(ctx, svc)
	if err != nil {
		return ctrl.Result{}, err
	}

	// o = makeStatefulSetMemcached(req, "memcached-metadata")
	// resources = append(resources, o)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MemcachedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cortexv1alpha1.Cortex{}).
		Complete(r)
}

type KubernetesResource struct {
	obj     client.Object
	ref     *corev1.LocalObjectReference
	mutator controllerutil.MutateFn
}

type KubernetesResourceReconciler struct {
	log    logr.Logger
	client client.Client
	cortex *cortexv1alpha1.Cortex
	scheme *runtime.Scheme
}

func (krr *KubernetesResourceReconciler) Reconcile(
	ctx context.Context,
	r *KubernetesResource,
) error {
	// Set up garbage collection. The object (resource.obj) will be
	// automatically deleted whent he owner (cortex) is deleted.
	err := controllerutil.SetOwnerReference(krr.cortex, r.obj, krr.scheme)
	if err != nil {
		krr.log.Error(
			err,
			"failed to set owner reference on resource",
			"kind", r.obj.GetObjectKind().GroupVersionKind().Kind,
			"name", r.obj.GetName(),
		)
		return err
	}

	op, err := controllerutil.CreateOrUpdate(ctx, krr.client, r.obj, r.mutator)
	if err != nil {
		krr.log.Error(
			err,
			"failed to reconcile resource",
			"kind", r.obj.GetObjectKind().GroupVersionKind().Kind,
			"name", r.obj.GetName(),
		)
		return err
	}

	err = krr.client.Status().Update(ctx, krr.cortex)
	if err != nil {
		krr.log.Error(
			err,
			"failed to reconcile resource status",
			"kind", r.obj.GetObjectKind().GroupVersionKind().Kind,
			"name", r.obj.GetName(),
		)
		return err
	}

	krr.log.Info(
		"Reconcile successful",
		"operation", op,
		"kind", r.obj.GetObjectKind().GroupVersionKind().Kind,
		"name", r.obj.GetName(),
	)

	return nil
}

func NewMemcachedService(req ctrl.Request, name string) *KubernetesResource {
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: req.Namespace}}
	ref := &corev1.LocalObjectReference{Name: name}
	return &KubernetesResource{
		obj: svc,
		ref: ref,
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

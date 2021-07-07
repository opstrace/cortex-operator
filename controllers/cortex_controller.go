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
type CortexReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

type kubernetesResource struct {
	obj     client.Object
	mutator controllerutil.MutateFn
}

const FinalizerName = "cortex.opstrace.io/finalizer"
const ServiceAccountName = "cortex"
const CortexConfigShasumAnnotationName = "cortex-operator/cortex-config-shasum"
const CortexConfigMapName = "cortex-config"
const CortexRuntimeConfigMapName = "cortex-runtime-config"
const GossipRingServiceName = "gossip-ring"

//+kubebuilder:rbac:groups=cortex.opstrace.io,resources=cortices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cortex.opstrace.io,resources=cortices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cortex.opstrace.io,resources=cortices/finalizers,verbs=update

//
// Setup RBAC to create and manage Kubernetes resources required to deploy Cortex.
//

//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete;scope=Cluster
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete;scope=Cluster
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete;scope=Cluster
//+kubebuilder:rbac:groups="apps",resources=statefulsets,verbs=get;list;watch;create;update;patch;delete;scope=Cluster
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete;scope=Cluster

//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *CortexReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	cm := NewCortexConfigMap(req, cortex)
	err := krr.Reconcile(ctx, cm)
	if err != nil {
		return ctrl.Result{}, err
	}

	sa := NewServiceAccount(req, cortex)
	err = krr.Reconcile(ctx, sa)
	if err != nil {
		return ctrl.Result{}, err
	}

	svc := NewGossipRingService(req)
	err = krr.Reconcile(ctx, svc)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CortexReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cortexv1alpha1.Cortex{}).
		Complete(r)
}

func NewCortexConfigMap(
	req ctrl.Request,
	cortex *cortexv1alpha1.Cortex,
) *KubernetesResource {
	// Validation errors were handled by the webhooks.
	y, _ := cortex.AsYAML()
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CortexConfigMapName,
			Namespace: req.Namespace,
		},
	}

	return &KubernetesResource{
		obj: configMap,
		mutator: func() error {
			configMap.Data = map[string]string{
				"config.yaml": string(y),
			}
			return nil
		},
	}
}

func NewServiceAccount(req ctrl.Request, cortex *cortexv1alpha1.Cortex) *KubernetesResource {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceAccountName,
			Namespace: req.Namespace,
		},
	}

	return &KubernetesResource{
		obj: serviceAccount,
		mutator: func() error {
			// nothing to do if nothing is set
			if cortex.Spec.ServiceAccountSpec == nil {
				return nil
			}

			serviceAccount.Annotations = cortex.Spec.ServiceAccountSpec.Annotations
			if len(cortex.Spec.ServiceAccountSpec.ImagePullSecrets) > 0 {
				serviceAccount.ImagePullSecrets = make(
					[]corev1.LocalObjectReference,
					len(cortex.Spec.ServiceAccountSpec.ImagePullSecrets),
				)
				copy(
					serviceAccount.ImagePullSecrets,
					cortex.Spec.ServiceAccountSpec.ImagePullSecrets,
				)
			}

			return nil
		},
	}
}

func NewGossipRingService(req ctrl.Request) *KubernetesResource {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GossipRingServiceName,
			Namespace: req.Namespace,
		},
	}

	return &KubernetesResource{
		obj: service,
		mutator: func() error {
			service.Labels = map[string]string{
				"name": GossipRingServiceName,
			}
			service.Spec.Ports = make([]corev1.ServicePort, 0)
			service.Spec.ClusterIP = "None"
			service.Spec.Ports = []corev1.ServicePort{
				{
					Name:       GossipRingServiceName,
					Port:       int32(7946),
					TargetPort: intstr.FromInt(7946),
				},
			}
			service.Spec.Selector = map[string]string{"memberlist": GossipRingServiceName}
			return nil
		},
	}
}

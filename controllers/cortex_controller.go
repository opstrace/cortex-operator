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
	"crypto/sha256"
	"encoding/hex"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"

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
const CortexConfigMapName = "cortex"

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

	resources := make([]*kubernetesResource, 0)

	cortexConfigStr, _, err := generateCortexConfig(cortex)
	if err != nil {
		log.Error(err, "failed to generate cortex configmap, will not retry")
		return ctrl.Result{Requeue: false}, nil
	}

	o := makeCortexConfigMap(req, cortexConfigStr)
	resources = append(resources, o)

	o = makeServiceAccount(req)
	resources = append(resources, o)

	o = makeHeadlessService(req, "gossip-ring", "memberlist", servicePort{"gossip-ring", 7946})
	resources = append(resources, o)

	for _, resource := range resources {
		// Set up garbage collection. The object (resource.obj) will be
		// automatically deleted whent he owner (cortex) is deleted.
		err := controllerutil.SetOwnerReference(cortex, resource.obj, r.Scheme)
		if err != nil {
			log.Error(
				err,
				"failed to set owner reference on resource",
				"kind", resource.obj.GetObjectKind().GroupVersionKind().Kind,
				"name", resource.obj.GetName(),
			)
			return ctrl.Result{}, err
		}

		op, err := controllerutil.CreateOrUpdate(ctx, r.Client, resource.obj, resource.mutator)
		if err != nil {
			log.Error(
				err,
				"failed to reconcile resource",
				"kind", resource.obj.GetObjectKind().GroupVersionKind().Kind,
				"name", resource.obj.GetName(),
			)
			return ctrl.Result{}, err
		}

		log.Info(
			"Reconcile successful",
			"operation", op,
			"kind", resource.obj.GetObjectKind().GroupVersionKind().Kind,
			"name", resource.obj.GetName(),
		)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CortexReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cortexv1alpha1.Cortex{}).
		Complete(r)
}

// generateCortexConfig returns a config yaml and the shasum of it or an error
// when templating the given cortex spec to yaml config accepted by cortex
// components.
func generateCortexConfig(cortex *cortexv1alpha1.Cortex) (string, string, error) {
	sha := sha256.Sum256(cortex.Spec.Config.Raw)
	s := hex.EncodeToString(sha[:])

	y, err := yaml.JSONToYAML(cortex.Spec.Config.Raw)
	if err != nil {
		return "", "", err
	}

	return string(y), s, nil
}

func makeCortexConfigMap(req ctrl.Request, cortexConfigStr string) *kubernetesResource {
	configMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: CortexConfigMapName, Namespace: req.Namespace}}

	return &kubernetesResource{
		obj: configMap,
		mutator: func() error {
			configMap.Data = map[string]string{
				"config.yaml": string(cortexConfigStr),
			}
			return nil
		},
	}
}

func makeServiceAccount(req ctrl.Request) *kubernetesResource {
	serviceAccount := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: ServiceAccountName, Namespace: req.Namespace}}

	return &kubernetesResource{
		obj: serviceAccount,
		mutator: func() error {
			return nil
		},
	}
}

type servicePort struct {
	Name string
	Port int
}

func makeHeadlessService(req ctrl.Request, name string, selector string, servicePorts ...servicePort) *kubernetesResource {
	service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: req.Namespace}}

	return &kubernetesResource{
		obj: service,
		mutator: func() error {
			service.Labels = map[string]string{
				"name": name,
			}
			service.Spec.Ports = make([]corev1.ServicePort, 0)
			service.Spec.ClusterIP = "None"
			for _, p := range servicePorts {
				service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{
					Name:       p.Name,
					Port:       int32(p.Port),
					TargetPort: intstr.FromInt(p.Port),
				})
			}
			service.Spec.Selector = map[string]string{selector: name}

			return nil
		},
	}
}

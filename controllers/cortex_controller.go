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
	"bytes"
	"context"
	"fmt"
	"html/template"

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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	cortexv1alpha1 "github.com/opstrace/cortex-operator/api/v1alpha1"
)

// CortexReconciler reconciles a Cortex object
type CortexReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

type resource struct {
	obj     client.Object
	mutator controllerutil.MutateFn
}

const FinalizerName = "cortex.opstrace.io/finalizer"

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

	resources := make([]*resource, 0)

	o, err := makeCortexConfigMap(req)
	if err != nil {
		log.Error(err, "failed to generate cortex configmap, will not retry")
		return ctrl.Result{Requeue: false}, err
	}
	resources = append(resources, o)

	o = makeServiceAccount(req)
	resources = append(resources, o)

	o = makeHeadlessService(req, "memcached-results", servicePort{"memcached-client", 11211})
	resources = append(resources, o)

	o = makeStatefulSetMemcached(req, "memcached-results")
	resources = append(resources, o)

	o = makeHeadlessService(req, "memcached-metadata", servicePort{"memcached-client", 11211})
	resources = append(resources, o)

	o = makeStatefulSetMemcached(req, "memcached-metadata")
	resources = append(resources, o)

	o = makeHeadlessService(req, "gossip-ring", servicePort{"gossip-ring", 7946})
	resources = append(resources, o)

	o = makeDeploymentDistributor(req, cortex)
	resources = append(resources, o)

	o = makeService(req, "distributor", servicePort{"http", 80}, servicePort{"distributor-grpc", 9095})
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

func generateCortexConfig(req ctrl.Request) (string, error) {
	// Anonymous struct with the fields required by the cortex config template
	data := struct {
		Namespace string
	}{
		req.Namespace,
	}

	t := template.New("cortexConfig")

	t, err := t.Parse(CortexConfigTemplate)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	err = t.Execute(&b, data)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func makeCortexConfigMap(req ctrl.Request) (*resource, error) {
	configStr, err := generateCortexConfig(req)
	if err != nil {
		return nil, err
	}

	configMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "cortex", Namespace: req.Namespace}}

	return &resource{
		obj: configMap,
		mutator: func() error {
			configMap.Data = map[string]string{
				"config.yaml": string(configStr),
			}
			return nil
		},
	}, nil
}

func makeServiceAccount(req ctrl.Request) *resource {
	serviceAccount := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "cortex", Namespace: req.Namespace}}

	return &resource{
		obj: serviceAccount,
		mutator: func() error {
			return nil
		},
	}
}

func makeStatefulSetMemcached(req ctrl.Request, name string) *resource {
	statefulSet := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: req.Namespace}}

	return &resource{
		obj: statefulSet,
		mutator: func() error {
			statefulSet.Spec.ServiceName = name
			statefulSet.Spec.Replicas = pointer.Int32Ptr(1)
			statefulSet.Spec.PodManagementPolicy = "Parallel"
			statefulSet.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: map[string]string{"name": name},
			}
			statefulSet.Spec.Template.ObjectMeta.Labels = map[string]string{
				"name": name,
			}
			statefulSet.Spec.Template.Spec.Affinity = WithPodAntiAffinity(name)
			statefulSet.Spec.Template.Spec.Containers = []corev1.Container{
				{
					Name:  "memcached",
					Image: "memcached:1.6.9-alpine",
					Args: []string{
						"-m 4096",
						"-I 2m",
						"-c 1024",
						"-v",
					},
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

func makeDeploymentDistributor(req ctrl.Request, cortex *cortexv1alpha1.Cortex) *resource {
	deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "distributor", Namespace: req.Namespace}}
	labels := map[string]string{
		"name":       "distributor",
		"memberlist": "gossip-ring",
	}

	return &resource{
		obj: deployment,
		mutator: func() error {
			deployment.Spec.Replicas = pointer.Int32Ptr(1)
			deployment.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: labels,
			}
			deployment.Spec.Template.Labels = labels
			deployment.Spec.Template.Spec.Affinity = WithPodAntiAffinity("distributor")
			deployment.Spec.Template.Spec.Containers = []corev1.Container{
				{
					Name:            "distributor",
					Image:           cortex.Spec.Image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Args: []string{
						"-target=distributor",
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
			deployment.Spec.Template.Spec.Volumes = []corev1.Volume{
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

type servicePort struct {
	Name string
	Port int32
}

func makeHeadlessService(req ctrl.Request, name string, servicePorts ...servicePort) *resource {
	service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: req.Namespace}}

	return &resource{
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
					Port:       p.Port,
					TargetPort: intstr.IntOrString{IntVal: p.Port},
				})
			}
			service.Spec.Selector = map[string]string{"name": name}

			return nil
		},
	}
}

func makeService(req ctrl.Request, name string, servicePorts ...servicePort) *resource {
	service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: req.Namespace}}

	return &resource{
		obj: service,
		mutator: func() error {
			service.Labels = map[string]string{
				"name": name,
				"job":  fmt.Sprintf("%s.%s", req.Namespace, name),
			}
			service.Spec.Ports = make([]corev1.ServicePort, 0)
			for _, p := range servicePorts {
				service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{
					Name:       p.Name,
					Port:       p.Port,
					TargetPort: intstr.IntOrString{IntVal: p.Port},
				})
			}
			service.Spec.Selector = map[string]string{"name": name}

			return nil
		},
	}
}

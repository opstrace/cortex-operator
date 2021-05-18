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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
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

	cortexConfigStr, cortexConfigSHA, err := generateCortexConfig(cortex)
	if err != nil {
		log.Error(err, "failed to generate cortex configmap, will not retry")
		return ctrl.Result{Requeue: false}, nil
	}

	o := makeCortexConfigMap(req, cortexConfigStr)
	resources = append(resources, o)

	o = makeServiceAccount(req)
	resources = append(resources, o)

	o = makeHeadlessService(req, "memcached", "name", servicePort{"memcached-client", 11211})
	resources = append(resources, o)

	o = makeStatefulSetMemcached(req, "memcached")
	resources = append(resources, o)

	o = makeHeadlessService(req, "memcached-index-queries", "name", servicePort{"memcached-client", 11211})
	resources = append(resources, o)

	o = makeStatefulSetMemcached(req, "memcached-index-queries")
	resources = append(resources, o)

	o = makeHeadlessService(req, "memcached-index-writes", "name", servicePort{"memcached-client", 11211})
	resources = append(resources, o)

	o = makeStatefulSetMemcached(req, "memcached-index-writes")
	resources = append(resources, o)

	o = makeHeadlessService(req, "memcached-results", "name", servicePort{"memcached-client", 11211})
	resources = append(resources, o)

	o = makeStatefulSetMemcached(req, "memcached-results")
	resources = append(resources, o)

	o = makeHeadlessService(req, "memcached-metadata", "name", servicePort{"memcached-client", 11211})
	resources = append(resources, o)

	o = makeStatefulSetMemcached(req, "memcached-metadata")
	resources = append(resources, o)

	o = makeHeadlessService(req, "gossip-ring", "memberlist", servicePort{"gossip-ring", 7946})
	resources = append(resources, o)

	o = makeDeployment(req, cortex, "distributor", cortexConfigSHA)
	resources = append(resources, o)

	o = makeService(req, "distributor", servicePort{"http", 80}, servicePort{"distributor-grpc", 9095})
	resources = append(resources, o)

	o = makeDeployment(req, cortex, "querier", cortexConfigSHA)
	resources = append(resources, o)

	o = makeService(req, "querier", servicePort{"http", 80}, servicePort{"querier-grpc", 9095})
	resources = append(resources, o)

	o = makeDeployment(req, cortex, "query-frontend", cortexConfigSHA)
	resources = append(resources, o)

	o = makeService(req, "query-frontend", servicePort{"http", 80}, servicePort{"querier-grpc", 9095})
	resources = append(resources, o)

	o = makeStatefulSetIngester(req, cortex, cortexConfigSHA)
	resources = append(resources, o)

	o = makeService(req, "ingester", servicePort{"http", 80}, servicePort{"ingester-grpc", 9095})
	resources = append(resources, o)

	o = makeStatefulSet(req, cortex, "store-gateway", cortexConfigSHA)
	resources = append(resources, o)

	o = makeService(req, "store-gateway", servicePort{"http", 80}, servicePort{"store-gateway-grpc", 9095})
	resources = append(resources, o)

	o = makeStatefulSet(req, cortex, "compactor", cortexConfigSHA)
	resources = append(resources, o)

	o = makeService(req, "compactor", servicePort{"http", 80}, servicePort{"compactor-grpc", 9095})
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
	t := template.New("cortexConfig")

	t, err := t.Parse(CortexConfigTemplate)
	if err != nil {
		return "", "", err
	}

	var b bytes.Buffer
	err = t.Execute(&b, cortex)
	if err != nil {
		return "", "", err
	}

	sha := sha256.Sum256(b.Bytes())
	s := hex.EncodeToString(sha[:])

	return b.String(), s, nil
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

func makeStatefulSetMemcached(req ctrl.Request, name string) *kubernetesResource {
	statefulSet := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: req.Namespace}}

	return &kubernetesResource{
		obj: statefulSet,
		mutator: func() error {
			statefulSet.Spec.ServiceName = name
			statefulSet.Spec.Replicas = pointer.Int32Ptr(1)
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

func makeDeployment(
	req ctrl.Request,
	cortex *cortexv1alpha1.Cortex,
	name string,
	cortexConfigSHA string,
) *kubernetesResource {
	deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: req.Namespace}}
	labels := map[string]string{
		"name":       name,
		"memberlist": "gossip-ring",
	}
	annotations := map[string]string{
		CortexConfigShasumAnnotationName: cortexConfigSHA,
	}

	return &kubernetesResource{
		obj: deployment,
		mutator: func() error {
			deployment.Spec.Replicas = pointer.Int32Ptr(2)
			deployment.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: labels,
			}
			deployment.Spec.Template.Labels = labels
			deployment.Spec.Template.Annotations = annotations
			deployment.Spec.Template.Spec.Affinity = WithPodAntiAffinity(name)
			deployment.Spec.Template.Spec.Containers = []corev1.Container{
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

func makeService(req ctrl.Request, name string, servicePorts ...servicePort) *kubernetesResource {
	service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: req.Namespace}}

	return &kubernetesResource{
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
					Port:       int32(p.Port),
					TargetPort: intstr.FromInt(p.Port),
				})
			}
			service.Spec.Selector = map[string]string{"name": name}

			return nil
		},
	}
}

func makeStatefulSetIngester(
	req ctrl.Request,
	cortex *cortexv1alpha1.Cortex,
	cortexConfigSHA string,
) *kubernetesResource {
	statefulSet := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "ingester", Namespace: req.Namespace}}
	labels := map[string]string{
		"name": "ingester",
	}
	annotations := map[string]string{
		CortexConfigShasumAnnotationName: cortexConfigSHA,
	}

	return &kubernetesResource{
		obj: statefulSet,
		mutator: func() error {
			statefulSet.Spec.ServiceName = "ingester"
			statefulSet.Spec.Replicas = pointer.Int32Ptr(2)
			statefulSet.Spec.PodManagementPolicy = appsv1.OrderedReadyPodManagement
			statefulSet.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: map[string]string{"name": "ingester"},
			}
			statefulSet.Spec.Template.Labels = labels
			statefulSet.Spec.Template.Annotations = annotations
			statefulSet.Spec.Template.Spec.Affinity = WithPodAntiAffinity("ingester")
			statefulSet.Spec.Template.Spec.Containers = []corev1.Container{
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
			statefulSet.Spec.Template.Spec.ServiceAccountName = ServiceAccountName
			// https://cortexmetrics.io/docs/guides/running-cortex-on-kubernetes/#take-extra-care-with-ingesters
			statefulSet.Spec.Template.Spec.TerminationGracePeriodSeconds = pointer.Int64Ptr(2400)
			statefulSet.Spec.Template.Spec.Volumes = []corev1.Volume{
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
				{
					Name: "datadir",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "datadir",
						},
					},
				},
			}
			statefulSet.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "datadir"},
					Spec: corev1.PersistentVolumeClaimSpec{
						// Uses the default storage class.
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"storage": resource.MustParse("1Gi"),
							},
						},
					},
				},
			}

			return nil
		},
	}
}

func makeStatefulSet(
	req ctrl.Request,
	cortex *cortexv1alpha1.Cortex,
	name string,
	cortexConfigSHA string,
) *kubernetesResource {
	statefulSet := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: req.Namespace}}
	labels := map[string]string{
		"name": name,
	}
	annotations := map[string]string{
		CortexConfigShasumAnnotationName: cortexConfigSHA,
	}
	return &kubernetesResource{
		obj: statefulSet,
		mutator: func() error {
			statefulSet.Spec.ServiceName = name
			statefulSet.Spec.Replicas = pointer.Int32Ptr(2)
			statefulSet.Spec.PodManagementPolicy = appsv1.OrderedReadyPodManagement
			statefulSet.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: map[string]string{"name": name},
			}
			statefulSet.Spec.Template.Labels = labels
			statefulSet.Spec.Template.Annotations = annotations
			statefulSet.Spec.Template.Spec.Affinity = WithPodAntiAffinity(name)
			statefulSet.Spec.Template.Spec.Containers = []corev1.Container{
				{
					Name:  name,
					Image: cortex.Spec.Image,
					Args: []string{
						"-target=" + name,
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
			statefulSet.Spec.Template.Spec.ServiceAccountName = ServiceAccountName
			statefulSet.Spec.Template.Spec.Volumes = []corev1.Volume{
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
				{
					Name: "datadir",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "datadir",
						},
					},
				},
			}
			statefulSet.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "datadir"},
					Spec: corev1.PersistentVolumeClaimSpec{
						// Uses the default storage class.
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"storage": resource.MustParse("1Gi"),
							},
						},
					},
				},
			}

			return nil
		},
	}
}

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	cortexv1alpha1 "github.com/opstrace/cortex-operator/api/v1alpha1"
)

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
	// automatically deleted when he owner (cortex) is deleted.
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

func NewService(req ctrl.Request, name string) *KubernetesResource {
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: req.Namespace}}
	ref := &corev1.LocalObjectReference{Name: name}
	return &KubernetesResource{
		obj: svc,
		ref: ref,
		mutator: func() error {
			svc.Labels = map[string]string{
				"name": name,
				"job":  fmt.Sprintf("%s.%s", req.Namespace, name),
			}
			svc.Spec.Ports = make([]corev1.ServicePort, 0)
			svc.Spec.Ports = []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
				{
					Name:       "grpc",
					Port:       9095,
					TargetPort: intstr.FromInt(9095),
				},
			}
			svc.Spec.Selector = map[string]string{"name": name}

			return nil
		},
	}
}

func NewDeployment(
	req ctrl.Request,
	name string,
	cortex *cortexv1alpha1.Cortex,
	spec *cortexv1alpha1.DeploymentSpec,
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
			deploy.Spec.Replicas = spec.Replicas
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
					Name: CortexRuntimeConfigMapName,
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: CortexRuntimeConfigMapName,
							},
						},
					},
				},
			}

			return nil
		},
	}
}

func NewStatefulSet(
	req ctrl.Request,
	name string,
	cortex *cortexv1alpha1.Cortex,
	spec *cortexv1alpha1.StatefulSetSpec,
) *KubernetesResource {
	sts := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: req.Namespace}}
	labels := map[string]string{
		"name": name,
	}
	annotations := map[string]string{
		CortexConfigShasumAnnotationName: cortex.Spec.ConfigSHA(),
	}
	ref := &corev1.LocalObjectReference{Name: name}

	return &KubernetesResource{
		obj: sts,
		ref: ref,
		mutator: func() error {
			sts.Spec.ServiceName = name
			sts.Spec.Replicas = spec.Replicas
			sts.Spec.PodManagementPolicy = appsv1.OrderedReadyPodManagement
			sts.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: map[string]string{"name": name},
			}
			sts.Spec.Template.Labels = labels
			sts.Spec.Template.Annotations = annotations
			sts.Spec.Template.Spec.Affinity = WithPodAntiAffinity(name)
			sts.Spec.Template.Spec.Containers = []corev1.Container{
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
			sts.Spec.Template.Spec.ServiceAccountName = ServiceAccountName
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
						StorageClassName: spec.StorageClassName,
						AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"storage": *spec.DatadirSize,
							},
						},
					},
				},
			}

			return nil
		},
	}
}

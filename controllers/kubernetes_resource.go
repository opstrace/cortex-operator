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
	"k8s.io/apimachinery/pkg/runtime"
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

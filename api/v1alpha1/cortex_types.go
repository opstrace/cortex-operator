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

package v1alpha1

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CortexSpec defines the desired state of Cortex
type CortexSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Image of Cortex to deploy.
	Image string `json:"image,omitempty"`

	ServiceAccountSpec *ServiceAccountSpec `json:"service_account_spec,omitempty"`

	IngesterSpec      *StatefulSetSpec `json:"ingester_spec,omitempty"`
	CompactorSpec     *StatefulSetSpec `json:"compactor_spec,omitempty"`
	StoreGatewaySpec  *StatefulSetSpec `json:"store_gateway_spec,omitempty"`
	DistributorSpec   *DeploymentSpec  `json:"distributor_spec,omitempty"`
	QuerierSpec       *DeploymentSpec  `json:"querier_spec,omitempty"`
	QueryFrontendSpec *DeploymentSpec  `json:"query_frontend_spec,omitempty"`
	AlertManagerSpec  *DeploymentSpec  `json:"alertmanager_spec,omitempty"`
	RulerSpec         *DeploymentSpec  `json:"ruler_spec,omitempty"`

	Memcached *MemcachedSpec `json:"memcached,omitempty"`

	// Config accepts any object, meaning it accepts any valid Cortex config
	// yaml. Defaulting and Validation are done in the webhooks.
	// +kubebuilder:pruning:PreserveUnknownFields
	Config runtime.RawExtension `json:"config,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields
	RuntimeConfig *runtime.RawExtension `json:"runtime_config,omitempty"`
}

func (c *CortexSpec) ConfigSHA() string {
	sha := sha256.Sum256(c.Config.Raw)
	return hex.EncodeToString(sha[:])
}

// CortexStatus defines the observed state of Cortex
type CortexStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	MemcachedRef     *MemcachedReference          `json:"memcached,omitempty"`
	IngesterRef      *corev1.LocalObjectReference `json:"ingester,omitempty"`
	DistributorRef   *corev1.LocalObjectReference `json:"distributor,omitempty"`
	QuerierRef       *corev1.LocalObjectReference `json:"querier,omitempty"`
	QueryFrontendRef *corev1.LocalObjectReference `json:"query_frontend,omitempty"`
	CompactorRef     *corev1.LocalObjectReference `json:"compactor,omitempty"`
	StoreGatewayRef  *corev1.LocalObjectReference `json:"store_gateway,omitempty"`
	AlertManagerRef  *corev1.LocalObjectReference `json:"alertmanager,omitempty"`
	RulerRef         *corev1.LocalObjectReference `json:"ruler,omitempty"`
}

// MemcachedReference holds references to all the Memcached StatefulSets
type MemcachedReference struct {
	ChunksCacheRef       *corev1.LocalObjectReference `json:"chunks_cache,omitempty"`
	IndexQueriesCacheRef *corev1.LocalObjectReference `json:"index_queries_cache,omitempty"`
	IndexWritesCacheRef  *corev1.LocalObjectReference `json:"index_writes_cache,omitempty"`
	ResultsCacheRef      *corev1.LocalObjectReference `json:"results_cache,omitempty"`
	MetadataCacheRef     *corev1.LocalObjectReference `json:"metadata_cache,omitempty"`
}

// IsSet returns true if all the resource references are not nil
func (r *MemcachedReference) IsSet() bool {
	return r != nil &&
		r.ChunksCacheRef != nil &&
		r.IndexQueriesCacheRef != nil &&
		r.IndexWritesCacheRef != nil &&
		r.ResultsCacheRef != nil &&
		r.MetadataCacheRef != nil
}

type StatefulSetSpec struct {
	DatadirSize      *resource.Quantity `json:"datadir_size,omitempty"`
	StorageClassName *string            `json:"storage_class_name,omitempty"`
	Replicas         *int32             `json:"replicas,omitempty"`
	Env              []corev1.EnvVar    `json:"env,omitempty"`
}

func (s *StatefulSetSpec) Default() {
	if s.DatadirSize == nil {
		r := resource.MustParse("1Gi")
		s.DatadirSize = &r
	}

	if s.Replicas == nil {
		s.Replicas = pointer.Int32Ptr(2)
	}

	if s.Env == nil {
		s.Env = make([]corev1.EnvVar, 0)
	}
}

type DeploymentSpec struct {
	Replicas *int32          `json:"replicas,omitempty"`
	Env      []corev1.EnvVar `json:"env,omitempty"`
}

func (s *DeploymentSpec) Default() {
	if s.Replicas == nil {
		s.Replicas = pointer.Int32Ptr(2)
	}

	if s.Env == nil {
		s.Env = make([]corev1.EnvVar, 0)
	}
}

type MemcachedSpec struct {
	Image string `json:"image,omitempty"`

	ChunksCacheSpec       *MemcachedStatefulSetSpec `json:"chunks_cache_spec,omitempty"`
	IndexQueriesCacheSpec *MemcachedStatefulSetSpec `json:"index_queries_cache_spec,omitempty"`
	IndexWritesCacheSpec  *MemcachedStatefulSetSpec `json:"index_writes_cache_spec,omitempty"`
	ResultsCacheSpec      *MemcachedStatefulSetSpec `json:"results_cache_spec,omitempty"`
	MetadataCacheSpec     *MemcachedStatefulSetSpec `json:"metadata_cache_spec,omitempty"`
}

func (m *MemcachedSpec) Default() {
	if m.Image == "" {
		m.Image = "memcached:1.6.9-alpine"
	}

	if m.ChunksCacheSpec == nil {
		m.ChunksCacheSpec = &MemcachedStatefulSetSpec{}
	}
	if m.IndexQueriesCacheSpec == nil {
		m.IndexQueriesCacheSpec = &MemcachedStatefulSetSpec{}
	}
	if m.IndexWritesCacheSpec == nil {
		m.IndexWritesCacheSpec = &MemcachedStatefulSetSpec{}
	}
	if m.ResultsCacheSpec == nil {
		m.ResultsCacheSpec = &MemcachedStatefulSetSpec{}
	}
	if m.MetadataCacheSpec == nil {
		m.MetadataCacheSpec = &MemcachedStatefulSetSpec{}
	}

	m.ChunksCacheSpec.Default()
	m.IndexQueriesCacheSpec.Default()
	m.IndexWritesCacheSpec.Default()
	m.ResultsCacheSpec.Default()
	m.MetadataCacheSpec.Default()
}

type MemcachedStatefulSetSpec struct {
	Replicas *int32 `json:"replicas,omitempty"`
	// MemoryLimit is the item memory in megabytes
	MemoryLimit *int32 `json:"memory_limit,omitempty"`
	// MaxItemSize adjusts max item size
	MaxItemSize *string `json:"max_item_size,omitempty"`
}

func (m *MemcachedStatefulSetSpec) AsArgs() []string {
	return []string{
		fmt.Sprintf("-m %d", *m.MemoryLimit),
		fmt.Sprintf("-I %s", *m.MaxItemSize),
	}
}

func (m *MemcachedStatefulSetSpec) Default() {
	if m == nil {
		m = &MemcachedStatefulSetSpec{}
	}
	if m.Replicas == nil {
		m.Replicas = pointer.Int32Ptr(2)
	}
	if m.MemoryLimit == nil {
		m.MemoryLimit = pointer.Int32Ptr(4096)
	}
	if m.MaxItemSize == nil {
		m.MaxItemSize = pointer.StringPtr("2m")
	}
}

type ServiceAccountSpec struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	// ImagePullSecrets is a list of references to secrets in the same namespace
	// to use for pulling any images in pods that reference this ServiceAccount.
	// More info:
	// https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod
	ImagePullSecrets []corev1.LocalObjectReference `json:"image_pull_secrets,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Cortex is the Schema for the cortices API
type Cortex struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CortexSpec   `json:"spec,omitempty"`
	Status CortexStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CortexList contains a list of Cortex
type CortexList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cortex `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cortex{}, &CortexList{})
}

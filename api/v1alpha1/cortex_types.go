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

	RuntimeConfig *RuntimeConfigSpec `json:"runtime_config,omitempty"`
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
	//+kubebuilder:default="1Gi"
	DatadirSize      *resource.Quantity `json:"datadir_size,omitempty"`
	StorageClassName *string            `json:"storage_class_name,omitempty"`
	//+kubebuilder:default=2
	Replicas *int32 `json:"replicas,omitempty"`
}

type DeploymentSpec struct {
	//+kubebuilder:default=2
	Replicas *int32 `json:"replicas,omitempty"`
}

type MemcachedSpec struct {
	//+kubebuilder:default="memcached:1.6.9-alpine"
	Image string `json:"image,omitempty"`

	ChunksCacheSpec       *MemcachedStatefulSetSpec `json:"chunks_cache_spec,omitempty"`
	IndexQueriesCacheSpec *MemcachedStatefulSetSpec `json:"index_queries_cache_spec,omitempty"`
	IndexWritesCacheSpec  *MemcachedStatefulSetSpec `json:"index_writes_cache_spec,omitempty"`
	ResultsCacheSpec      *MemcachedStatefulSetSpec `json:"results_cache_spec,omitempty"`
	MetadataCacheSpec     *MemcachedStatefulSetSpec `json:"metadata_cache_spec,omitempty"`
}

type MemcachedStatefulSetSpec struct {
	//+kubebuilder:default=2
	Replicas *int32 `json:"replicas,omitempty"`
	// MemoryLimit is the item memory in megabytes
	//+kubebuilder:default=4096
	MemoryLimit *int32 `json:"memory_limit,omitempty"`
	// MaxItemSize adjusts max item size
	//+kubebuilder:default="2m"
	MaxItemSize *string `json:"max_item_size,omitempty"`
}

func (m *MemcachedStatefulSetSpec) AsArgs() []string {
	return []string{
		fmt.Sprintf("-m %d", *m.MemoryLimit),
		fmt.Sprintf("-I %s", *m.MaxItemSize),
	}
}

type RuntimeConfigSpec struct {
	// +kubebuilder:pruning:PreserveUnknownFields
	Overrides runtime.RawExtension `json:"overrides,omitempty"`
}

type ServiceAccountSpec struct {
	Annotations map[string]string `json:"annotations,omitempty"`
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

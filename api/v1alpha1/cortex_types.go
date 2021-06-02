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

	IngesterSpec      *StatefulSetSpec `json:"ingester_spec,omitempty"`
	CompactorSpec     *StatefulSetSpec `json:"compactor_spec,omitempty"`
	StoreGatewaySpec  *StatefulSetSpec `json:"store_gateway_spec,omitempty"`
	DistributorSpec   *DeploymentSpec  `json:"distributor_spec,omitempty"`
	QuerierSpec       *DeploymentSpec  `json:"querier_spec,omitempty"`
	QueryFrontendSpec *DeploymentSpec  `json:"query_frontend_spec,omitempty"`

	// Config accepts any object, meaning it accepts any valid Cortex config
	// yaml. Defaulting and Validation are done in the webhooks.
	// +kubebuilder:pruning:PreserveUnknownFields
	Config runtime.RawExtension `json:"config,omitempty"`
}

func (c *CortexSpec) ConfigSHA() string {
	sha := sha256.Sum256(c.Config.Raw)
	return hex.EncodeToString(sha[:])
}

// CortexStatus defines the observed state of Cortex
type CortexStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	MemcachedRef     *MemcachedReference     `json:"memcached,omitempty"`
	IngesterRef      *IngesterReference      `json:"ingester,omitempty"`
	DistributorRef   *DistributorReference   `json:"distributor,omitempty"`
	QuerierRef       *QuerierReference       `json:"querier,omitempty"`
	QueryFrontendRef *QueryFrontendReference `json:"query_frontend,omitempty"`
	CompactorRef     *CompactorReference     `json:"compactor,omitempty"`
	StoreGatewayRef  *StoreGatewayReference  `json:"store_gateway,omitempty"`
}

// MemcachedReference holds references to all the Memcached resources
type MemcachedReference struct {
	MemcachedSvc             *corev1.LocalObjectReference `json:"memcached_svc,omitempty"`
	MemcachedSts             *corev1.LocalObjectReference `json:"memcached_sts,omitempty"`
	MemcachedIndexQueriesSvc *corev1.LocalObjectReference `json:"memcached_index_queries_svc,omitempty"`
	MemcachedIndexQueriesSts *corev1.LocalObjectReference `json:"memcached_index_queries_sts,omitempty"`
	MemcachedIndexWritesSvc  *corev1.LocalObjectReference `json:"memcached_index_writes_svc,omitempty"`
	MemcachedIndexWritesSts  *corev1.LocalObjectReference `json:"memcached_index_writes_sts,omitempty"`
	MemcachedResultsSvc      *corev1.LocalObjectReference `json:"memcached_results_svc,omitempty"`
	MemcachedResultsSts      *corev1.LocalObjectReference `json:"memcached_results_sts,omitempty"`
	MemcachedMetadataSvc     *corev1.LocalObjectReference `json:"memcached_metadata_svc,omitempty"`
	MemcachedMetadataSts     *corev1.LocalObjectReference `json:"memcached_metadata_sts,omitempty"`
}

// IsSet returns true if all the resource references are not nil
func (r *MemcachedReference) IsSet() bool {
	return r != nil &&
		r.MemcachedSvc != nil && r.MemcachedSts != nil &&
		r.MemcachedIndexQueriesSvc != nil && r.MemcachedIndexQueriesSts != nil &&
		r.MemcachedIndexWritesSvc != nil && r.MemcachedIndexWritesSts != nil &&
		r.MemcachedResultsSvc != nil && r.MemcachedResultsSts != nil &&
		r.MemcachedMetadataSvc != nil && r.MemcachedMetadataSts != nil
}

// IngesterReference holds references to the Service and Statefulset required to
// run the Ingesters
type IngesterReference struct {
	Svc *corev1.LocalObjectReference `json:"ingester_svc,omitempty"`
	Sts *corev1.LocalObjectReference `json:"ingester_sts,omitempty"`
}

// DistributorReference holds references to the Service and Deployment required to
// run the Distributors
type DistributorReference struct {
	Svc    *corev1.LocalObjectReference `json:"distributor_svc,omitempty"`
	Deploy *corev1.LocalObjectReference `json:"distributor_deploy,omitempty"`
}

// QuerierReference holds references to the Service and Deployment required to
// run the Queriers
type QuerierReference struct {
	Svc    *corev1.LocalObjectReference `json:"querier_svc,omitempty"`
	Deploy *corev1.LocalObjectReference `json:"querier_deploy,omitempty"`
}

// QueryFrontendReference holds references to the Service and Deployment required to
// run the Query Frontends
type QueryFrontendReference struct {
	Svc    *corev1.LocalObjectReference `json:"query_frontend_svc,omitempty"`
	Deploy *corev1.LocalObjectReference `json:"query_frontend_deploy,omitempty"`
}

// CompactorReference holds references to the Service and Statefulset required to
// run the Compactor
type CompactorReference struct {
	Svc *corev1.LocalObjectReference `json:"compactor_svc,omitempty"`
	Sts *corev1.LocalObjectReference `json:"compactor_deploy,omitempty"`
}

// StoreGatewayReference holds references to the Service and Statefulset required to
// run the Store Gateway
type StoreGatewayReference struct {
	Svc *corev1.LocalObjectReference `json:"store_gateway_svc,omitempty"`
	Sts *corev1.LocalObjectReference `json:"store_gateway_deploy,omitempty"`
}

type StatefulSetSpec struct {
	//+kubebuilder:default="1Gi"
	DatadirSize *resource.Quantity `json:"datadir_size,omitempty"`
	//+kubebuilder:default=2
	Replicas *int32 `json:"replicas,omitempty"`
}

type DeploymentSpec struct {
	//+kubebuilder:default=2
	Replicas *int32 `json:"replicas,omitempty"`
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

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
	"bytes"
	"fmt"
	"html/template"

	"github.com/miracl/conflate"
	yamlv2 "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/yaml"

	"github.com/cortexproject/cortex/pkg/cortex"
	"github.com/cortexproject/cortex/pkg/util/flagext"
)

// log is for logging in this package.
var cortexlog = logf.Log.WithName("cortex-resource")

func (r *Cortex) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:webhookVersions={v1beta1},path=/mutate-cortex-opstrace-io-v1alpha1-cortex,mutating=true,failurePolicy=fail,sideEffects=None,groups=cortex.opstrace.io,resources=cortices,verbs=create;update,versions=v1alpha1,name=mcortex.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &Cortex{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Cortex) Default() {
	cortexlog.Info("default", "name", r.Name)

	if r.Spec.Memcached == nil {
		r.Spec.Memcached = &MemcachedSpec{}
	}
	r.Spec.Memcached.Default()

	if r.Spec.IngesterSpec == nil {
		r.Spec.IngesterSpec = &StatefulSetSpec{}
	}
	r.Spec.IngesterSpec.Default()

	if r.Spec.CompactorSpec == nil {
		r.Spec.CompactorSpec = &StatefulSetSpec{}
	}
	r.Spec.CompactorSpec.Default()

	if r.Spec.StoreGatewaySpec == nil {
		r.Spec.StoreGatewaySpec = &StatefulSetSpec{}
	}
	r.Spec.StoreGatewaySpec.Default()

	if r.Spec.DistributorSpec == nil {
		r.Spec.DistributorSpec = &DeploymentSpec{}
	}
	r.Spec.DistributorSpec.Default()

	if r.Spec.QuerierSpec == nil {
		r.Spec.QuerierSpec = &DeploymentSpec{}
	}
	r.Spec.QuerierSpec.Default()

	if r.Spec.QueryFrontendSpec == nil {
		r.Spec.QueryFrontendSpec = &DeploymentSpec{}
	}
	r.Spec.QueryFrontendSpec.Default()

	if r.Spec.AlertManagerSpec == nil {
		r.Spec.AlertManagerSpec = &DeploymentSpec{}
	}
	r.Spec.AlertManagerSpec.Default()

	if r.Spec.RulerSpec == nil {
		r.Spec.RulerSpec = &DeploymentSpec{}
	}
	r.Spec.RulerSpec.Default()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:webhookVersions={v1beta1},path=/validate-cortex-opstrace-io-v1alpha1-cortex,mutating=false,failurePolicy=fail,sideEffects=None,groups=cortex.opstrace.io,resources=cortices,verbs=create;update,versions=v1alpha1,name=vcortex.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &Cortex{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Cortex) ValidateCreate() error {
	cortexlog.Info("validate create", "name", r.Name)
	return r.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Cortex) ValidateUpdate(old runtime.Object) error {
	cortexlog.Info("validate update", "name", r.Name)
	return r.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Cortex) ValidateDelete() error {
	cortexlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *Cortex) RuntimeConfigAsYAML() ([]byte, error) {
	if r == nil {
		return []byte{}, nil
	}

	return yaml.Marshal(r.Spec.RuntimeConfig)
}

// generateCortexConfig returns a config yaml with the cortex-operator default
// configuration.
func (r *Cortex) generateCortexConfig() ([]byte, error) {
	t := template.New("cortex-config")

	t, err := t.Parse(DefaultCortexConfigTemplate)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	err = t.Execute(&b, r)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (r *Cortex) AsJSON() ([]byte, error) {
	// Generate the default configuration.
	defaultCortexConfig, err := r.generateCortexConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to generate default configuration: %w", err)
	}
	// Convert it to JSON to be able to merge with the user config.
	y, err := yaml.YAMLToJSON(defaultCortexConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to convert default configuration: %w", err)
	}
	// Merge the user config and the defaults. The defaults will override fields
	// set by the user.
	c, err := conflate.FromData(r.Spec.Config.Raw, y)
	if err != nil {
		return nil, fmt.Errorf("failed to merge default configuration with user settings: %w", err)
	}
	// Convert the data to json.
	j, err := c.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to convert user configuration: %w", err)
	}
	return j, nil
}

func (r *Cortex) AsYAML() ([]byte, error) {
	j, err := r.AsJSON()
	if err != nil {
		return nil, err
	}
	return yaml.JSONToYAML(j)
}

// AsCortexConfig converts the configuration to an upstream cortex.Config
// object.
func (r *Cortex) AsCortexConfig() (*cortex.Config, error) {
	j, err := r.AsJSON()
	if err != nil {
		return nil, err
	}
	// Unmarshal the json to a cortex.Config structure to validate it.
	cfg := &cortex.Config{}
	// Set up the cortex config defaults otherwise validation will fail because
	// fields are not set.
	flagext.DefaultValues(cfg)
	// Unmarshal the desired cortex config into the object overriding the
	// defaults.
	err = yamlv2.UnmarshalStrict(j, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to : %w", err)
	}
	return cfg, nil
}

func (r *Cortex) Validate() error {
	cfg, err := r.AsCortexConfig()
	if err != nil {
		return err
	}
	// Validate the cortex configuration.
	return cfg.Validate(nil)
}

const DefaultCortexConfigTemplate = `
http_prefix: ''
api:
  alertmanager_http_prefix: /alertmanager
  response_compression_enabled: true
auth_enabled: true
distributor:
  shard_by_all_labels: true
  pool:
    health_check_ingesters: true
  ha_tracker:
    enable_ha_tracker: false
memberlist:
  abort_if_cluster_join_fails: true
  bind_port: 7946
  join_members:
  - 'gossip-ring.{{.Namespace}}.svc.cluster.local:7946'
querier:
  batch_iterators: true
  ingester_streaming: true
  store_gateway_addresses: 'store-gateway.{{.Namespace}}.svc.cluster.local:9095'
query_range:
  align_queries_with_step: true
  cache_results: true
  results_cache:
    cache:
      memcached_client:
        consistent_hash: true
        host: memcached-results.{{.Namespace}}.svc.cluster.local
        service: memcached-client
frontend_worker:
  frontend_address: 'query-frontend.{{.Namespace}}.svc.cluster.local:9095'
ingester:
  lifecycler:
    ring:
      kvstore:
        store: memberlist
blocks_storage:
  tsdb:
    dir: /cortex/tsdb
    wal_compression_enabled: true
  bucket_store:
    sync_dir: /cortex/tsdb-sync
    index_cache:
      backend: memcached
      memcached:
        addresses: 'dnssrv+memcached-index-queries.{{.Namespace}}.svc.cluster.local:11211'
    chunks_cache:
      backend: memcached
      memcached:
        addresses: 'dnssrv+memcached-chunks.{{.Namespace}}.svc.cluster.local:11211'
    metadata_cache:
      backend: memcached
      memcached:
        addresses: 'dnssrv+memcached-metadata.{{.Namespace}}.svc.cluster.local:11211'
store_gateway:
  sharding_enabled: true
  sharding_ring:
    kvstore:
      store: memberlist
compactor:
  data_dir: /cortex/compactor
  sharding_enabled: true
  sharding_ring:
    kvstore:
      store: memberlist
purger:
  enable: true
storage:
  engine: blocks
alertmanager:
  enable_api: true
  cluster:
    peers: 'alertmanager.{{.Namespace}}.svc.cluster.local:9094'
  sharding_enabled: true
  sharding_ring:
    kvstore:
      store: memberlist
  external_url: /alertmanager
ruler:
  enable_api: true
  enable_sharding: true
  sharding_strategy: shuffle-sharding
  ring:
    kvstore:
      store: memberlist
  alertmanager_url: 'http://alertmanager.{{.Namespace}}.svc.cluster.local/alertmanager/'
runtime_config:
  file: /etc/cortex/runtime-config.yaml
  period: 5s
`

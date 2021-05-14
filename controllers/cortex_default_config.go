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

const CortexConfigTemplate = `
http_prefix: ''
api:
  alertmanager_http_prefix: /alertmanager
auth_enabled: false # true
distributor:
  shard_by_all_labels: true
  pool:
    health_check_ingesters: true
  ha_tracker:
    enable_ha_tracker: false
server:
  grpc_server_max_recv_msg_size: 41943040
  grpc_server_max_send_msg_size: 41943040
memberlist:
  abort_if_cluster_join_fails: true
  bind_port: 7946
  join_members:
  - 'gossip-ring.{{.Namespace}}.svc.cluster.local:7946'
  max_join_backoff: 1m
  max_join_retries: 20
  min_join_backoff: 1s
querier:
  batch_iterators: true
  ingester_streaming: true
  store_gateway_addresses: 'store-gateway.{{.Namespace}}.svc.cluster.local:9095'
query_range:
  split_queries_by_interval: 24h
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
limits:
  compactor_blocks_retention_period: 192h
  ingestion_rate: 100000
  ingestion_rate_strategy: local
  ingestion_burst_size: 200000
  max_global_series_per_user: 10000000
  max_series_per_user: 5000000
  accept_ha_samples: true
  ha_cluster_label: prometheus
  ha_replica_label: prometheus_replica
  ruler_tenant_shard_size: 3
ingester:
  lifecycler:
    join_after: 30s
    observe_period: 30s
    num_tokens: 512
    ring:
      kvstore:
        store: memberlist
blocks_storage:
  tsdb:
    dir: /cortex/tsdb
    wal_compression_enabled: true
    retention_period: 6h
  backend: {{.Spec.BlocksStorage.Backend}}
  s3:
     bucket_name: {{.Spec.BlocksStorage.S3.BucketName}}
     endpoint: {{.Spec.BlocksStorage.S3.Endpoint}}
  bucket_store:
    sync_dir: /cortex/tsdb-sync
    index_cache:
      backend: memcached
      memcached:
        addresses: 'dnssrv+memcached-index-queries.{{.Namespace}}.svc.cluster.local:11211'
    chunks_cache:
      backend: memcached
      memcached:
        addresses: 'dnssrv+memcached.{{.Namespace}}.svc.cluster.local:11211'
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
configs:
  database:
    uri: >-
      // TODO(sreis) configurable database uri
    migrations_dir: /migrations
alertmanager:
  enable_api: true
  cluster:
    peers: 'alertmanager.{{.Namespace}}.svc.cluster.local:9094'
  sharding_enabled: false
  sharding_ring:
    kvstore:
      store: memberlist
  external_url: /alertmanager
alertmanager_storage:
  backend: {{.Spec.AlertManagerStorage.Backend}}
  s3:
    bucket_name: {{.Spec.AlertManagerStorage.S3.BucketName}}
    endpoint: {{.Spec.AlertManagerStorage.S3.Endpoint}}
ruler:
  enable_api: true
  enable_sharding: true
  sharding_strategy: shuffle-sharding
  ring:
   kvstore:
     store: memberlist
  alertmanager_url: 'http://alertmanager.{{.Namespace}}.svc.cluster.local/alertmanager/'
ruler_storage:
  backend: {{.Spec.RulerStorage.Backend}}
  s3:
    bucket_name: {{.Spec.RulerStorage.S3.BucketName}}
    endpoint: {{.Spec.RulerStorage.S3.Endpoint}}
`

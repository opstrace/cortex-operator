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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/opstrace/cortex-operator/api/v1alpha1"
)

var _ = Describe("Cortex controller", func() {
	const (
		CortexName      = "test-cortex"
		CortexNamespace = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	var (
		testStorageConfig = v1alpha1.StorageConfig{
			Backend: "s3",
			S3: &v1alpha1.StorageConfigS3{
				BucketName: "test-bucket",
				Endpoint:   "s3.test-region.amazonaws.com",
			},
		}

		testCortex = &v1alpha1.Cortex{
			ObjectMeta: metav1.ObjectMeta{
				Name:      CortexName,
				Namespace: CortexNamespace,
			},
			Spec: v1alpha1.CortexSpec{
				Image:               "opstrace/cortex-operator:test",
				BlocksStorage:       testStorageConfig,
				RulerStorage:        testStorageConfig,
				AlertManagerStorage: testStorageConfig,
			},
		}
	)

	Context("should manage resouces necessary for a cortex deployment ", func() {

		It("should allow to create a resource", func() {
			ctx := context.Background()
			Expect(k8sClient.Create(ctx, testCortex)).Should(Succeed())
		})

		DescribeTable("should create resources necessary for a cortex deployment",
			func(lookupKey types.NamespacedName, createdObj client.Object, validator func(createdObj client.Object)) {
				ctx := context.Background()
				Eventually(func() error {
					return k8sClient.Get(ctx, lookupKey, createdObj)
				}, timeout, interval).Should(Succeed())

				validator(createdObj)
			},
			Entry(
				"creates a config map with cortex configuration",
				types.NamespacedName{Name: CortexConfigMapName, Namespace: CortexNamespace},
				&corev1.ConfigMap{},
				func(createdObj client.Object) {
					obj, ok := createdObj.(*corev1.ConfigMap)

					Expect(ok).To(BeTrue())
					Expect(obj.Data).Should(Equal(map[string]string{
						"config.yaml": "\nhttp_prefix: ''\napi:\n  alertmanager_http_prefix: /alertmanager\nauth_enabled: false # true\ndistributor:\n  shard_by_all_labels: true\n  pool:\n    health_check_ingesters: true\n  ha_tracker:\n    enable_ha_tracker: false\nserver:\n  grpc_server_max_recv_msg_size: 41943040\n  grpc_server_max_send_msg_size: 41943040\nmemberlist:\n  abort_if_cluster_join_fails: true\n  bind_port: 7946\n  join_members:\n  - 'gossip-ring.default.svc.cluster.local:7946'\n  max_join_backoff: 1m\n  max_join_retries: 20\n  min_join_backoff: 1s\nquerier:\n  batch_iterators: true\n  ingester_streaming: true\n  store_gateway_addresses: 'store-gateway.default.svc.cluster.local:9095'\nquery_range:\n  split_queries_by_interval: 24h\n  align_queries_with_step: true\n  cache_results: true\n  results_cache:\n    cache:\n      memcached_client:\n        consistent_hash: true\n        host: memcached-results.default.svc.cluster.local\n        service: memcached-client\nfrontend_worker:\n  frontend_address: 'query-frontend.default.svc.cluster.local:9095'\nlimits:\n  compactor_blocks_retention_period: 192h\n  ingestion_rate: 100000\n  ingestion_rate_strategy: local\n  ingestion_burst_size: 200000\n  max_global_series_per_user: 10000000\n  max_series_per_user: 5000000\n  accept_ha_samples: true\n  ha_cluster_label: prometheus\n  ha_replica_label: prometheus_replica\n  ruler_tenant_shard_size: 3\ningester:\n  lifecycler:\n    join_after: 30s\n    observe_period: 30s\n    num_tokens: 512\n    ring:\n      kvstore:\n        store: memberlist\nblocks_storage:\n  tsdb:\n    dir: /cortex/tsdb\n    wal_compression_enabled: true\n    retention_period: 6h\n  backend: s3\n  s3:\n     bucket_name: test-bucket\n     endpoint: s3.test-region.amazonaws.com\n  bucket_store:\n    sync_dir: /cortex/tsdb-sync\n    index_cache:\n      backend: memcached\n      memcached:\n        addresses: 'dnssrv+memcached-index-queries.default.svc.cluster.local:11211'\n    chunks_cache:\n      backend: memcached\n      memcached:\n        addresses: 'dnssrv+memcached.default.svc.cluster.local:11211'\n    metadata_cache:\n      backend: memcached\n      memcached:\n        addresses: 'dnssrv+memcached-metadata.default.svc.cluster.local:11211'\n\nstore_gateway:\n  sharding_enabled: true\n  sharding_ring:\n    kvstore:\n      store: memberlist\ncompactor:\n  data_dir: /cortex/compactor\n  sharding_enabled: true\n  sharding_ring:\n    kvstore:\n      store: memberlist\npurger:\n  enable: true\nstorage:\n  engine: blocks\nconfigs:\n  database:\n    uri: >-\n      // TODO(sreis) configurable database uri\n    migrations_dir: /migrations\nalertmanager:\n  enable_api: true\n  cluster:\n    peers: 'alertmanager.default.svc.cluster.local:9094'\n  sharding_enabled: false\n  sharding_ring:\n    kvstore:\n      store: memberlist\n  external_url: /alertmanager\nalertmanager_storage:\n  backend: s3\n  s3:\n    bucket_name: test-bucket\n    endpoint: s3.test-region.amazonaws.com\nruler:\n  enable_api: true\n  enable_sharding: true\n  sharding_strategy: shuffle-sharding\n  ring:\n   kvstore:\n     store: memberlist\n  alertmanager_url: 'http://alertmanager.default.svc.cluster.local/alertmanager/'\nruler_storage:\n  backend: s3\n  s3:\n    bucket_name: test-bucket\n    endpoint: s3.test-region.amazonaws.com\n",
					}))
				},
			),
			Entry(
				"creates a service account",
				types.NamespacedName{Name: CortexConfigMapName, Namespace: CortexNamespace},
				&corev1.ServiceAccount{},
				func(createdObj client.Object) {
					_, ok := createdObj.(*corev1.ServiceAccount)
					Expect(ok).To(BeTrue())
					// nothing else to check, service account only has a name
					// and a namespace
				},
			),
		)
	})
})

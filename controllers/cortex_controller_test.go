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
	"io/ioutil"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	cortexv1alpha1 "github.com/opstrace/cortex-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Cortex controller", func() {
	const (
		CortexName      = "test-cortex"
		configMapName   = CortexName + CortexConfigMapNameSuffix
		CortexNamespace = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("should manage resouces necessary for a cortex deployment ", func() {

		It("should allow to create a resource", func() {
			ctx := context.Background()
			manifest := filepath.Join("..", "config", "samples", "cortex_v1alpha1_cortex.yaml")
			testCortex := GetCortexTestSample(manifest)
			testCortex.Name = CortexName
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
				types.NamespacedName{Name: configMapName, Namespace: CortexNamespace},
				&corev1.ConfigMap{},
				func(createdObj client.Object) {
					_, ok := createdObj.(*corev1.ConfigMap)

					Expect(ok).To(BeTrue())
					// TODO(sreis): validate config map without webhook defaulting?
				},
			),
			Entry(
				"creates a service account",
				types.NamespacedName{Name: ServiceAccountName, Namespace: CortexNamespace},
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

func GetCortexTestSample(filepath string) *cortexv1alpha1.Cortex {
	b, err := ioutil.ReadFile(filepath)
	Expect(err).ToNot(HaveOccurred())

	deserializer := serializer.NewCodecFactory(k8sClientScheme).UniversalDeserializer()
	obj, _, err := deserializer.Decode(b, nil, nil)
	Expect(err).ToNot(HaveOccurred())
	Expect(obj).To(BeAssignableToTypeOf(&cortexv1alpha1.Cortex{}))

	c := obj.(*cortexv1alpha1.Cortex)
	c.Namespace = "default"

	return c
}

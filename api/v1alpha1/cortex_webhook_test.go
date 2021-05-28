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
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var _ = Describe("Cortex validation webhook", func() {

	BeforeEach(func(done Done) {
		fmt.Fprintf(GinkgoWriter, "BeforeEach\n")
		close(done)
	})

	It("should validate sample", func() {
		manifest := filepath.Join("..", "..", "config", "samples", "cortex_v1alpha1_cortex.yaml")
		c := GetCortexTestSample(manifest)
		// err := c.ValidateCreate()
		err := k8sClient.Create(context.Background(), c)
		Expect(err).ToNot(HaveOccurred())
	})

})

func GetCortexTestSample(filepath string) *Cortex {
	b, err := ioutil.ReadFile(filepath)
	Expect(err).ToNot(HaveOccurred())

	deserializer := serializer.NewCodecFactory(k8sClientScheme).UniversalDeserializer()
	obj, _, err := deserializer.Decode(b, nil, nil)
	Expect(err).ToNot(HaveOccurred())
	Expect(obj).To(BeAssignableToTypeOf(&Cortex{}))

	c := obj.(*Cortex)
	c.Namespace = "default"

	return c
}

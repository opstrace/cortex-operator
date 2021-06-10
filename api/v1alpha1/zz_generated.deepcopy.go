// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Cortex) DeepCopyInto(out *Cortex) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Cortex.
func (in *Cortex) DeepCopy() *Cortex {
	if in == nil {
		return nil
	}
	out := new(Cortex)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Cortex) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CortexList) DeepCopyInto(out *CortexList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Cortex, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CortexList.
func (in *CortexList) DeepCopy() *CortexList {
	if in == nil {
		return nil
	}
	out := new(CortexList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CortexList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CortexSpec) DeepCopyInto(out *CortexSpec) {
	*out = *in
	if in.IngesterSpec != nil {
		in, out := &in.IngesterSpec, &out.IngesterSpec
		*out = new(StatefulSetSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.CompactorSpec != nil {
		in, out := &in.CompactorSpec, &out.CompactorSpec
		*out = new(StatefulSetSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.StoreGatewaySpec != nil {
		in, out := &in.StoreGatewaySpec, &out.StoreGatewaySpec
		*out = new(StatefulSetSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.DistributorSpec != nil {
		in, out := &in.DistributorSpec, &out.DistributorSpec
		*out = new(DeploymentSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.QuerierSpec != nil {
		in, out := &in.QuerierSpec, &out.QuerierSpec
		*out = new(DeploymentSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.QueryFrontendSpec != nil {
		in, out := &in.QueryFrontendSpec, &out.QueryFrontendSpec
		*out = new(DeploymentSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.AlertManagerSpec != nil {
		in, out := &in.AlertManagerSpec, &out.AlertManagerSpec
		*out = new(DeploymentSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.RulerSpec != nil {
		in, out := &in.RulerSpec, &out.RulerSpec
		*out = new(DeploymentSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Memcached != nil {
		in, out := &in.Memcached, &out.Memcached
		*out = new(MemcachedSpec)
		(*in).DeepCopyInto(*out)
	}
	in.Config.DeepCopyInto(&out.Config)
	if in.RuntimeConfig != nil {
		in, out := &in.RuntimeConfig, &out.RuntimeConfig
		*out = new(RuntimeConfigSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CortexSpec.
func (in *CortexSpec) DeepCopy() *CortexSpec {
	if in == nil {
		return nil
	}
	out := new(CortexSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CortexStatus) DeepCopyInto(out *CortexStatus) {
	*out = *in
	if in.MemcachedRef != nil {
		in, out := &in.MemcachedRef, &out.MemcachedRef
		*out = new(MemcachedReference)
		(*in).DeepCopyInto(*out)
	}
	if in.IngesterRef != nil {
		in, out := &in.IngesterRef, &out.IngesterRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.DistributorRef != nil {
		in, out := &in.DistributorRef, &out.DistributorRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.QuerierRef != nil {
		in, out := &in.QuerierRef, &out.QuerierRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.QueryFrontendRef != nil {
		in, out := &in.QueryFrontendRef, &out.QueryFrontendRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.CompactorRef != nil {
		in, out := &in.CompactorRef, &out.CompactorRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.StoreGatewayRef != nil {
		in, out := &in.StoreGatewayRef, &out.StoreGatewayRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.AlertManagerRef != nil {
		in, out := &in.AlertManagerRef, &out.AlertManagerRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.RulerRef != nil {
		in, out := &in.RulerRef, &out.RulerRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CortexStatus.
func (in *CortexStatus) DeepCopy() *CortexStatus {
	if in == nil {
		return nil
	}
	out := new(CortexStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeploymentSpec) DeepCopyInto(out *DeploymentSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeploymentSpec.
func (in *DeploymentSpec) DeepCopy() *DeploymentSpec {
	if in == nil {
		return nil
	}
	out := new(DeploymentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MemcachedReference) DeepCopyInto(out *MemcachedReference) {
	*out = *in
	if in.ChunksCacheRef != nil {
		in, out := &in.ChunksCacheRef, &out.ChunksCacheRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.IndexQueriesCacheRef != nil {
		in, out := &in.IndexQueriesCacheRef, &out.IndexQueriesCacheRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.IndexWritesCacheRef != nil {
		in, out := &in.IndexWritesCacheRef, &out.IndexWritesCacheRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.ResultsCacheRef != nil {
		in, out := &in.ResultsCacheRef, &out.ResultsCacheRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.MetadataCacheRef != nil {
		in, out := &in.MetadataCacheRef, &out.MetadataCacheRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MemcachedReference.
func (in *MemcachedReference) DeepCopy() *MemcachedReference {
	if in == nil {
		return nil
	}
	out := new(MemcachedReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MemcachedSpec) DeepCopyInto(out *MemcachedSpec) {
	*out = *in
	if in.ChunksCacheSpec != nil {
		in, out := &in.ChunksCacheSpec, &out.ChunksCacheSpec
		*out = new(MemcachedStatefulSetSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.IndexQueriesCacheSpec != nil {
		in, out := &in.IndexQueriesCacheSpec, &out.IndexQueriesCacheSpec
		*out = new(MemcachedStatefulSetSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.IndexWritesCacheSpec != nil {
		in, out := &in.IndexWritesCacheSpec, &out.IndexWritesCacheSpec
		*out = new(MemcachedStatefulSetSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.ResultsCacheSpec != nil {
		in, out := &in.ResultsCacheSpec, &out.ResultsCacheSpec
		*out = new(MemcachedStatefulSetSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.MetadataCacheSpec != nil {
		in, out := &in.MetadataCacheSpec, &out.MetadataCacheSpec
		*out = new(MemcachedStatefulSetSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MemcachedSpec.
func (in *MemcachedSpec) DeepCopy() *MemcachedSpec {
	if in == nil {
		return nil
	}
	out := new(MemcachedSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MemcachedStatefulSetSpec) DeepCopyInto(out *MemcachedStatefulSetSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.MemoryLimit != nil {
		in, out := &in.MemoryLimit, &out.MemoryLimit
		*out = new(int32)
		**out = **in
	}
	if in.MaxItemSize != nil {
		in, out := &in.MaxItemSize, &out.MaxItemSize
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MemcachedStatefulSetSpec.
func (in *MemcachedStatefulSetSpec) DeepCopy() *MemcachedStatefulSetSpec {
	if in == nil {
		return nil
	}
	out := new(MemcachedStatefulSetSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RuntimeConfigSpec) DeepCopyInto(out *RuntimeConfigSpec) {
	*out = *in
	in.Overrides.DeepCopyInto(&out.Overrides)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RuntimeConfigSpec.
func (in *RuntimeConfigSpec) DeepCopy() *RuntimeConfigSpec {
	if in == nil {
		return nil
	}
	out := new(RuntimeConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StatefulSetSpec) DeepCopyInto(out *StatefulSetSpec) {
	*out = *in
	if in.DatadirSize != nil {
		in, out := &in.DatadirSize, &out.DatadirSize
		x := (*in).DeepCopy()
		*out = &x
	}
	if in.StorageClassName != nil {
		in, out := &in.StorageClassName, &out.StorageClassName
		*out = new(string)
		**out = **in
	}
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StatefulSetSpec.
func (in *StatefulSetSpec) DeepCopy() *StatefulSetSpec {
	if in == nil {
		return nil
	}
	out := new(StatefulSetSpec)
	in.DeepCopyInto(out)
	return out
}

// +build !ignore_autogenerated

/*
Copyright 2018 Jetstack Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This file was autogenerated by deepcopy-gen. Do not edit it manually!

package v1alpha1

import (
	version "github.com/jetstack/navigator/pkg/cassandra/version"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CassandraCluster) DeepCopyInto(out *CassandraCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CassandraCluster.
func (in *CassandraCluster) DeepCopy() *CassandraCluster {
	if in == nil {
		return nil
	}
	out := new(CassandraCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CassandraCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CassandraClusterList) DeepCopyInto(out *CassandraClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CassandraCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CassandraClusterList.
func (in *CassandraClusterList) DeepCopy() *CassandraClusterList {
	if in == nil {
		return nil
	}
	out := new(CassandraClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CassandraClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CassandraClusterNodePool) DeepCopyInto(out *CassandraClusterNodePool) {
	*out = *in
	in.Persistence.DeepCopyInto(&out.Persistence)
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.Resources.DeepCopyInto(&out.Resources)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CassandraClusterNodePool.
func (in *CassandraClusterNodePool) DeepCopy() *CassandraClusterNodePool {
	if in == nil {
		return nil
	}
	out := new(CassandraClusterNodePool)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CassandraClusterNodePoolStatus) DeepCopyInto(out *CassandraClusterNodePoolStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CassandraClusterNodePoolStatus.
func (in *CassandraClusterNodePoolStatus) DeepCopy() *CassandraClusterNodePoolStatus {
	if in == nil {
		return nil
	}
	out := new(CassandraClusterNodePoolStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CassandraClusterSpec) DeepCopyInto(out *CassandraClusterSpec) {
	*out = *in
	in.NavigatorClusterConfig.DeepCopyInto(&out.NavigatorClusterConfig)
	if in.NodePools != nil {
		in, out := &in.NodePools, &out.NodePools
		*out = make([]CassandraClusterNodePool, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		if *in == nil {
			*out = nil
		} else {
			*out = new(ImageSpec)
			**out = **in
		}
	}
	out.Version = in.Version
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CassandraClusterSpec.
func (in *CassandraClusterSpec) DeepCopy() *CassandraClusterSpec {
	if in == nil {
		return nil
	}
	out := new(CassandraClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CassandraClusterStatus) DeepCopyInto(out *CassandraClusterStatus) {
	*out = *in
	if in.NodePools != nil {
		in, out := &in.NodePools, &out.NodePools
		*out = make(map[string]CassandraClusterNodePoolStatus, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CassandraClusterStatus.
func (in *CassandraClusterStatus) DeepCopy() *CassandraClusterStatus {
	if in == nil {
		return nil
	}
	out := new(CassandraClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CassandraPilotStatus) DeepCopyInto(out *CassandraPilotStatus) {
	*out = *in
	if in.Version != nil {
		in, out := &in.Version, &out.Version
		if *in == nil {
			*out = nil
		} else {
			*out = new(version.Version)
			**out = **in
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CassandraPilotStatus.
func (in *CassandraPilotStatus) DeepCopy() *CassandraPilotStatus {
	if in == nil {
		return nil
	}
	out := new(CassandraPilotStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchCluster) DeepCopyInto(out *ElasticsearchCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchCluster.
func (in *ElasticsearchCluster) DeepCopy() *ElasticsearchCluster {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticsearchCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchClusterList) DeepCopyInto(out *ElasticsearchClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ElasticsearchCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchClusterList.
func (in *ElasticsearchClusterList) DeepCopy() *ElasticsearchClusterList {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticsearchClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchClusterNodePool) DeepCopyInto(out *ElasticsearchClusterNodePool) {
	*out = *in
	if in.Roles != nil {
		in, out := &in.Roles, &out.Roles
		*out = make([]ElasticsearchClusterRole, len(*in))
		copy(*out, *in)
	}
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.Resources.DeepCopyInto(&out.Resources)
	in.Persistence.DeepCopyInto(&out.Persistence)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchClusterNodePool.
func (in *ElasticsearchClusterNodePool) DeepCopy() *ElasticsearchClusterNodePool {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchClusterNodePool)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchClusterNodePoolStatus) DeepCopyInto(out *ElasticsearchClusterNodePoolStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchClusterNodePoolStatus.
func (in *ElasticsearchClusterNodePoolStatus) DeepCopy() *ElasticsearchClusterNodePoolStatus {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchClusterNodePoolStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchClusterSpec) DeepCopyInto(out *ElasticsearchClusterSpec) {
	*out = *in
	in.NavigatorClusterConfig.DeepCopyInto(&out.NavigatorClusterConfig)
	out.Version = in.Version
	if in.Plugins != nil {
		in, out := &in.Plugins, &out.Plugins
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.NodePools != nil {
		in, out := &in.NodePools, &out.NodePools
		*out = make([]ElasticsearchClusterNodePool, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		if *in == nil {
			*out = nil
		} else {
			*out = new(ImageSpec)
			**out = **in
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchClusterSpec.
func (in *ElasticsearchClusterSpec) DeepCopy() *ElasticsearchClusterSpec {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchClusterStatus) DeepCopyInto(out *ElasticsearchClusterStatus) {
	*out = *in
	if in.NodePools != nil {
		in, out := &in.NodePools, &out.NodePools
		*out = make(map[string]ElasticsearchClusterNodePoolStatus, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchClusterStatus.
func (in *ElasticsearchClusterStatus) DeepCopy() *ElasticsearchClusterStatus {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchPilotStatus) DeepCopyInto(out *ElasticsearchPilotStatus) {
	*out = *in
	if in.Documents != nil {
		in, out := &in.Documents, &out.Documents
		if *in == nil {
			*out = nil
		} else {
			*out = new(int64)
			**out = **in
		}
	}
	out.Version = in.Version
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchPilotStatus.
func (in *ElasticsearchPilotStatus) DeepCopy() *ElasticsearchPilotStatus {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchPilotStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ImageSpec) DeepCopyInto(out *ImageSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ImageSpec.
func (in *ImageSpec) DeepCopy() *ImageSpec {
	if in == nil {
		return nil
	}
	out := new(ImageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NavigatorClusterConfig) DeepCopyInto(out *NavigatorClusterConfig) {
	*out = *in
	out.PilotImage = in.PilotImage
	in.SecurityContext.DeepCopyInto(&out.SecurityContext)
	if in.Sysctls != nil {
		in, out := &in.Sysctls, &out.Sysctls
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NavigatorClusterConfig.
func (in *NavigatorClusterConfig) DeepCopy() *NavigatorClusterConfig {
	if in == nil {
		return nil
	}
	out := new(NavigatorClusterConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NavigatorSecurityContext) DeepCopyInto(out *NavigatorSecurityContext) {
	*out = *in
	if in.RunAsUser != nil {
		in, out := &in.RunAsUser, &out.RunAsUser
		if *in == nil {
			*out = nil
		} else {
			*out = new(int64)
			**out = **in
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NavigatorSecurityContext.
func (in *NavigatorSecurityContext) DeepCopy() *NavigatorSecurityContext {
	if in == nil {
		return nil
	}
	out := new(NavigatorSecurityContext)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PersistenceConfig) DeepCopyInto(out *PersistenceConfig) {
	*out = *in
	out.Size = in.Size.DeepCopy()
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PersistenceConfig.
func (in *PersistenceConfig) DeepCopy() *PersistenceConfig {
	if in == nil {
		return nil
	}
	out := new(PersistenceConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Pilot) DeepCopyInto(out *Pilot) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Pilot.
func (in *Pilot) DeepCopy() *Pilot {
	if in == nil {
		return nil
	}
	out := new(Pilot)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Pilot) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PilotCondition) DeepCopyInto(out *PilotCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PilotCondition.
func (in *PilotCondition) DeepCopy() *PilotCondition {
	if in == nil {
		return nil
	}
	out := new(PilotCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PilotElasticsearchSpec) DeepCopyInto(out *PilotElasticsearchSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PilotElasticsearchSpec.
func (in *PilotElasticsearchSpec) DeepCopy() *PilotElasticsearchSpec {
	if in == nil {
		return nil
	}
	out := new(PilotElasticsearchSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PilotList) DeepCopyInto(out *PilotList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Pilot, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PilotList.
func (in *PilotList) DeepCopy() *PilotList {
	if in == nil {
		return nil
	}
	out := new(PilotList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PilotList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PilotSpec) DeepCopyInto(out *PilotSpec) {
	*out = *in
	if in.Elasticsearch != nil {
		in, out := &in.Elasticsearch, &out.Elasticsearch
		if *in == nil {
			*out = nil
		} else {
			*out = new(PilotElasticsearchSpec)
			**out = **in
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PilotSpec.
func (in *PilotSpec) DeepCopy() *PilotSpec {
	if in == nil {
		return nil
	}
	out := new(PilotSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PilotStatus) DeepCopyInto(out *PilotStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]PilotCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Elasticsearch != nil {
		in, out := &in.Elasticsearch, &out.Elasticsearch
		if *in == nil {
			*out = nil
		} else {
			*out = new(ElasticsearchPilotStatus)
			(*in).DeepCopyInto(*out)
		}
	}
	if in.Cassandra != nil {
		in, out := &in.Cassandra, &out.Cassandra
		if *in == nil {
			*out = nil
		} else {
			*out = new(CassandraPilotStatus)
			(*in).DeepCopyInto(*out)
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PilotStatus.
func (in *PilotStatus) DeepCopy() *PilotStatus {
	if in == nil {
		return nil
	}
	out := new(PilotStatus)
	in.DeepCopyInto(out)
	return out
}

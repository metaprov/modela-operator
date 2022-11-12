//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApiGatewaySpec) DeepCopyInto(out *ApiGatewaySpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(v1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApiGatewaySpec.
func (in *ApiGatewaySpec) DeepCopy() *ApiGatewaySpec {
	if in == nil {
		return nil
	}
	out := new(ApiGatewaySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertManagerSpec) DeepCopyInto(out *CertManagerSpec) {
	*out = *in
	in.Values.DeepCopyInto(&out.Values)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertManagerSpec.
func (in *CertManagerSpec) DeepCopy() *CertManagerSpec {
	if in == nil {
		return nil
	}
	out := new(CertManagerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControlPlaneSpec) DeepCopyInto(out *ControlPlaneSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(v1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControlPlaneSpec.
func (in *ControlPlaneSpec) DeepCopy() *ControlPlaneSpec {
	if in == nil {
		return nil
	}
	out := new(ControlPlaneSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneSpec) DeepCopyInto(out *DataPlaneSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(v1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneSpec.
func (in *DataPlaneSpec) DeepCopy() *DataPlaneSpec {
	if in == nil {
		return nil
	}
	out := new(DataPlaneSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Modela) DeepCopyInto(out *Modela) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Modela.
func (in *Modela) DeepCopy() *Modela {
	if in == nil {
		return nil
	}
	out := new(Modela)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Modela) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModelaAccessSpec) DeepCopyInto(out *ModelaAccessSpec) {
	*out = *in
	in.NginxValues.DeepCopyInto(&out.NginxValues)
	if in.Hostname != nil {
		in, out := &in.Hostname, &out.Hostname
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModelaAccessSpec.
func (in *ModelaAccessSpec) DeepCopy() *ModelaAccessSpec {
	if in == nil {
		return nil
	}
	out := new(ModelaAccessSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModelaCondition) DeepCopyInto(out *ModelaCondition) {
	*out = *in
	if in.LastTransitionTime != nil {
		in, out := &in.LastTransitionTime, &out.LastTransitionTime
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModelaCondition.
func (in *ModelaCondition) DeepCopy() *ModelaCondition {
	if in == nil {
		return nil
	}
	out := new(ModelaCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModelaLicenseSpec) DeepCopyInto(out *ModelaLicenseSpec) {
	*out = *in
	if in.LicenseKey != nil {
		in, out := &in.LicenseKey, &out.LicenseKey
		*out = new(string)
		**out = **in
	}
	if in.LinkLicense != nil {
		in, out := &in.LinkLicense, &out.LinkLicense
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModelaLicenseSpec.
func (in *ModelaLicenseSpec) DeepCopy() *ModelaLicenseSpec {
	if in == nil {
		return nil
	}
	out := new(ModelaLicenseSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModelaList) DeepCopyInto(out *ModelaList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Modela, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModelaList.
func (in *ModelaList) DeepCopy() *ModelaList {
	if in == nil {
		return nil
	}
	out := new(ModelaList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ModelaList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModelaSpec) DeepCopyInto(out *ModelaSpec) {
	*out = *in
	in.Observability.DeepCopyInto(&out.Observability)
	in.Ingress.DeepCopyInto(&out.Ingress)
	in.License.DeepCopyInto(&out.License)
	if in.Tenants != nil {
		in, out := &in.Tenants, &out.Tenants
		*out = make([]*TenantSpec, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(TenantSpec)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	in.CertManager.DeepCopyInto(&out.CertManager)
	in.ObjectStore.DeepCopyInto(&out.ObjectStore)
	in.SystemDatabase.DeepCopyInto(&out.SystemDatabase)
	in.ControlPlane.DeepCopyInto(&out.ControlPlane)
	in.DataPlane.DeepCopyInto(&out.DataPlane)
	in.ApiGateway.DeepCopyInto(&out.ApiGateway)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModelaSpec.
func (in *ModelaSpec) DeepCopy() *ModelaSpec {
	if in == nil {
		return nil
	}
	out := new(ModelaSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModelaStatus) DeepCopyInto(out *ModelaStatus) {
	*out = *in
	if in.Tenants != nil {
		in, out := &in.Tenants, &out.Tenants
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.LicenseToken != nil {
		in, out := &in.LicenseToken, &out.LicenseToken
		*out = new(v1.ObjectReference)
		**out = **in
	}
	if in.LinkLicenseUrl != nil {
		in, out := &in.LinkLicenseUrl, &out.LinkLicenseUrl
		*out = new(string)
		**out = **in
	}
	if in.FailureMessage != nil {
		in, out := &in.FailureMessage, &out.FailureMessage
		*out = new(string)
		**out = **in
	}
	if in.LastUpdated != nil {
		in, out := &in.LastUpdated, &out.LastUpdated
		*out = (*in).DeepCopy()
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]ModelaCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModelaStatus.
func (in *ModelaStatus) DeepCopy() *ModelaStatus {
	if in == nil {
		return nil
	}
	out := new(ModelaStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ObjectStorageSpec) DeepCopyInto(out *ObjectStorageSpec) {
	*out = *in
	in.Values.DeepCopyInto(&out.Values)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ObjectStorageSpec.
func (in *ObjectStorageSpec) DeepCopy() *ObjectStorageSpec {
	if in == nil {
		return nil
	}
	out := new(ObjectStorageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ObservabilitySpec) DeepCopyInto(out *ObservabilitySpec) {
	*out = *in
	in.PrometheusValues.DeepCopyInto(&out.PrometheusValues)
	in.LokiValues.DeepCopyInto(&out.LokiValues)
	in.GrafanaValues.DeepCopyInto(&out.GrafanaValues)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ObservabilitySpec.
func (in *ObservabilitySpec) DeepCopy() *ObservabilitySpec {
	if in == nil {
		return nil
	}
	out := new(ObservabilitySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemDatabaseSpec) DeepCopyInto(out *SystemDatabaseSpec) {
	*out = *in
	in.Values.DeepCopyInto(&out.Values)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemDatabaseSpec.
func (in *SystemDatabaseSpec) DeepCopy() *SystemDatabaseSpec {
	if in == nil {
		return nil
	}
	out := new(SystemDatabaseSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TenantSpec) DeepCopyInto(out *TenantSpec) {
	*out = *in
	if in.AdminPassword != nil {
		in, out := &in.AdminPassword, &out.AdminPassword
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantSpec.
func (in *TenantSpec) DeepCopy() *TenantSpec {
	if in == nil {
		return nil
	}
	out := new(TenantSpec)
	in.DeepCopyInto(out)
	return out
}
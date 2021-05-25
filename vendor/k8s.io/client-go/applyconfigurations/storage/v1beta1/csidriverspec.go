/*
Copyright The Kubernetes Authors.

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

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1beta1

import (
	v1beta1 "k8s.io/api/storage/v1beta1"
)

// CSIDriverSpecApplyConfiguration represents an declarative configuration of the CSIDriverSpec type for use
// with apply.
type CSIDriverSpecApplyConfiguration struct {
	AttachRequired               *bool                            `json:"attachRequired,omitempty"`
	PodInfoOnMount               *bool                            `json:"podInfoOnMount,omitempty"`
	VolumeLifecycleModes         []v1beta1.VolumeLifecycleMode    `json:"volumeLifecycleModes,omitempty"`
	StorageCapacity              *bool                            `json:"storageCapacity,omitempty"`
	FSGroupPolicy                *v1beta1.FSGroupPolicy           `json:"fsGroupPolicy,omitempty"`
	TokenRequests                []TokenRequestApplyConfiguration `json:"tokenRequests,omitempty"`
	RequiresRepublish            *bool                            `json:"requiresRepublish,omitempty"`
	RecoveryFromExpansionFailure *bool                            `json:"recoveryFromExpansionFailure,omitempty"`
}

// CSIDriverSpecApplyConfiguration constructs an declarative configuration of the CSIDriverSpec type for use with
// apply.
func CSIDriverSpec() *CSIDriverSpecApplyConfiguration {
	return &CSIDriverSpecApplyConfiguration{}
}

// WithAttachRequired sets the AttachRequired field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the AttachRequired field is set to the value of the last call.
func (b *CSIDriverSpecApplyConfiguration) WithAttachRequired(value bool) *CSIDriverSpecApplyConfiguration {
	b.AttachRequired = &value
	return b
}

// WithPodInfoOnMount sets the PodInfoOnMount field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the PodInfoOnMount field is set to the value of the last call.
func (b *CSIDriverSpecApplyConfiguration) WithPodInfoOnMount(value bool) *CSIDriverSpecApplyConfiguration {
	b.PodInfoOnMount = &value
	return b
}

// WithVolumeLifecycleModes adds the given value to the VolumeLifecycleModes field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the VolumeLifecycleModes field.
func (b *CSIDriverSpecApplyConfiguration) WithVolumeLifecycleModes(values ...v1beta1.VolumeLifecycleMode) *CSIDriverSpecApplyConfiguration {
	for i := range values {
		b.VolumeLifecycleModes = append(b.VolumeLifecycleModes, values[i])
	}
	return b
}

// WithStorageCapacity sets the StorageCapacity field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the StorageCapacity field is set to the value of the last call.
func (b *CSIDriverSpecApplyConfiguration) WithStorageCapacity(value bool) *CSIDriverSpecApplyConfiguration {
	b.StorageCapacity = &value
	return b
}

// WithFSGroupPolicy sets the FSGroupPolicy field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the FSGroupPolicy field is set to the value of the last call.
func (b *CSIDriverSpecApplyConfiguration) WithFSGroupPolicy(value v1beta1.FSGroupPolicy) *CSIDriverSpecApplyConfiguration {
	b.FSGroupPolicy = &value
	return b
}

// WithTokenRequests adds the given value to the TokenRequests field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the TokenRequests field.
func (b *CSIDriverSpecApplyConfiguration) WithTokenRequests(values ...*TokenRequestApplyConfiguration) *CSIDriverSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithTokenRequests")
		}
		b.TokenRequests = append(b.TokenRequests, *values[i])
	}
	return b
}

// WithRequiresRepublish sets the RequiresRepublish field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RequiresRepublish field is set to the value of the last call.
func (b *CSIDriverSpecApplyConfiguration) WithRequiresRepublish(value bool) *CSIDriverSpecApplyConfiguration {
	b.RequiresRepublish = &value
	return b
}

// WithRecoveryFromExpansionFailure sets the RecoveryFromExpansionFailure field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RecoveryFromExpansionFailure field is set to the value of the last call.
func (b *CSIDriverSpecApplyConfiguration) WithRecoveryFromExpansionFailure(value bool) *CSIDriverSpecApplyConfiguration {
	b.RecoveryFromExpansionFailure = &value
	return b
}

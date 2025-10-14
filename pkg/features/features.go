/*
Copyright 2018 The Kubernetes Authors.

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

package features

import (
	"slices"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/discovery"
	"k8s.io/component-base/featuregate"
)

const (
	// owner: @sunpa93
	// alpha: v1.22
	AnnotateFsResize featuregate.Feature = "AnnotateFsResize"

	// owner: @gnufied
	// alpha: v1.23
	// beta: v1.32
	//
	// Allows users to recover from volume expansion failures
	RecoverVolumeExpansionFailure featuregate.Feature = "RecoverVolumeExpansionFailure"

	// owner: @sunnylovestiramisu
	// kep: https://kep.k8s.io/3751
	// alpha: v1.29
	// beta: v1.31
	// GA: v1.34
	//
	// Pass VolumeAttributesClass parameters to supporting CSI drivers during ModifyVolume
	VolumeAttributesClass featuregate.Feature = "VolumeAttributesClass"

	// owner: @rhrmo
	// alpha: v1.34
	//
	// Releases leader election lease on sigterm / sigint.
	ReleaseLeaderElectionOnExit featuregate.Feature = "ReleaseLeaderElectionOnExit"
)

func init() {
	utilfeature.DefaultMutableFeatureGate.Add(defaultResizerFeatureGates)
}

var defaultResizerFeatureGates = map[featuregate.Feature]featuregate.FeatureSpec{
	AnnotateFsResize:              {Default: false, PreRelease: featuregate.Alpha},
	RecoverVolumeExpansionFailure: {Default: true, PreRelease: featuregate.Beta},
	VolumeAttributesClass:         {Default: true, PreRelease: featuregate.GA},
	ReleaseLeaderElectionOnExit:   {Default: false, PreRelease: featuregate.Alpha},
}

// IsVolumeAttributesClassV1Enabled checks if the VolumeAttributesClass v1 API is enabled.
func IsVolumeAttributesClassV1Enabled(d discovery.DiscoveryInterface) (bool, error) {
	return resourceExists(d, "storage.k8s.io/v1", "VolumeAttributesClass")
}

func resourceExists(d discovery.DiscoveryInterface, groupVersion, kind string) (bool, error) {
	res, err := d.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return slices.ContainsFunc(res.APIResources, func(r metav1.APIResource) bool {
		return r.Kind == kind
	}), nil
}

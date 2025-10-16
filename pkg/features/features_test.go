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
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fake "k8s.io/client-go/kubernetes/fake"
)

// kubectl get --raw '/apis/storage.k8s.io/v1'
const (
	discoveryV1_34 = `{"kind":"APIResourceList","apiVersion":"v1","groupVersion":"storage.k8s.io/v1","resources":[{"name":"csidrivers","singularName":"csidriver","namespaced":false,"kind":"CSIDriver","verbs":["create","delete","deletecollection","get","list","patch","update","watch"],"storageVersionHash":"hL6j/rwBV5w="},{"name":"csinodes","singularName":"csinode","namespaced":false,"kind":"CSINode","verbs":["create","delete","deletecollection","get","list","patch","update","watch"],"storageVersionHash":"Pe62DkZtjuo="},{"name":"csistoragecapacities","singularName":"csistoragecapacity","namespaced":true,"kind":"CSIStorageCapacity","verbs":["create","delete","deletecollection","get","list","patch","update","watch"],"storageVersionHash":"xeVl+2Ly1kE="},{"name":"storageclasses","singularName":"storageclass","namespaced":false,"kind":"StorageClass","verbs":["create","delete","deletecollection","get","list","patch","update","watch"],"shortNames":["sc"],"storageVersionHash":"K+m6uJwbjGY="},{"name":"volumeattachments","singularName":"volumeattachment","namespaced":false,"kind":"VolumeAttachment","verbs":["create","delete","deletecollection","get","list","patch","update","watch"],"storageVersionHash":"tJx/ezt6UDU="},{"name":"volumeattachments/status","singularName":"","namespaced":false,"kind":"VolumeAttachment","verbs":["get","patch","update"]},{"name":"volumeattributesclasses","singularName":"volumeattributesclass","namespaced":false,"kind":"VolumeAttributesClass","verbs":["create","delete","deletecollection","get","list","patch","update","watch"],"shortNames":["vac"],"storageVersionHash":"Bl3MtjZ/n/s="}]}`
	discoveryV1_32 = `{"kind":"APIResourceList","apiVersion":"v1","groupVersion":"storage.k8s.io/v1","resources":[{"name":"csidrivers","singularName":"csidriver","namespaced":false,"kind":"CSIDriver","verbs":["create","delete","deletecollection","get","list","patch","update","watch"],"storageVersionHash":"hL6j/rwBV5w="},{"name":"csinodes","singularName":"csinode","namespaced":false,"kind":"CSINode","verbs":["create","delete","deletecollection","get","list","patch","update","watch"],"storageVersionHash":"Pe62DkZtjuo="},{"name":"csistoragecapacities","singularName":"csistoragecapacity","namespaced":true,"kind":"CSIStorageCapacity","verbs":["create","delete","deletecollection","get","list","patch","update","watch"],"storageVersionHash":"xeVl+2Ly1kE="},{"name":"storageclasses","singularName":"storageclass","namespaced":false,"kind":"StorageClass","verbs":["create","delete","deletecollection","get","list","patch","update","watch"],"shortNames":["sc"],"storageVersionHash":"K+m6uJwbjGY="},{"name":"volumeattachments","singularName":"volumeattachment","namespaced":false,"kind":"VolumeAttachment","verbs":["create","delete","deletecollection","get","list","patch","update","watch"],"storageVersionHash":"tJx/ezt6UDU="},{"name":"volumeattachments/status","singularName":"","namespaced":false,"kind":"VolumeAttachment","verbs":["get","patch","update"]}]}`
)

func TestIsVolumeAttributesClassV1Enabled(t *testing.T) {
	cases := []struct {
		name    string
		resp    string
		enabled bool
	}{
		{
			name:    "v1.34",
			resp:    discoveryV1_34,
			enabled: true,
		}, {
			name:    "v1.32",
			resp:    discoveryV1_32,
			enabled: false,
		}, {
			name:    "none",
			resp:    `{}`,
			enabled: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var l metav1.APIResourceList
			err := json.Unmarshal([]byte(tc.resp), &l)
			if err != nil {
				t.Fatalf("error unmarshalling discovery response: %v", err)
			}

			client := fake.NewSimpleClientset()
			client.Fake.Resources = append(client.Fake.Resources, &l)
			enabled, err := IsVolumeAttributesClassV1Enabled(client.Discovery())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if enabled != tc.enabled {
				t.Errorf("expected %v, got %v", tc.enabled, enabled)
			}
		})
	}

}

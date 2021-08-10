/*
Copyright 2020 The Kubernetes Authors.

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

package controller

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const (
	defaultNS       = "default"
	defaultPodName  = "pod1"
	defaultNodeName = "node1"
	defaultUID      = "uid1"
)

func pod() *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultPodName,
			Namespace: defaultNS,
			UID:       defaultUID,
		},
		Spec: v1.PodSpec{
			NodeName: defaultNodeName,
		},
		Status: v1.PodStatus{
			Phase: v1.PodPending,
		},
	}
}

func withPVC(pvcName string, pod *v1.Pod) *v1.Pod {
	volume := v1.Volume{
		Name: pvcName,
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvcName,
			},
		},
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes, volume)
	return pod
}

func withUID(uid types.UID, pod *v1.Pod) *v1.Pod {
	pod.ObjectMeta.UID = uid
	return pod
}

func withStatus(phase v1.PodPhase, pod *v1.Pod) *v1.Pod {
	pod.Status.Phase = phase
	return pod
}

func TestAddRemovePods(t *testing.T) {
	tests := []struct {
		name             string
		initialObjects   []runtime.Object
		addPod           *v1.Pod
		removePod        *v1.Pod
		expectedInUsePVC map[string]map[UniquePodName]UniquePodName
	}{
		{
			name:           "with no initial pods",
			initialObjects: []runtime.Object{},
			addPod:         withPVC("no-init-pod", pod()),
			expectedInUsePVC: map[string]map[UniquePodName]UniquePodName{
				"default/no-init-pod": {
					UniquePodName(defaultUID): UniquePodName(defaultUID),
				},
			},
		},
		{
			name:             "adding failed pod",
			initialObjects:   []runtime.Object{},
			addPod:           withStatus(v1.PodFailed, withPVC("foobar", pod())),
			expectedInUsePVC: map[string]map[UniquePodName]UniquePodName{},
		},
		{
			name:             "with success pod",
			initialObjects:   []runtime.Object{},
			addPod:           withStatus(v1.PodSucceeded, withPVC("foobar", pod())),
			expectedInUsePVC: map[string]map[UniquePodName]UniquePodName{},
		},
		{
			name: "removing running pod",
			initialObjects: []runtime.Object{
				withUID(types.UID("foobar-pod"), pod()),
			},
			removePod:        withUID(types.UID("foobar-pod"), pod()),
			expectedInUsePVC: map[string]map[UniquePodName]UniquePodName{},
		},
		{
			name: "removing failed pod",
			initialObjects: []runtime.Object{
				withUID(types.UID("foobar-pod"), pod()),
			},
			removePod:        withStatus(v1.PodFailed, withUID(types.UID("foobar-pod"), pod())),
			expectedInUsePVC: map[string]map[UniquePodName]UniquePodName{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pvcStore := newUsedPVCStore()

			if len(test.initialObjects) > 0 {
				for _, commonObj := range test.initialObjects {
					podObj, ok := commonObj.(*v1.Pod)
					if ok {
						pvcStore.addPod(podObj)
					}
				}
			}

			if test.addPod != nil {
				pvcStore.addPod(test.addPod)
			}

			if test.removePod != nil {
				pvcStore.removePod(test.removePod)
			}

			expectedPVCs := pvcStore.inUsePVC

			if len(test.expectedInUsePVC) == 0 && len(expectedPVCs) != 0 {
				t.Errorf("for %s: expected no in-use PVCs found %v", test.name, expectedPVCs)
				return
			}

			if !reflect.DeepEqual(test.expectedInUsePVC, expectedPVCs) {
				t.Errorf("for %s: expected %v got %v", test.name, test.expectedInUsePVC, expectedPVCs)
			}
		})
	}
}

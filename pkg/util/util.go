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

package util

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
)

var knownResizeConditions = map[v1.PersistentVolumeClaimConditionType]bool{
	v1.PersistentVolumeClaimResizing:                true,
	v1.PersistentVolumeClaimFileSystemResizePending: true,
}

// PVCKey returns an unique key of a PVC object,
func PVCKey(pvc *v1.PersistentVolumeClaim) string {
	return fmt.Sprintf("%s/%s", pvc.Namespace, pvc.Name)
}

// MergeResizeConditionsOfPVC updates pvc with requested resize conditions
// leaving other conditions untouched.
func MergeResizeConditionsOfPVC(oldConditions, newConditions []v1.PersistentVolumeClaimCondition) []v1.PersistentVolumeClaimCondition {
	newConditionSet := make(map[v1.PersistentVolumeClaimConditionType]v1.PersistentVolumeClaimCondition, len(newConditions))
	for _, condition := range newConditions {
		newConditionSet[condition.Type] = condition
	}

	var resultConditions []v1.PersistentVolumeClaimCondition
	for _, condition := range oldConditions {
		// If Condition is of not resize type, we keep it.
		if _, ok := knownResizeConditions[condition.Type]; !ok {
			newConditions = append(newConditions, condition)
			continue
		}
		if newCondition, ok := newConditionSet[condition.Type]; ok {
			// Use the new condition to replace old condition with same type.
			resultConditions = append(resultConditions, newCondition)
			delete(newConditionSet, condition.Type)
		}

		// Drop old conditions whose type not exist in new conditions.
	}

	// Append remains resize conditions.
	for _, condition := range newConditionSet {
		resultConditions = append(resultConditions, condition)
	}

	return resultConditions
}

// PatchPVCStatus updates PVC status using PATCH verb
func PatchPVCStatus(
	oldPVC *v1.PersistentVolumeClaim,
	newPVC *v1.PersistentVolumeClaim,
	kubeClient kubernetes.Interface) (*v1.PersistentVolumeClaim, error) {
	patchBytes, err := getPVCPatchData(oldPVC, newPVC)
	if err != nil {
		return nil, fmt.Errorf("can't patch status of PVC %s as generate path data failed: %v", PVCKey(oldPVC), err)
	}
	updatedClaim, updateErr := kubeClient.CoreV1().PersistentVolumeClaims(oldPVC.Namespace).
		Patch(context.TODO(), oldPVC.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{}, "status")
	if updateErr != nil {
		return nil, fmt.Errorf("can't patch status of  PVC %s with %v", PVCKey(oldPVC), updateErr)
	}
	return updatedClaim, nil
}

func getPVCPatchData(oldPVC, newPVC *v1.PersistentVolumeClaim) ([]byte, error) {
	patchBytes, err := getPatchData(oldPVC, newPVC)
	if err != nil {
		return patchBytes, err
	}

	patchBytes, err = addResourceVersion(patchBytes, oldPVC.ResourceVersion)
	if err != nil {
		return nil, fmt.Errorf("apply ResourceVersion to patch data failed: %v", err)
	}
	return patchBytes, nil
}

func addResourceVersion(patchBytes []byte, resourceVersion string) ([]byte, error) {
	var patchMap map[string]interface{}
	err := json.Unmarshal(patchBytes, &patchMap)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling patch with %v", err)
	}
	u := unstructured.Unstructured{Object: patchMap}
	a, err := meta.Accessor(&u)
	if err != nil {
		return nil, fmt.Errorf("error creating accessor with  %v", err)
	}
	a.SetResourceVersion(resourceVersion)
	versionBytes, err := json.Marshal(patchMap)
	if err != nil {
		return nil, fmt.Errorf("error marshalling json patch with %v", err)
	}
	return versionBytes, nil
}

// UpdatePVCapacity updates PVC capacity with requested size.
func UpdatePVCapacity(pv *v1.PersistentVolume, newCapacity resource.Quantity, kubeClient kubernetes.Interface) error {
	newPV := pv.DeepCopy()
	newPV.Spec.Capacity[v1.ResourceStorage] = newCapacity
	patchBytes, err := getPatchData(pv, newPV)
	if err != nil {
		return fmt.Errorf("can't update capacity of PV %s as generate path data failed: %v", pv.Name, err)
	}
	_, updateErr := kubeClient.CoreV1().PersistentVolumes().Patch(context.TODO(), pv.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	if updateErr != nil {
		return fmt.Errorf("update capacity of PV %s failed: %v", pv.Name, updateErr)
	}
	return nil
}

func getPatchData(oldObj, newObj interface{}) ([]byte, error) {
	oldData, err := json.Marshal(oldObj)
	if err != nil {
		return nil, fmt.Errorf("marshal old object failed: %v", err)
	}
	newData, err := json.Marshal(newObj)
	if err != nil {
		return nil, fmt.Errorf("marshal new object failed: %v", err)
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, oldObj)
	if err != nil {
		return nil, fmt.Errorf("CreateTwoWayMergePatch failed: %v", err)
	}
	return patchBytes, nil
}

// HasFileSystemResizePendingCondition returns true if a pvc has a FileSystemResizePending condition.
// This means the controller side resize operation is finished, and kubelet side operation is in progress.
func HasFileSystemResizePendingCondition(pvc *v1.PersistentVolumeClaim) bool {
	for _, condition := range pvc.Status.Conditions {
		if condition.Type == v1.PersistentVolumeClaimFileSystemResizePending && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

// SanitizeName changes any name to a sanitized name which can be accepted by kubernetes.
func SanitizeName(name string) string {
	re := regexp.MustCompile("[^a-zA-Z0-9-]")
	name = re.ReplaceAllString(name, "-")
	if name[len(name)-1] == '-' {
		// name must not end with '-'
		name = name + "X"
	}
	return name
}

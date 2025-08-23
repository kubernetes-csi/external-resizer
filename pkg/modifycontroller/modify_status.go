/*
Copyright 2023 The Kubernetes Authors.

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

package modifycontroller

import (
	"fmt"

	"github.com/kubernetes-csi/external-resizer/pkg/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// markControllerModifyVolumeStatus will mark ModifyVolumeStatus other than completed in the PVC
func (ctrl *modifyController) markControllerModifyVolumeStatus(
	pvc *v1.PersistentVolumeClaim,
	modifyVolumeStatus v1.PersistentVolumeClaimModifyVolumeStatus,
	err error) (*v1.PersistentVolumeClaim, error) {

	newPVC := pvc.DeepCopy()
	newPVC.Status.ModifyVolumeStatus = &v1.ModifyVolumeStatus{
		Status:                          modifyVolumeStatus,
		TargetVolumeAttributesClassName: ptr.Deref(pvc.Spec.VolumeAttributesClassName, ""),
	}
	// Update PVC's Condition to indicate modification
	pvcCondition := v1.PersistentVolumeClaimCondition{
		Type:               v1.PersistentVolumeClaimVolumeModifyingVolume,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
	}
	conditionMessage := ""
	switch modifyVolumeStatus {
	case v1.PersistentVolumeClaimModifyVolumeInProgress:
		conditionMessage = "ModifyVolume operation in progress."
	case v1.PersistentVolumeClaimModifyVolumeInfeasible:
		newPVC.Status.ModifyVolumeStatus.TargetVolumeAttributesClassName = pvc.Status.ModifyVolumeStatus.TargetVolumeAttributesClassName
		conditionMessage = "ModifyVolume failed with error " + err.Error() + ". Waiting for retry."
	}
	pvcCondition.Message = conditionMessage
	// Do not change conditions for pending modifications and keep existing conditions
	if modifyVolumeStatus != v1.PersistentVolumeClaimModifyVolumePending {
		newPVC.Status.Conditions = util.MergeModifyVolumeConditionsOfPVC(newPVC.Status.Conditions,
			[]v1.PersistentVolumeClaimCondition{pvcCondition})
	}

	updatedPVC, err := util.PatchClaim(ctrl.kubeClient, pvc, newPVC, true /* addResourceVersionCheck */)
	if err != nil {
		return pvc, fmt.Errorf("mark PVC %q as modify volume failed, errored with: %v", pvc.Name, err)
	}
	return updatedPVC, nil
}

func (ctrl *modifyController) updateConditionBasedOnError(pvc *v1.PersistentVolumeClaim, err error) (*v1.PersistentVolumeClaim, error) {
	newPVC := pvc.DeepCopy()
	pvcCondition := v1.PersistentVolumeClaimCondition{
		Type:               v1.PersistentVolumeClaimVolumeModifyVolumeError,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            "ModifyVolume failed with error. Waiting for retry.",
	}

	if err != nil {
		pvcCondition.Message = "ModifyVolume failed with error: " + err.Error() + ". Waiting for retry."
	}

	newPVC.Status.Conditions = util.MergeModifyVolumeConditionsOfPVC(newPVC.Status.Conditions,
		[]v1.PersistentVolumeClaimCondition{pvcCondition})

	updatedPVC, err := util.PatchClaim(ctrl.kubeClient, pvc, newPVC, false /* addResourceVersionCheck */)
	if err != nil {
		return pvc, fmt.Errorf("mark PVC %q as controller expansion failed, errored with: %v", pvc.Name, err)
	}
	return updatedPVC, nil
}

// markControllerModifyVolumeStatus will mark ModifyVolumeStatus as completed in the PVC
// and update CurrentVolumeAttributesClassName, clear the conditions
func (ctrl *modifyController) markControllerModifyVolumeCompleted(pvc *v1.PersistentVolumeClaim, pv *v1.PersistentVolume) (*v1.PersistentVolumeClaim, *v1.PersistentVolume, error) {
	modifiedVacName := pvc.Status.ModifyVolumeStatus.TargetVolumeAttributesClassName

	// Update PVC
	newPVC := pvc.DeepCopy()

	// Update ModifyVolumeStatus to completed
	newPVC.Status.ModifyVolumeStatus = nil

	// Update CurrentVolumeAttributesClassName
	newPVC.Status.CurrentVolumeAttributesClassName = &modifiedVacName

	// Clear all the conditions related to modify volume
	newPVC.Status.Conditions = clearModifyVolumeConditions(newPVC.Status.Conditions)

	// Update PV
	newPV := pv.DeepCopy()
	newPV.Spec.VolumeAttributesClassName = &modifiedVacName

	// Update PV before PVC to avoid PV not getting updated but PVC did
	updatedPV, err := util.PatchPersistentVolume(ctrl.kubeClient, pv, newPV)
	if err != nil {
		return pvc, pv, fmt.Errorf("update pv.Spec.VolumeAttributesClassName for PVC %q failed, errored with: %v", pvc.Name, err)
	}
	updatedPVC, err := util.PatchClaim(ctrl.kubeClient, pvc, newPVC, false /* addResourceVersionCheck */)

	if err != nil {
		return pvc, pv, fmt.Errorf("mark PVC %q as ModifyVolumeCompleted failed, errored with: %v", pvc.Name, err)
	}

	return updatedPVC, updatedPV, nil
}

// markControllerModifyVolumeStatus clears all the conditions related to modify volume and only
// leave other condition types
func clearModifyVolumeConditions(conditions []v1.PersistentVolumeClaimCondition) []v1.PersistentVolumeClaimCondition {
	knownConditions := []v1.PersistentVolumeClaimCondition{}
	for _, value := range conditions {
		// Only keep conditions that are not related to modify volume
		if value.Type != v1.PersistentVolumeClaimVolumeModifyVolumeError && value.Type != v1.PersistentVolumeClaimVolumeModifyingVolume {
			knownConditions = append(knownConditions, value)
		}
	}
	return knownConditions
}

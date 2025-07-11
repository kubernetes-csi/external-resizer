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
	storagev1beta1 "k8s.io/api/storage/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

const (
	pvcNameKey      = "csi.storage.k8s.io/pvc/name"
	pvcNamespaceKey = "csi.storage.k8s.io/pvc/namespace"
	pvNameKey       = "csi.storage.k8s.io/pv/name"
)

// The return value bool is only used as a sentinel value when function returns without actually performing modification
func (ctrl *modifyController) modify(pvc *v1.PersistentVolumeClaim, pv *v1.PersistentVolume) (*v1.PersistentVolumeClaim, *v1.PersistentVolume, error, bool) {
	pvcSpecVACName := pvc.Spec.VolumeAttributesClassName

	if !isModificationNeeded(pvc) {
		return pvc, pv, nil, false
	}

	// Requeue PVC if modification recently failed with infeasible error.
	if recentlyInfeasibleErr := ctrl.delayModificationIfRecentlyInfeasible(pvc); recentlyInfeasibleErr != nil {
		return pvc, pv, recentlyInfeasibleErr, false
	}

	// Validate VAC exists
	vac, err := ctrl.vacLister.Get(*pvcSpecVACName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			ctrl.eventRecorder.Eventf(pvc, v1.EventTypeWarning, util.VolumeModifyFailed, "VAC "+*pvcSpecVACName+" does not exist.")
		}
		klog.Errorf("Get VAC with vac name %s in VACInformer cache failed: %v", *pvcSpecVACName, err)
		pvc, statusErr := ctrl.markControllerModifyVolumeStatus(pvc, v1.PersistentVolumeClaimModifyVolumePending, nil)
		if statusErr != nil {
			return pvc, pv, statusErr, false
		}
		return pvc, pv, nil, false
	}

	// If we haven't attempted modification to current VAC yet,
	// set ModifyVolumeInProgress and clear any ModifyVolume conditions from outdated modifications
	if isFirstTimeModifyingVacInPvcSpec(pvc) {
		pvc, err = ctrl.markControllerModifyVolumeFirstAttempt(pvc)
		if err != nil {
			return pvc, pv, err, false
		}
	}

	// Call plugin
	ctrl.eventRecorder.Event(pvc, v1.EventTypeNormal, util.VolumeModify,
		fmt.Sprintf("external resizer is modifying volume %s with vac %s", pvc.Name, *pvcSpecVACName))
	return ctrl.controllerModifyVolumeWithTarget(pvc, pv, vac)
}

func isModificationNeeded(pvc *v1.PersistentVolumeClaim) bool {
	pvcSpecVacName := pvc.Spec.VolumeAttributesClassName
	currentVacName := pvc.Status.CurrentVolumeAttributesClassName

	return (currentVacName == nil && pvcSpecVacName != nil) ||
		(pvcSpecVacName != nil && *currentVacName != *pvcSpecVacName)
}

func isFirstTimeModifyingVacInPvcSpec(pvc *v1.PersistentVolumeClaim) bool {
	pvcSpecVacName := pvc.Spec.VolumeAttributesClassName

	return pvc.Status.ModifyVolumeStatus == nil || // Never attempted modification
		// ModifyVolumeStatus referencing outdated modification
		(pvcSpecVacName != nil && *pvcSpecVacName != pvc.Status.ModifyVolumeStatus.TargetVolumeAttributesClassName)

}

func isVacRolledBack(pvc *v1.PersistentVolumeClaim) bool {
	pvcSpecVacName := pvc.Spec.VolumeAttributesClassName
	curVacName := pvc.Status.CurrentVolumeAttributesClassName
	targetVacName := pvc.Status.ModifyVolumeStatus.TargetVolumeAttributesClassName
	// Case 1: rollback to nil
	// Case 2: rollback to previous VAC
	return (pvcSpecVacName == nil && curVacName == nil && targetVacName != "") ||
		(pvcSpecVacName != nil && curVacName != nil &&
			*pvcSpecVacName == *curVacName && targetVacName != *curVacName)
}

func currentModificationInfeasible(pvc *v1.PersistentVolumeClaim) bool {
	return pvc.Status.ModifyVolumeStatus != nil && pvc.Status.ModifyVolumeStatus.Status == v1.PersistentVolumeClaimModifyVolumeInfeasible
}

func (ctrl *modifyController) validateVACAndRollback(
	pvc *v1.PersistentVolumeClaim,
	pv *v1.PersistentVolume) (*v1.PersistentVolumeClaim, *v1.PersistentVolume, error, bool) {
	// The controller does not trigger ModifyVolume because it is only
	// for rolling back infeasible errors
	// Record an event to indicate that external resizer is rolling back this volume.
	rollbackVACName := "nil"
	if pvc.Spec.VolumeAttributesClassName != nil {
		rollbackVACName = *pvc.Spec.VolumeAttributesClassName
	}
	ctrl.eventRecorder.Event(pvc, v1.EventTypeNormal, util.VolumeModify,
		fmt.Sprintf("external resizer is rolling back volume %s with infeasible error to VAC %s", pvc.Name, rollbackVACName))
	// Mark pvc.Status.ModifyVolumeStatus as completed
	pvc, pv, err := ctrl.markControllerModifyVolumeRollbackCompleted(pvc, pv)
	if err != nil {
		return pvc, pv, fmt.Errorf("rollback volume %s modification with error: %v ", pvc.Name, err), false
	}
	return pvc, pv, nil, false
}

// func controllerModifyVolumeWithTarget trigger the CSI ControllerModifyVolume API call
// and handle both success and error scenarios
func (ctrl *modifyController) controllerModifyVolumeWithTarget(pvc *v1.PersistentVolumeClaim, pv *v1.PersistentVolume, vacObj *storagev1beta1.VolumeAttributesClass) (*v1.PersistentVolumeClaim, *v1.PersistentVolume, error, bool) {
	var pluginErr error
	pvc, pv, pluginErr = ctrl.callModifyVolumeOnPlugin(pvc, pv, vacObj)
	if pluginErr == nil {
		klog.V(4).Infof("Update volumeAttributesClass of PV %q to %s succeeded", pv.Name, *pv.Spec.VolumeAttributesClassName)
		ctrl.eventRecorder.Eventf(pvc, v1.EventTypeNormal, util.VolumeModifySuccess, fmt.Sprintf("external resizer modified volume %s with vac %s successfully ", pvc.Name, vacObj.Name))
		return pvc, pv, nil, true
	}

	// Update PVC and record an event to indicate that modify operation is failed.
	pvc, markErr := ctrl.markControllerModifyVolumeFailed(pvc, pluginErr)
	ctrl.eventRecorder.Eventf(pvc, v1.EventTypeWarning, util.VolumeModifyFailed, pluginErr.Error())
	return pvc, pv, markErr, false
}

func (ctrl *modifyController) callModifyVolumeOnPlugin(
	pvc *v1.PersistentVolumeClaim,
	pv *v1.PersistentVolume,
	vac *storagev1beta1.VolumeAttributesClass) (*v1.PersistentVolumeClaim, *v1.PersistentVolume, error) {
	if ctrl.extraModifyMetadata {
		vac.Parameters[pvcNameKey] = pvc.GetName()
		vac.Parameters[pvcNamespaceKey] = pvc.GetNamespace()
		vac.Parameters[pvNameKey] = pv.GetName()
	}
	err := ctrl.modifier.Modify(pv, vac.Parameters)

	if err != nil {
		return pvc, pv, err
	}

	pvc, pv, err = ctrl.markControllerModifyVolumeCompleted(pvc, pv)
	if err != nil {
		return pvc, pv, fmt.Errorf("modify volume failed to mark pvc %s modify volume completed: %v ", pvc.Name, err)
	}
	return pvc, pv, nil
}

// func delayModificationIfRecentlyInfeasible returns a delayRetryError if PVC modification recently failed with
// infeasible error
func (ctrl *modifyController) delayModificationIfRecentlyInfeasible(pvc *v1.PersistentVolumeClaim) error {
	pvcKey, err := cache.MetaNamespaceKeyFunc(pvc)
	if err != nil {
		return err
	}

	// Do not delay modification if PVC updated with new VAC
	s := pvc.Status.ModifyVolumeStatus
	if s == nil || pvc.Spec.VolumeAttributesClassName == nil || s.TargetVolumeAttributesClassName != *pvc.Spec.VolumeAttributesClassName {
		// remove key from slowSet because new VAC may be feasible
		ctrl.slowSet.Remove(pvcKey)
		return nil
	}

	inSlowSet := ctrl.slowSet.Contains(pvcKey)
	ctrl.markForSlowRetry(pvc, pvcKey)

	if inSlowSet {
		msg := fmt.Sprintf("skipping volume modification for pvc %s, because modification previously failed with infeasible error", pvcKey)
		klog.V(4).Infof(msg)
		delayRetryError := util.NewDelayRetryError(msg, ctrl.slowSet.TimeRemaining(pvcKey))
		return delayRetryError
	}
	return nil
}

// func markForSlowRetry adds PVC to controller's slowSet IF PVC's ModifyVolumeStatus is Infeasible
func (ctrl *modifyController) markForSlowRetry(pvc *v1.PersistentVolumeClaim, pvcKey string) {
	s := pvc.Status.ModifyVolumeStatus
	if s != nil && s.Status == v1.PersistentVolumeClaimModifyVolumeInfeasible {
		ctrl.slowSet.Add(pvcKey)
	}
}

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
	"time"

	"github.com/kubernetes-csi/csi-lib-utils/slowset"
	"github.com/kubernetes-csi/external-resizer/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	pvcSpecVacName := pvc.Spec.VolumeAttributesClassName
	curVacName := pvc.Status.CurrentVolumeAttributesClassName
	pvcKey, err := cache.MetaNamespaceKeyFunc(pvc)
	if err != nil {
		return pvc, pv, err, false
	}

	// Requeue PVC if modification recently failed with infeasible error.
	delayModificationErr := ctrl.delayModificationIfRecentlyInfeasible(pvc, pvcKey)
	if delayModificationErr != nil {
		return pvc, pv, delayModificationErr, false
	}

	if pvcSpecVacName != nil && curVacName == nil {
		// First time adding VAC to a PVC
		return ctrl.validateVACAndModifyVolumeWithTarget(pvc, pv)
	} else if pvcSpecVacName != nil && curVacName != nil && *pvcSpecVacName != *curVacName {
		// Check if PVC in uncertain state
		_, inUncertainState := ctrl.uncertainPVCs[pvcKey]
		if !inUncertainState {
			klog.V(3).InfoS("previous operation on the PVC failed with a final error, retrying")
			return ctrl.validateVACAndModifyVolumeWithTarget(pvc, pv)
		} else {
			vac, err := ctrl.vacLister.Get(*pvcSpecVacName)
			if err != nil {
				if apierrors.IsNotFound(err) {
					ctrl.eventRecorder.Eventf(pvc, v1.EventTypeWarning, util.VolumeModifyFailed, "VAC "+*pvcSpecVacName+" does not exist.")
				}
				return pvc, pv, err, false
			}
			return ctrl.controllerModifyVolumeWithTarget(pvc, pv, vac, pvcSpecVacName)
		}
	}

	// No modification required
	return pvc, pv, nil, false
}

// func validateVACAndModifyVolumeWithTarget validate the VAC. The function sets pvc.Status.ModifyVolumeStatus
// to Pending if VAC does not exist and proceeds to trigger ModifyVolume if VAC exists
func (ctrl *modifyController) validateVACAndModifyVolumeWithTarget(
	pvc *v1.PersistentVolumeClaim,
	pv *v1.PersistentVolume) (*v1.PersistentVolumeClaim, *v1.PersistentVolume, error, bool) {
	// The controller only triggers ModifyVolume if pvcSpecVacName is not nil nor empty
	pvcSpecVacName := pvc.Spec.VolumeAttributesClassName
	// Check if pvcSpecVac is valid and exist
	vac, err := ctrl.vacLister.Get(*pvcSpecVacName)
	if err == nil {
		// Mark pvc.Status.ModifyVolumeStatus as in progress
		pvc, err = ctrl.markControllerModifyVolumeStatus(pvc, v1.PersistentVolumeClaimModifyVolumeInProgress, nil)
		if err != nil {
			return pvc, pv, err, false
		}
		// Record an event to indicate that external resizer is modifying this volume.
		ctrl.eventRecorder.Event(pvc, v1.EventTypeNormal, util.VolumeModify,
			fmt.Sprintf("external resizer is modifying volume %s with vac %s", pvc.Name, *pvcSpecVacName))
		return ctrl.controllerModifyVolumeWithTarget(pvc, pv, vac, pvcSpecVacName)
	} else {
		if apierrors.IsNotFound(err) {
			ctrl.eventRecorder.Eventf(pvc, v1.EventTypeWarning, util.VolumeModifyFailed, "VAC "+*pvcSpecVacName+" does not exist.")
		}
		klog.Errorf("Get VAC with vac name %s in VACInformer cache failed: %v", *pvcSpecVacName, err)
		// Mark pvc.Status.ModifyVolumeStatus as pending
		pvc, err = ctrl.markControllerModifyVolumeStatus(pvc, v1.PersistentVolumeClaimModifyVolumePending, nil)
		return pvc, pv, err, false
	}
}

// func controllerModifyVolumeWithTarget trigger the CSI ControllerModifyVolume API call
// and handle both success and error scenarios
func (ctrl *modifyController) controllerModifyVolumeWithTarget(
	pvc *v1.PersistentVolumeClaim,
	pv *v1.PersistentVolume,
	vacObj *storagev1beta1.VolumeAttributesClass,
	pvcSpecVacName *string) (*v1.PersistentVolumeClaim, *v1.PersistentVolume, error, bool) {
	var err error
	pvc, pv, err = ctrl.callModifyVolumeOnPlugin(pvc, pv, vacObj)
	if err == nil {
		klog.V(4).Infof("Update volumeAttributesClass of PV %q to %s succeeded", pv.Name, *pvcSpecVacName)
		// Record an event to indicate that modify operation is successful.
		ctrl.eventRecorder.Eventf(pvc, v1.EventTypeNormal, util.VolumeModifySuccess, fmt.Sprintf("external resizer modified volume %s with vac %s successfully ", pvc.Name, vacObj.Name))
		return pvc, pv, nil, true
	} else {
		errStatus, ok := status.FromError(err)
		if ok {
			pvc, updateConditionErr := ctrl.updateConditionBasedOnError(pvc, err)
			if updateConditionErr != nil {
				return nil, nil, err, false
			}
			pvcKey, keyErr := cache.MetaNamespaceKeyFunc(pvc)
			if keyErr != nil {
				return pvc, pv, keyErr, false
			}
			if !util.IsFinalError(keyErr) {
				// update conditions and cache pvc as uncertain
				ctrl.uncertainPVCs[pvcKey] = *pvc
			} else {
				// Only InvalidArgument can be set to Infeasible state
				// Final errors other than InvalidArgument will still be in InProgress state
				if errStatus.Code() == codes.InvalidArgument {
					// Mark pvc.Status.ModifyVolumeStatus as infeasible
					pvc, markModifyVolumeInfeasibleError := ctrl.markControllerModifyVolumeStatus(pvc, v1.PersistentVolumeClaimModifyVolumeInfeasible, err)
					if markModifyVolumeInfeasibleError != nil {
						return pvc, pv, markModifyVolumeInfeasibleError, false
					}
					ctrl.markForSlowRetry(pvc, pvcKey)
				}
				ctrl.removePVCFromModifyVolumeUncertainCache(pvcKey)
			}
		} else {
			return pvc, pv, fmt.Errorf("cannot get error status from modify volume err: %v ", err), false
		}
		// Record an event to indicate that modify operation is failed.
		ctrl.eventRecorder.Eventf(pvc, v1.EventTypeWarning, util.VolumeModifyFailed, err.Error())
		return pvc, pv, err, false
	}
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
func (ctrl *modifyController) delayModificationIfRecentlyInfeasible(pvc *v1.PersistentVolumeClaim, pvcKey string) error {
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
		klog.V(4).Info(msg)
		delayRetryError := util.NewDelayRetryError(msg, ctrl.slowSet.TimeRemaining(pvcKey))
		return delayRetryError
	}
	return nil
}

// func markForSlowRetry adds PVC to controller's slowSet IF PVC's ModifyVolumeStatus is Infeasible
func (ctrl *modifyController) markForSlowRetry(pvc *v1.PersistentVolumeClaim, pvcKey string) {
	s := pvc.Status.ModifyVolumeStatus
	if s != nil && s.Status == v1.PersistentVolumeClaimModifyVolumeInfeasible {
		ctrl.slowSet.Add(pvcKey, slowset.ObjectData{
			Timestamp: time.Now(),
		})
	}
}

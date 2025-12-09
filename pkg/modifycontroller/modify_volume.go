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
	"context"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/kubernetes-csi/csi-lib-utils/slowset"
	"github.com/kubernetes-csi/external-resizer/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

const (
	pvcNameKey      = "csi.storage.k8s.io/pvc/name"
	pvcNamespaceKey = "csi.storage.k8s.io/pvc/namespace"
	pvNameKey       = "csi.storage.k8s.io/pv/name"
)

// The return value bool is only used as a sentinel value when function returns without actually performing modification
func (ctrl *modifyController) modify(pvc *v1.PersistentVolumeClaim, pv *v1.PersistentVolume) (*v1.PersistentVolumeClaim, *v1.PersistentVolume, error, bool) {
	pvcKey, err := cache.MetaNamespaceKeyFunc(pvc)
	if err != nil {
		return pvc, pv, err, false
	}

	// Requeue PVC if modification recently failed with infeasible error.
	delayModificationErr := ctrl.delayModificationIfRecentlyInfeasible(pvc, pvcKey)
	if delayModificationErr != nil {
		return pvc, pv, delayModificationErr, false
	}

	pvcSpecVacName := ptr.Deref(pvc.Spec.VolumeAttributesClassName, "")
	curVacName := ptr.Deref(pvc.Status.CurrentVolumeAttributesClassName, "")
	status := pvc.Status.ModifyVolumeStatus

	if status == nil && pvcSpecVacName == curVacName {
		// No modification required, already reached target state
		return pvc, pv, nil, false
	}

	// Last modification failed with non-infeasible error
	inProgress := status != nil && status.Status == v1.PersistentVolumeClaimModifyVolumeInProgress

	if pvcSpecVacName == "" && !inProgress {
		// User don't care the target state, and we've reached a relatively stable state. Just keep it here.
		// Note: APIServer generally not allowing setting pvcSpecVacName to empty when curVacName is not empty.
		klog.V(4).InfoS("stop reconcile for rolled back PVC", "PV", klog.KObj(pv))
		pvc, err := ctrl.rolledBack(pvc)
		return pvc, pv, err, false
	}

	ctx := context.TODO()
	inUncertainState := false
	if inProgress {
		_, inUncertainState = ctrl.uncertainPVCs.Load(pvcKey)
	} else {
		// we either see a stall PVC, or the status was updated externally.
		// For stall PVC, we will get Conflict error when marking InProgress.
		// For status updated externally, we respect the user's choice and try the new target, as if it were not uncertain.
		// `ctrl.uncertainPVCs` will be updated after the next ControllerModifyVolume call.
	}
	// Check if we should change our target
	if inUncertainState || pvcSpecVacName == "" {
		// No. Continue our previous modification
		vac, err := ctrl.getTargetVAC(pvc, status.TargetVolumeAttributesClassName)
		if err != nil {
			return pvc, pv, err, false
		}
		return ctrl.controllerModifyVolumeWithTarget(ctx, pvc, pv, vac)
	}

	return ctrl.validateVACAndModifyVolumeWithTarget(ctx, pvc, pv)
}

func (ctrl *modifyController) rolledBack(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	if slices.ContainsFunc(pvc.Status.Conditions, func(condition v1.PersistentVolumeClaimCondition) bool {
		return condition.Type == v1.PersistentVolumeClaimVolumeModifyingVolume
	}) {
		ctrl.eventRecorder.Eventf(pvc, v1.EventTypeNormal, util.VolumeModifyCancelled, "Cancelled modify.")
		return ctrl.markRolledBack(pvc)
	}
	// Don't try to revert Status.ModifyVolumeStatus here, because we only record the result of the last modification.
	// We don't know what happened before. User can switch between InProgress/Infeasible/Pending status
	// freely by modifying the spec.
	return pvc, nil
}

func (ctrl *modifyController) getTargetVAC(pvc *v1.PersistentVolumeClaim, vacName string) (*storagev1.VolumeAttributesClass, error) {
	vac, err := ctrl.vacLister.Get(vacName)
	// Check if pvcSpecVac is valid and exist
	if err != nil {
		if apierrors.IsNotFound(err) {
			ctrl.eventRecorder.Eventf(pvc, v1.EventTypeWarning, util.VolumeModifyFailed, "VAC %q does not exist.", vacName)
		}
		return nil, fmt.Errorf("get VAC with vac name %s in VACInformer cache failed: %w", vacName, err)
	}
	return vac, nil
}

// func validateVACAndModifyVolumeWithTarget validate the VAC. The function sets pvc.Status.ModifyVolumeStatus
// to Pending if VAC does not exist and proceeds to trigger ModifyVolume if VAC exists
func (ctrl *modifyController) validateVACAndModifyVolumeWithTarget(
	ctx context.Context,
	pvc *v1.PersistentVolumeClaim,
	pv *v1.PersistentVolume) (*v1.PersistentVolumeClaim, *v1.PersistentVolume, error, bool) {

	vac, err := ctrl.getTargetVAC(pvc, *pvc.Spec.VolumeAttributesClassName)
	if err != nil {
		// Mark pvc.Status.ModifyVolumeStatus as pending
		pvc, err = ctrl.markControllerModifyVolumeStatus(pvc, v1.PersistentVolumeClaimModifyVolumePending, nil)
		return pvc, pv, err, false
	}

	// Mark pvc.Status.ModifyVolumeStatus as in progress
	pvc, err = ctrl.markControllerModifyVolumeStatus(pvc, v1.PersistentVolumeClaimModifyVolumeInProgress, nil)
	if err != nil {
		return pvc, pv, err, false
	}
	// Record an event to indicate that external resizer is modifying this volume.
	ctrl.eventRecorder.Event(pvc, v1.EventTypeNormal, util.VolumeModify,
		fmt.Sprintf("external resizer is modifying volume %s with vac %s", pvc.Name, vac.Name))
	return ctrl.controllerModifyVolumeWithTarget(ctx, pvc, pv, vac)
}

// func controllerModifyVolumeWithTarget trigger the CSI ControllerModifyVolume API call
// and handle both success and error scenarios
func (ctrl *modifyController) controllerModifyVolumeWithTarget(
	ctx context.Context,
	pvc *v1.PersistentVolumeClaim,
	pv *v1.PersistentVolume,
	vacObj *storagev1.VolumeAttributesClass,
) (*v1.PersistentVolumeClaim, *v1.PersistentVolume, error, bool) {
	var err error
	pvc, pv, err = ctrl.callModifyVolumeOnPlugin(ctx, pvc, pv, vacObj)
	if err == nil {
		klog.V(4).Infof("Update volumeAttributesClass of PV %q to %s succeeded", pv.Name, vacObj.Name)
		// Record an event to indicate that modify operation is successful.
		ctrl.eventRecorder.Eventf(pvc, v1.EventTypeNormal, util.VolumeModifySuccess, fmt.Sprintf("external resizer modified volume %s with vac %s successfully", pvc.Name, vacObj.Name))
		return pvc, pv, nil, true
	} else {
		errStatus, ok := status.FromError(err)
		errMsg := err.Error()
		if ok {
			targetStatus := v1.PersistentVolumeClaimModifyVolumeInProgress
			pvcKey, keyErr := cache.MetaNamespaceKeyFunc(pvc)
			if keyErr != nil {
				return pvc, pv, keyErr, false
			}
			if !util.IsFinalError(err) {
				// update conditions and cache pvc as uncertain
				ctrl.uncertainPVCs.Store(pvcKey, pvc)
				errMsg += ". Still modifying to VAC " + vacObj.Name
			} else {
				// Only InvalidArgument can be set to Infeasible state
				// Final errors other than InvalidArgument will still be in InProgress state
				if errStatus.Code() == codes.InvalidArgument {
					targetStatus = v1.PersistentVolumeClaimModifyVolumeInfeasible
				}
				ctrl.uncertainPVCs.Delete(pvcKey)
			}
			var markErr error
			pvc, markErr = ctrl.markControllerModifyVolumeStatus(pvc, targetStatus, err)
			if markErr != nil {
				return pvc, pv, markErr, false
			}
			ctrl.markForSlowRetry(pvc, pvcKey)
		} else {
			return pvc, pv, fmt.Errorf("cannot get error status from modify volume err: %v", err), false
		}
		// Record an event to indicate that modify operation is failed.
		ctrl.eventRecorder.Event(pvc, v1.EventTypeWarning, util.VolumeModifyFailed, errMsg)
		return pvc, pv, err, false
	}
}

func (ctrl *modifyController) callModifyVolumeOnPlugin(
	ctx context.Context,
	pvc *v1.PersistentVolumeClaim,
	pv *v1.PersistentVolume,
	vac *storagev1.VolumeAttributesClass) (*v1.PersistentVolumeClaim, *v1.PersistentVolume, error) {
	parameters := vac.Parameters
	if ctrl.extraModifyMetadata {
		if len(parameters) == 0 {
			parameters = make(map[string]string, 3)
		} else {
			parameters = maps.Clone(parameters)
		}
		parameters[pvcNameKey] = pvc.GetName()
		parameters[pvcNamespaceKey] = pvc.GetNamespace()
		parameters[pvNameKey] = pv.GetName()
	}
	err := ctrl.modifier.Modify(ctx, pv, parameters)

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

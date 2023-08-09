/*
Copyright 2022 The Kubernetes Authors.

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
	"fmt"

	"github.com/kubernetes-csi/external-resizer/pkg/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// markControllerResizeInProgress will mark PVC for controller resize, this function is newer version that uses
// resizeStatus and sets allocatedResources.
func (ctrl *resizeController) markControllerResizeInProgress(
	pvc *v1.PersistentVolumeClaim, newSize resource.Quantity) (*v1.PersistentVolumeClaim, error) {

	progressCondition := v1.PersistentVolumeClaimCondition{
		Type:               v1.PersistentVolumeClaimResizing,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
	}
	controllerResizeInProgress := v1.PersistentVolumeClaimControllerResizeInProgress
	conditions := []v1.PersistentVolumeClaimCondition{progressCondition}

	newPVC := pvc.DeepCopy()
	newPVC.Status.Conditions = util.MergeResizeConditionsOfPVC(newPVC.Status.Conditions, conditions)
	newPVC.Status.AllocatedResourceStatuses[v1.ResourceStorage] = controllerResizeInProgress
	newPVC.Status.AllocatedResources = v1.ResourceList{v1.ResourceStorage: newSize}
	updatedPVC, err := ctrl.patchClaim(pvc, newPVC, true /* addResourceVersionCheck */)
	if err != nil {
		return pvc, err
	}
	return updatedPVC, nil
}

// markForPendingNodeExpansion is new set of functions designed around feature RecoverVolumeExpansionFailure
// which correctly sets pvc.Status.ResizeStatus
func (ctrl *resizeController) markForPendingNodeExpansion(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	pvcCondition := v1.PersistentVolumeClaimCondition{
		Type:               v1.PersistentVolumeClaimFileSystemResizePending,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            "Waiting for user to (re-)start a pod to finish file system resize of volume on node.",
	}

	nodeResizePendingVar := v1.PersistentVolumeClaimNodeResizePending
	newPVC := pvc.DeepCopy()
	newPVC.Status.Conditions = util.MergeResizeConditionsOfPVC(newPVC.Status.Conditions,
		[]v1.PersistentVolumeClaimCondition{pvcCondition})
	newPVC.Status.AllocatedResourceStatuses[v1.ResourceStorage] = nodeResizePendingVar
	updatedPVC, err := ctrl.patchClaim(pvc, newPVC, true /* addResourceVersionCheck */)

	if err != nil {
		return updatedPVC, fmt.Errorf("Mark PVC %q as node resize required failed: %v", util.PVCKey(pvc), err)
	}

	klog.V(4).Infof("Mark PVC %q as file system resize required", util.PVCKey(pvc))
	ctrl.eventRecorder.Eventf(pvc, v1.EventTypeNormal,
		util.FileSystemResizeRequired, "Require file system resize of volume on node")

	return updatedPVC, nil
}

func (ctrl *resizeController) markControllerExpansionFailed(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	resizeFailedOnController := v1.PersistentVolumeClaimControllerResizeFailed
	newPVC := pvc.DeepCopy()
	newPVC.Status.AllocatedResourceStatuses[v1.ResourceStorage] = resizeFailedOnController

	// We are setting addResourceVersionCheck as false as an optimization
	// because if resize fails on controller and somehow we can't update PVC
	// because our version of object is slightly older then the entire resize
	// operation must be restarted before ResizeStatus can be set to Expansionfailedoncontroller.
	// Setting addResourceVersionCheck to `false` ensures that we set `ResizeStatus`
	// even if our version of PVC was slightly older.
	updatedPVC, err := ctrl.patchClaim(pvc, newPVC, false /* addResourceVersionCheck */)
	if err != nil {
		return pvc, fmt.Errorf("Mark PVC %q as controller resize failed, errored with: %v", util.PVCKey(pvc), err)
	}
	return updatedPVC, nil
}

func (ctrl *resizeController) markOverallExpansionAsFinished(
	pvc *v1.PersistentVolumeClaim,
	newSize resource.Quantity) (*v1.PersistentVolumeClaim, error) {
	newPVC := pvc.DeepCopy()
	newPVC.Status.Capacity[v1.ResourceStorage] = newSize
	newPVC.Status.Conditions = util.MergeResizeConditionsOfPVC(pvc.Status.Conditions, []v1.PersistentVolumeClaimCondition{})
	var resizeFinished v1.ClaimResourceStatus
	newPVC.Status.AllocatedResourceStatuses[v1.ResourceStorage] = resizeFinished

	updatedPVC, err := ctrl.patchClaim(pvc, newPVC, true /* addResourceVersionCheck */)
	if err != nil {
		return pvc, fmt.Errorf("Mark PVC %q as resize finished failed: %v", util.PVCKey(pvc), err)
	}

	klog.V(4).Infof("Resize PVC %q finished", util.PVCKey(pvc))
	ctrl.eventRecorder.Eventf(pvc, v1.EventTypeNormal, util.VolumeResizeSuccess, "Resize volume succeeded")

	return updatedPVC, nil
}

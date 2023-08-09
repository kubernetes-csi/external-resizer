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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
)

func (ctrl *resizeController) expandAndRecover(pvc *v1.PersistentVolumeClaim, pv *v1.PersistentVolume) (*v1.PersistentVolumeClaim, *v1.PersistentVolume, error, bool) {
	if !ctrl.pvCanBeExpanded(pv, pvc) {
		klog.V(4).Infof("No need to resize PV %q", pv.Name)
		return pvc, pv, nil, false
	}
	// only used as a sentinel value when function returns without
	// actually performing resize on the volume.
	resizeNotCalled := false

	// if we are here that already means pvc.Spec.Size > pvc.Status.Size
	pvcSpecSize := pvc.Spec.Resources.Requests[v1.ResourceStorage]
	pvcStatusSize := pvc.Status.Capacity[v1.ResourceStorage]
	pvSize := pv.Spec.Capacity[v1.ResourceStorage]

	newSize := pvcSpecSize
	var resizeStatus v1.ClaimResourceStatus
	if status, ok := pvc.Status.AllocatedResourceStatuses[v1.ResourceStorage]; ok {
		resizeStatus = status
	}

	var allocatedSize *resource.Quantity
	t, ok := pvc.Status.AllocatedResources[v1.ResourceStorage]
	if ok {
		allocatedSize = &t
	}

	if pvSize.Cmp(pvcSpecSize) < 0 {
		// PV is smaller than user requested size. In general some control-plane volume resize
		// is necessary at this point but if we were expanding the PVC before we should let
		// previous operation finish before starting resize to new user requested size.
		switch resizeStatus {
		case v1.PersistentVolumeClaimControllerResizeInProgress,
			v1.PersistentVolumeClaimNodeResizeFailed:
			if allocatedSize != nil {
				newSize = *allocatedSize
			}
		case v1.PersistentVolumeClaimNodeResizePending,
			v1.PersistentVolumeClaimNodeResizeInProgress:
			if allocatedSize != nil {
				newSize = *allocatedSize
			}
			// If PV is already greater or equal to whatever newSize is, then we should
			// let previously issued node resize finish before starting new one.
			// This case is essentially an optimization on previous case, generally it is safe
			// to let volume plugin receive the resize call (resize calls are idempotent)
			// but having this check here *avoids* unnecessary RPC calls to plugin.
			if pvSize.Cmp(newSize) >= 0 {
				return pvc, pv, nil, resizeNotCalled
			}
		default:
			newSize = pvcSpecSize
		}
	} else {
		// PV has already been expanded and hence we can be here for following reasons:
		//   1. If resize is pending on the node and this was just a spurious update event
		//      we don't need to do anything and let kubelet handle it.
		//   2. It could be that - although we successfully expanded the volume, we failed to
		//      record our work in API objects, in which case - we should resume resizing operation
		//      and let API objects be updated.
		//   3. Controller successfully expanded the volume, but resize is failing on the node
		//      and before kubelet can retry failed node resize - controller must verify if it is
		//      safe to do so.
		//   4. While resize was still pending on the node, user reduced the pvc size.
		switch resizeStatus {
		case v1.PersistentVolumeClaimNodeResizeInProgress,
			v1.PersistentVolumeClaimNodeResizePending:
			// we don't need to do any work. We could be here because of a spurious update event.
			// This is case #1
			return pvc, pv, nil, resizeNotCalled
		case v1.PersistentVolumeClaimNodeResizeFailed:
			// This is case#3, we need to reset the pvc status in such a way that kubelet can safely retry volume
			// resize.
			if ctrl.resizer.DriverSupportsControlPlaneExpansion() && allocatedSize != nil {
				newSize = *allocatedSize
			} else {
				newSize = pvcSpecSize
			}
		case v1.PersistentVolumeClaimControllerResizeInProgress,
			v1.PersistentVolumeClaimControllerResizeFailed:
			// This is case#2 or it could also be case#4 when user manually shrunk the PVC
			// after expanding it.
			if allocatedSize != nil {
				newSize = *allocatedSize
			}
		default:
			// It is impossible for ResizeStatus to be "" and allocatedSize to be not nil but somehow
			// if we do end up in this state, it is safest to resume resize to last recorded size in
			// allocatedSize variable.
			if pvc.Status.AllocatedResourceStatuses[v1.ResourceStorage] == "" && allocatedSize != nil {
				newSize = *allocatedSize
			} else {
				newSize = pvcSpecSize
			}
		}
	}
	var err error
	pvc, err = ctrl.markControllerResizeInProgress(pvc, newSize)
	if err != nil {
		return pvc, pv, fmt.Errorf("marking pvc %q as resizing failed: %v", util.PVCKey(pvc), err), resizeNotCalled
	}

	// if pvc previously failed to expand because it can't be expanded when in-use
	// we must not try resize here
	if ctrl.usedPVCs.hasInUseErrors(pvc) && ctrl.usedPVCs.checkForUse(pvc) {
		// Record an event to indicate that resizer is not expanding the pvc
		msg := fmt.Sprintf("Unable to expand %s because CSI driver %s only supports offline resize and volume is currently in-use", util.PVCKey(pvc), ctrl.resizer.Name())
		ctrl.eventRecorder.Event(pvc, v1.EventTypeWarning, util.VolumeResizeFailed, msg)
		return pvc, pv, fmt.Errorf(msg), resizeNotCalled
	}

	// Record an event to indicate that external resizer is resizing this volume.
	ctrl.eventRecorder.Event(pvc, v1.EventTypeNormal, util.VolumeResizing,
		fmt.Sprintf("External resizer is resizing volume %s", pv.Name))

	// before trying resize we will remove the PVC from map
	// that tracks PVCs which can't be expanded when in-use. If
	// pvc indeed can not be expanded when in-use then it will be added
	// back when resize fails with in-use error.
	ctrl.usedPVCs.removePVCWithInUseError(pvc)
	pvc, pv, err = ctrl.callResizeOnPlugin(pvc, pv, newSize, pvcStatusSize)

	if err != nil {
		// Record an event to indicate that resize operation is failed.
		ctrl.eventRecorder.Eventf(pvc, v1.EventTypeWarning, util.VolumeResizeFailed, err.Error())
		return pvc, pv, err, true
	}

	klog.V(4).Infof("Update capacity of PV %q to %s succeeded", pv.Name, newSize.String())
	return pvc, pv, nil, true
}

func (ctrl *resizeController) callResizeOnPlugin(
	pvc *v1.PersistentVolumeClaim,
	pv *v1.PersistentVolume,
	newSize, oldSize resource.Quantity) (*v1.PersistentVolumeClaim, *v1.PersistentVolume, error) {
	updatedSize, fsResizeRequired, err := ctrl.resizer.Resize(pv, newSize)

	if err != nil {
		// if this error was a in-use error then it must be tracked so as we don't retry without
		// first verifying if volume is in-use
		if inUseError(err) {
			ctrl.usedPVCs.addPVCWithInUseError(pvc)
		}
		if isFinalError(err) {
			var markExpansionFailedError error
			pvc, markExpansionFailedError = ctrl.markControllerExpansionFailed(pvc)
			if markExpansionFailedError != nil {
				return pvc, pv, fmt.Errorf("resizing failed in controller with %v but failed to update PVC %s with: %v", err, util.PVCKey(pvc), markExpansionFailedError)
			}
		}
		return pvc, pv, fmt.Errorf("resize volume %q by resizer %q failed: %v", pv.Name, ctrl.name, err)
	}

	klog.V(4).Infof("Resize volume succeeded for volume %q, start to update PV's capacity", pv.Name)

	pv, err = ctrl.updatePVCapacity(pv, oldSize, updatedSize, fsResizeRequired)
	if err != nil {
		return pvc, pv, fmt.Errorf("error updating pv %q by resizer: %v", pv.Name, err)
	}

	if fsResizeRequired {
		pvc, err = ctrl.markForPendingNodeExpansion(pvc)
		return pvc, pv, err
	}
	pvc, err = ctrl.markOverallExpansionAsFinished(pvc, updatedSize)
	return pvc, pv, err
}

// checks if pv can be expanded
func (ctrl *resizeController) pvCanBeExpanded(pv *v1.PersistentVolume, pvc *v1.PersistentVolumeClaim) bool {
	if !ctrl.resizer.CanSupport(pv, pvc) {
		klog.V(4).Infof("Resizer %q doesn't support PV %q", ctrl.name, pv.Name)
		return false
	}

	if (pv.Spec.ClaimRef == nil) || (pvc.Namespace != pv.Spec.ClaimRef.Namespace) || (pvc.UID != pv.Spec.ClaimRef.UID) {
		klog.V(4).Infof("persistent volume is not bound to PVC being updated: %s", util.PVCKey(pvc))
		return false
	}
	return true
}

func isFinalError(err error) bool {
	// Sources:
	// https://github.com/grpc/grpc/blob/master/doc/statuscodes.md
	// https://github.com/container-storage-interface/spec/blob/master/spec.md
	st, ok := status.FromError(err)
	if !ok {
		// This is not gRPC error. The operation must have failed before gRPC
		// method was called, otherwise we would get gRPC error.
		// We don't know if any previous volume operation is in progress, be on the safe side.
		return false
	}
	switch st.Code() {
	case codes.Canceled, // gRPC: Client Application cancelled the request
		codes.DeadlineExceeded,  // gRPC: Timeout
		codes.Unavailable,       // gRPC: Server shutting down, TCP connection broken - previous volume operation may be still in progress.
		codes.ResourceExhausted, // gRPC: Server temporarily out of resources - previous volume operation may be still in progress.
		codes.Aborted:           // CSI: Operation pending for volume
		return false
	}
	// All other errors mean that operation either did not
	// even start or failed. It is for sure not in progress.
	return true
}

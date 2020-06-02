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

package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/kubernetes-csi/external-resizer/pkg/resizer"
	"github.com/kubernetes-csi/external-resizer/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

// ResizeController watches PVCs and checks if they are requesting an resizing operation.
// If requested, it will resize according PVs and update PVCs' status to reflect the new size.
type ResizeController interface {
	// Run starts the controller.
	Run(workers int, ctx context.Context)
}

type resizeController struct {
	name          string
	resizer       resizer.Resizer
	kubeClient    kubernetes.Interface
	claimQueue    workqueue.RateLimitingInterface
	eventRecorder record.EventRecorder
	pvLister      corelisters.PersistentVolumeLister
	pvSynced      cache.InformerSynced
	pvcLister     corelisters.PersistentVolumeClaimLister
	pvcSynced     cache.InformerSynced

	usedPVCs *inUsePVCStore

	podLister       corelisters.PodLister
	podListerSynced cache.InformerSynced
}

// NewResizeController returns a ResizeController.
func NewResizeController(
	name string,
	resizer resizer.Resizer,
	kubeClient kubernetes.Interface,
	resyncPeriod time.Duration,
	informerFactory informers.SharedInformerFactory,
	pvcRateLimiter workqueue.RateLimiter) ResizeController {
	pvInformer := informerFactory.Core().V1().PersistentVolumes()
	pvcInformer := informerFactory.Core().V1().PersistentVolumeClaims()

	// list pods so as we can identify PVC that are in-use
	podInformer := informerFactory.Core().V1().Pods()

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events(v1.NamespaceAll)})
	eventRecorder := eventBroadcaster.NewRecorder(scheme.Scheme,
		v1.EventSource{Component: fmt.Sprintf("external-resizer %s", name)})

	claimQueue := workqueue.NewNamedRateLimitingQueue(
		pvcRateLimiter, fmt.Sprintf("%s-pvc", name))

	ctrl := &resizeController{
		name:            name,
		resizer:         resizer,
		kubeClient:      kubeClient,
		pvLister:        pvInformer.Lister(),
		pvSynced:        pvInformer.Informer().HasSynced,
		pvcLister:       pvcInformer.Lister(),
		pvcSynced:       pvcInformer.Informer().HasSynced,
		podLister:       podInformer.Lister(),
		podListerSynced: podInformer.Informer().HasSynced,
		claimQueue:      claimQueue,
		eventRecorder:   eventRecorder,
		usedPVCs:        newUsedPVCStore(),
	}

	// Add a resync period as the PVC's request size can be resized again when we handling
	// a previous resizing request of the same PVC.
	pvcInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    ctrl.addPVC,
		UpdateFunc: ctrl.updatePVC,
		DeleteFunc: ctrl.deletePVC,
	}, resyncPeriod)

	podInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    ctrl.addPod,
		DeleteFunc: ctrl.deletePod,
	}, resyncPeriod)

	return ctrl
}

func (ctrl *resizeController) addPVC(obj interface{}) {
	objKey, err := getObjectKey(obj)
	if err != nil {
		return
	}
	ctrl.claimQueue.Add(objKey)
}

func (ctrl *resizeController) addPod(obj interface{}) {
	pod := parsePod(obj)
	if pod != nil {
		return
	}
	ctrl.usedPVCs.addPod(pod)
}

func (ctrl *resizeController) deletePod(obj interface{}) {
	pod := parsePod(obj)
	if pod != nil {
		return
	}
	ctrl.usedPVCs.removePod(pod)
}

func (ctrl *resizeController) updatePVC(oldObj, newObj interface{}) {
	oldPVC, ok := oldObj.(*v1.PersistentVolumeClaim)
	if !ok || oldPVC == nil {
		return
	}

	newPVC, ok := newObj.(*v1.PersistentVolumeClaim)
	if !ok || newPVC == nil {
		return
	}

	newSize := newPVC.Spec.Resources.Requests[v1.ResourceStorage]
	oldSize := oldPVC.Spec.Resources.Requests[v1.ResourceStorage]

	newResizerName := newPVC.Annotations[util.VolumeResizerKey]
	oldResizerName := oldPVC.Annotations[util.VolumeResizerKey]

	// We perform additional checks to avoid double processing of PVCs, as we will also receive Update event when:
	// 1. Administrator or users may introduce other changes(such as add labels, modify annotations, etc.)
	//    unrelated to volume resize.
	// 2. Informer will resync and send Update event periodically without any changes.
	//
	// We add the PVC into work queue when the new size is larger then the old size
	// or when the resizer name changes. This is needed for CSI migration for the follow two cases:
	//
	// 1. First time a migrated PVC is expanded:
	// It does not yet have the annotation because annotation is only added by in-tree resizer when it receives a volume
	// expansion request. So first update event that will be received by external-resizer will be ignored because it won't
	// know how to support resizing of a "un-annotated" in-tree PVC. When in-tree resizer does add the annotation, a second
	// update even will be received and we add the pvc to workqueue. If annotation matches the registered driver name in
	// csi_resizer object, we proceeds with expansion internally or we discard the PVC.
	// 2. An already expanded in-tree PVC:
	// An in-tree PVC is resized with in-tree resizer. And later, CSI migration is turned on and resizer name is updated from
	// in-tree resizer name to CSI driver name.
	if newSize.Cmp(oldSize) > 0 || newResizerName != oldResizerName {
		ctrl.addPVC(newObj)
	} else {
		// PVC's size not changed, so this Update event maybe caused by:
		//
		// 1. Administrators or users introduce other changes(such as add labels, modify annotations, etc.)
		//    unrelated to volume resize.
		// 2. Informer resynced the PVC and send this Update event without any changes.
		//
		// If it is case 1, we can just discard this event. If case 2, we need to put it into the queue to
		// perform a resync operation.
		if newPVC.ResourceVersion == oldPVC.ResourceVersion {
			// This is case 2.
			ctrl.addPVC(newObj)
		}
	}
}

func (ctrl *resizeController) deletePVC(obj interface{}) {
	objKey, err := getObjectKey(obj)
	if err != nil {
		return
	}
	ctrl.claimQueue.Forget(objKey)
}

func getObjectKey(obj interface{}) (string, error) {
	if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
		obj = unknown.Obj
	}
	objKey, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Errorf("Failed to get key from object: %v", err)
		return "", err
	}
	return objKey, nil
}

// Run starts the controller.
func (ctrl *resizeController) Run(
	workers int, ctx context.Context) {
	defer ctrl.claimQueue.ShutDown()

	klog.Infof("Starting external resizer %s", ctrl.name)
	defer klog.Infof("Shutting down external resizer %s", ctrl.name)

	stopCh := ctx.Done()

	if !cache.WaitForCacheSync(stopCh, ctrl.pvSynced, ctrl.pvcSynced, ctrl.podListerSynced) {
		klog.Errorf("Cannot sync pod, pv or pvc caches")
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(ctrl.syncPVCs, 0, stopCh)
	}

	<-stopCh
}

// syncPVCs is the main worker.
func (ctrl *resizeController) syncPVCs() {
	key, quit := ctrl.claimQueue.Get()
	if quit {
		return
	}
	defer ctrl.claimQueue.Done(key)

	if err := ctrl.syncPVC(key.(string)); err != nil {
		// Put PVC back to the queue so that we can retry later.
		ctrl.claimQueue.AddRateLimited(key)
	} else {
		ctrl.claimQueue.Forget(key)
	}
}

// syncPVC checks if a pvc requests resizing, and execute the resize operation if requested.
func (ctrl *resizeController) syncPVC(key string) error {
	klog.V(4).Infof("Started PVC processing %q", key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Errorf("Split meta namespace key of pvc %s failed: %v", key, err)
		return err
	}

	pvc, err := ctrl.pvcLister.PersistentVolumeClaims(namespace).Get(name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.V(3).Infof("PVC %s/%s is deleted, no need to process it", namespace, name)
			return nil
		}
		klog.Errorf("Get PVC %s/%s failed: %v", namespace, name, err)
		return err
	}

	if !ctrl.pvcNeedResize(pvc) {
		klog.V(4).Infof("No need to resize PVC %q", util.PVCKey(pvc))
		return nil
	}

	pv, err := ctrl.pvLister.Get(pvc.Spec.VolumeName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.V(3).Infof("PV %s is deleted, no need to process it", pvc.Spec.VolumeName)
			return nil
		}
		klog.Errorf("Get PV %q of pvc %q failed: %v", pvc.Spec.VolumeName, util.PVCKey(pvc), err)
		return err
	}

	if !ctrl.pvNeedResize(pvc, pv) {
		klog.V(4).Infof("No need to resize PV %q", pv.Name)
		return nil
	}

	return ctrl.resizePVC(pvc, pv)
}

// pvcNeedResize returns true is a pvc requests a resize operation.
func (ctrl *resizeController) pvcNeedResize(pvc *v1.PersistentVolumeClaim) bool {
	// Only Bound pvc can be expanded.
	if pvc.Status.Phase != v1.ClaimBound {
		return false
	}
	if pvc.Spec.VolumeName == "" {
		return false
	}
	actualSize := pvc.Status.Capacity[v1.ResourceStorage]
	requestSize := pvc.Spec.Resources.Requests[v1.ResourceStorage]
	return requestSize.Cmp(actualSize) > 0
}

// pvNeedResize returns true if a pv supports and also requests resize.
func (ctrl *resizeController) pvNeedResize(pvc *v1.PersistentVolumeClaim, pv *v1.PersistentVolume) bool {
	if !ctrl.resizer.CanSupport(pv, pvc) {
		klog.V(4).Infof("Resizer %q doesn't support PV %q", ctrl.name, pv.Name)
		return false
	}

	if (pv.Spec.ClaimRef == nil) || (pvc.Namespace != pv.Spec.ClaimRef.Namespace) || (pvc.UID != pv.Spec.ClaimRef.UID) {
		klog.V(4).Infof("persistent volume is not bound to PVC being updated: %s", util.PVCKey(pvc))
		return false
	}

	pvSize := pv.Spec.Capacity[v1.ResourceStorage]
	requestSize := pvc.Spec.Resources.Requests[v1.ResourceStorage]
	if pvSize.Cmp(requestSize) >= 0 {
		// If PV size is equal or bigger than request size, that means we have already resized PV.
		// In this case we need to check PVC's condition.
		// 1. If PVC in PersistentVolumeClaimResizing condition, we should continue to perform the
		//    resizing operation as we need to know if file system resize if required. (What's more,
		//    we hope the driver can find that the actual size already matched the request size and do nothing).
		// 2. If PVC in PersistentVolumeClaimFileSystemResizePending condition, we need to
		//    do nothing as kubelet will finish file system resizing and mark resize as finished.
		if util.HasFileSystemResizePendingCondition(pvc) {
			// This is case 2.
			return false
		}
		// This is case 1.
		return true
	}

	// PV size is smaller than request size, we need to resize the volume.
	return true
}

// resizePVC will:
// 1. Mark pvc as resizing.
// 2. Resize the volume and the pv object.
// 3. Mark pvc as resizing finished(no error, no need to resize fs), need resizing fs or resize failed.
func (ctrl *resizeController) resizePVC(pvc *v1.PersistentVolumeClaim, pv *v1.PersistentVolume) error {
	if updatedPVC, err := ctrl.markPVCResizeInProgress(pvc); err != nil {
		klog.Errorf("Mark pvc %q as resizing failed: %v", util.PVCKey(pvc), err)
		return err
	} else if updatedPVC != nil {
		pvc = updatedPVC
	}

	// if pvc previously failed to expand because it can't be expanded when in-use
	// we must not try expansion here
	if ctrl.usedPVCs.hasInUseErrors(pvc) && ctrl.usedPVCs.checkForUse(pvc) {
		// Record an event to indicate that resizer is not expanding the pvc
		ctrl.eventRecorder.Event(pvc, v1.EventTypeWarning, util.VolumeResizeFailed,
			fmt.Sprintf("CSI resizer is not expanding %s because it is in-use", pv.Name))
		return fmt.Errorf("csi resizer is not expanding %s because it is in-use", pv.Name)
	}

	// Record an event to indicate that external resizer is resizing this volume.
	ctrl.eventRecorder.Event(pvc, v1.EventTypeNormal, util.VolumeResizing,
		fmt.Sprintf("External resizer is resizing volume %s", pv.Name))

	err := func() error {
		newSize, fsResizeRequired, err := ctrl.resizeVolume(pvc, pv)
		if err != nil {
			return err
		}

		if fsResizeRequired {
			// Resize volume succeeded and need to resize file system by kubelet, mark it as file system resizing required.
			return ctrl.markPVCAsFSResizeRequired(pvc)
		}
		// Resize volume succeeded and no need to resize file system by kubelet, mark it as resizing finished.
		return ctrl.markPVCResizeFinished(pvc, newSize)
	}()

	if err != nil {
		// Record an event to indicate that resize operation is failed.
		ctrl.eventRecorder.Eventf(pvc, v1.EventTypeWarning, util.VolumeResizeFailed, err.Error())
	}

	return err
}

// resizeVolume resize the volume to request size, and update PV's capacity if succeeded.
func (ctrl *resizeController) resizeVolume(
	pvc *v1.PersistentVolumeClaim,
	pv *v1.PersistentVolume) (resource.Quantity, bool, error) {

	// before trying expansion we will remove the PVC from map
	// that tracks PVCs which can't be expanded when in-use. If
	// pvc indeed can not be expanded when in-use then it will be added
	// back when expansion fails with in-use error.
	ctrl.usedPVCs.removePVCWithInUseError(pvc)

	requestSize := pvc.Spec.Resources.Requests[v1.ResourceStorage]

	newSize, fsResizeRequired, err := ctrl.resizer.Resize(pv, requestSize)

	if err != nil {
		klog.Errorf("Resize volume %q by resizer %q failed: %v", pv.Name, ctrl.name, err)
		// if this error was a in-use error then it must be tracked so as we don't retry without
		// first verifying if volume is in-use
		if inUseError(err) {
			ctrl.usedPVCs.addPVCWithInUseError(pvc)
		}
		return newSize, fsResizeRequired, fmt.Errorf("resize volume %s failed: %v", pv.Name, err)
	}
	klog.V(4).Infof("Resize volume succeeded for volume %q, start to update PV's capacity", pv.Name)

	if err := util.UpdatePVCapacity(pv, newSize, ctrl.kubeClient); err != nil {
		klog.Errorf("Update capacity of PV %q to %s failed: %v", pv.Name, newSize.String(), err)
		return newSize, fsResizeRequired, err
	}
	klog.V(4).Infof("Update capacity of PV %q to %s succeeded", pv.Name, newSize.String())

	return newSize, fsResizeRequired, nil
}

func (ctrl *resizeController) markPVCResizeInProgress(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	// Mark PVC as Resize Started
	progressCondition := v1.PersistentVolumeClaimCondition{
		Type:               v1.PersistentVolumeClaimResizing,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
	}
	newPVC := pvc.DeepCopy()
	newPVC.Status.Conditions = util.MergeResizeConditionsOfPVC(newPVC.Status.Conditions,
		[]v1.PersistentVolumeClaimCondition{progressCondition})
	return util.PatchPVCStatus(pvc, newPVC, ctrl.kubeClient)
}

func (ctrl *resizeController) markPVCResizeFinished(
	pvc *v1.PersistentVolumeClaim,
	newSize resource.Quantity) error {
	newPVC := pvc.DeepCopy()
	newPVC.Status.Capacity[v1.ResourceStorage] = newSize
	newPVC.Status.Conditions = util.MergeResizeConditionsOfPVC(pvc.Status.Conditions, []v1.PersistentVolumeClaimCondition{})
	if _, err := util.PatchPVCStatus(pvc, newPVC, ctrl.kubeClient); err != nil {
		klog.Errorf("Mark PVC %q as resize finished failed: %v", util.PVCKey(pvc), err)
		return err
	}

	klog.V(4).Infof("Resize PVC %q finished", util.PVCKey(pvc))
	ctrl.eventRecorder.Eventf(pvc, v1.EventTypeNormal, util.VolumeResizeSuccess, "Resize volume succeeded")

	return nil
}

func (ctrl *resizeController) markPVCAsFSResizeRequired(pvc *v1.PersistentVolumeClaim) error {
	pvcCondition := v1.PersistentVolumeClaimCondition{
		Type:               v1.PersistentVolumeClaimFileSystemResizePending,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            "Waiting for user to (re-)start a pod to finish file system resize of volume on node.",
	}
	newPVC := pvc.DeepCopy()
	newPVC.Status.Conditions = util.MergeResizeConditionsOfPVC(newPVC.Status.Conditions,
		[]v1.PersistentVolumeClaimCondition{pvcCondition})

	if _, err := util.PatchPVCStatus(pvc, newPVC, ctrl.kubeClient); err != nil {
		klog.Errorf("Mark PVC %q as file system resize required failed: %v", util.PVCKey(pvc), err)
		return err
	}
	klog.V(4).Infof("Mark PVC %q as file system resize required", util.PVCKey(pvc))
	ctrl.eventRecorder.Eventf(pvc, v1.EventTypeNormal,
		util.FileSystemResizeRequired, "Require file system resize of volume on node")

	return nil
}

func parsePod(obj interface{}) *v1.Pod {
	if obj == nil {
		return nil
	}
	pod, ok := obj.(*v1.Pod)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return nil
		}
		pod, ok = tombstone.Obj.(*v1.Pod)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a Pod %#v", obj))
			return nil
		}
	}
	return pod
}

func inUseError(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		// not a grpc error
		return false
	}
	// if this is a failed precondition error then that means driver does not support expansion
	// of in-use volumes
	if st.Code() == codes.FailedPrecondition {
		return true
	}
	return false
}

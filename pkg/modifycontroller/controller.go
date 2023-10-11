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
	"time"

	"github.com/kubernetes-csi/external-resizer/pkg/util"

	"github.com/kubernetes-csi/external-resizer/pkg/modifier"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	storagev1alpha1listers "k8s.io/client-go/listers/storage/v1alpha1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

// ModifyController watches PVCs and checks if they are requesting an modify operation.
// If requested, it will modify the volume according to parameters in VolumeAttributesClass
type ModifyController interface {
	// Run starts the controller.
	Run(workers int, ctx context.Context)
}

type modifyController struct {
	name            string
	modifier        modifier.Modifier
	kubeClient      kubernetes.Interface
	claimQueue      workqueue.RateLimitingInterface
	eventRecorder   record.EventRecorder
	pvLister        corelisters.PersistentVolumeLister
	pvListerSynced  cache.InformerSynced
	pvcLister       corelisters.PersistentVolumeClaimLister
	pvcListerSynced cache.InformerSynced
	vacLister       storagev1alpha1listers.VolumeAttributesClassLister
	vacListerSynced cache.InformerSynced
	// the key of the map is {PVC_NAMESPACE}/{PVC_NAME}
	uncertainPVCs map[string]v1.PersistentVolumeClaim
}

// NewModifyController returns a ModifyController.
func NewModifyController(
	name string,
	modifier modifier.Modifier,
	kubeClient kubernetes.Interface,
	resyncPeriod time.Duration,
	informerFactory informers.SharedInformerFactory,
	pvcRateLimiter workqueue.RateLimiter) ModifyController {
	pvInformer := informerFactory.Core().V1().PersistentVolumes()
	pvcInformer := informerFactory.Core().V1().PersistentVolumeClaims()
	vacInformer := informerFactory.Storage().V1alpha1().VolumeAttributesClasses()
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events(v1.NamespaceAll)})
	eventRecorder := eventBroadcaster.NewRecorder(scheme.Scheme,
		v1.EventSource{Component: fmt.Sprintf("external-resizer %s", name)})

	claimQueue := workqueue.NewNamedRateLimitingQueue(
		pvcRateLimiter, fmt.Sprintf("%s-pvc", name))

	ctrl := &modifyController{
		name:            name,
		modifier:        modifier,
		kubeClient:      kubeClient,
		pvListerSynced:  pvInformer.Informer().HasSynced,
		pvLister:        pvInformer.Lister(),
		pvcListerSynced: pvcInformer.Informer().HasSynced,
		pvcLister:       pvcInformer.Lister(),
		vacListerSynced: vacInformer.Informer().HasSynced,
		vacLister:       vacInformer.Lister(),
		claimQueue:      claimQueue,
		eventRecorder:   eventRecorder,
	}

	// Add a resync period as the PVC's request modify can be modified again when we handling
	// a previous modify request of the same PVC.
	pvcInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    ctrl.addPVC,
		UpdateFunc: ctrl.updatePVC,
		DeleteFunc: ctrl.deletePVC,
	}, resyncPeriod)

	// Add a resync period as the VAC can be created after a PVC is created
	// VAC is immutable
	vacInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{}, resyncPeriod)

	return ctrl
}

func (ctrl *modifyController) initUncertainPVCs() error {
	ctrl.uncertainPVCs = make(map[string]v1.PersistentVolumeClaim)
	allPVCs, err := ctrl.pvcLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to list pvcs when init uncertain pvcs: %v", err)
		return err
	}
	for _, pvc := range allPVCs {
		if pvc.Status.ModifyVolumeStatus != nil && (pvc.Status.ModifyVolumeStatus.Status == v1.PersistentVolumeClaimModifyVolumeInProgress || pvc.Status.ModifyVolumeStatus.Status == v1.PersistentVolumeClaimModifyVolumeInfeasible) {
			pvcKey, err := cache.MetaNamespaceKeyFunc(pvc)
			if err != nil {
				return err
			}
			ctrl.uncertainPVCs[pvcKey] = *pvc.DeepCopy()
		}
	}

	return nil
}

func (ctrl *modifyController) addPVC(obj interface{}) {
	objKey, err := util.GetObjectKey(obj)
	if err != nil {
		return
	}
	ctrl.claimQueue.Add(objKey)
}

func (ctrl *modifyController) updatePVC(oldObj, newObj interface{}) {
	oldPVC, ok := oldObj.(*v1.PersistentVolumeClaim)
	if !ok || oldPVC == nil {
		return
	}

	newPVC, ok := newObj.(*v1.PersistentVolumeClaim)
	if !ok || newPVC == nil {
		return
	}

	// Only trigger modify volume if the following conditions are met
	// 1. Non empty vac name
	// 2. oldVacName != newVacName
	// 3. PVC is in Bound state
	oldVacName := oldPVC.Spec.VolumeAttributesClassName
	newVacName := newPVC.Spec.VolumeAttributesClassName
	if newVacName != nil && *newVacName != "" && (*newVacName != *oldVacName || oldVacName == nil) && oldPVC.Status.Phase == v1.ClaimBound {
		_, err := ctrl.pvLister.Get(oldPVC.Spec.VolumeName)
		if err != nil {
			klog.Errorf("Get PV %q of pvc %q in PVInformer cache failed: %v", oldPVC.Spec.VolumeName, klog.KObj(oldPVC), err)
			return
		}
		// Handle modify volume by adding to the claimQueue to avoid race conditions
		ctrl.addPVC(newObj)
	} else {
		klog.V(4).InfoS("No need to modify PVC", "PVC", klog.KObj(newPVC))
	}
}

func (ctrl *modifyController) deletePVC(obj interface{}) {
	objKey, err := util.GetObjectKey(obj)
	if err != nil {
		return
	}
	ctrl.claimQueue.Forget(objKey)
}

// modifyPVC modifies the PVC and PV based on VAC
func (ctrl *modifyController) modifyPVC(pvc *v1.PersistentVolumeClaim, pv *v1.PersistentVolume) error {
	var err error
	if isFirstTimeModifyVolumeWithPVC(pvc, pv) {
		// If it is first time adding a vac, always validate and then call modify volume
		_, _, err, _ = ctrl.validateVACAndModifyVolumeWithTarget(pvc, pv)
	} else {
		_, _, err, _ = ctrl.modify(pvc, pv)
	}
	return err
}

func isFirstTimeModifyVolumeWithPVC(pvc *v1.PersistentVolumeClaim, pv *v1.PersistentVolume) bool {
	if pv.Spec.VolumeAttributesClassName == nil && pvc.Spec.VolumeAttributesClassName != nil {
		return true
	}
	return false
}

// Run starts the controller.
func (ctrl *modifyController) Run(
	workers int, ctx context.Context) {
	defer ctrl.claimQueue.ShutDown()

	klog.InfoS("Starting external resizer for modify volume", "controller", ctrl.name)
	defer klog.InfoS("Shutting down external resizer", "controller", ctrl.name)

	stopCh := ctx.Done()
	informersSyncd := []cache.InformerSynced{ctrl.pvListerSynced, ctrl.pvcListerSynced}
	informersSyncd = append(informersSyncd, ctrl.vacListerSynced)

	if !cache.WaitForCacheSync(stopCh, informersSyncd...) {
		klog.ErrorS(nil, "Cannot sync pod, pv, pvc or vac caches")
		return
	}

	// Cache all the InProgress/Infeasible PVCs as Uncertain for ModifyVolume
	err := ctrl.initUncertainPVCs()
	if err != nil {
		klog.ErrorS(err, "Failed to initialize uncertain pvcs")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(ctrl.sync, 0, stopCh)
	}

	<-stopCh
}

// sync is the main worker to sync PVCs.
func (ctrl *modifyController) sync() {
	key, quit := ctrl.claimQueue.Get()
	if quit {
		return
	}
	defer ctrl.claimQueue.Done(key)

	if err := ctrl.syncPVC(key.(string)); err != nil {
		// Put PVC back to the queue so that we can retry later.
		klog.ErrorS(err, "Error syncing PVC")
		ctrl.claimQueue.AddRateLimited(key)
	} else {
		ctrl.claimQueue.Forget(key)
	}
}

// syncPVC checks if a pvc requests resizing, and execute the resize operation if requested.
func (ctrl *modifyController) syncPVC(key string) error {
	klog.V(4).InfoS("Started PVC processing for modify controller", "key", key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("getting namespace and name from key %s failed: %v", key, err)
	}

	pvc, err := ctrl.pvcLister.PersistentVolumeClaims(namespace).Get(name)
	if err != nil {
		return fmt.Errorf("getting PVC %s/%s failed: %v", namespace, name, err)
	}

	if pvc.Spec.VolumeName == "" {
		klog.V(4).InfoS("PV bound to PVC is not created yet", "PVC", klog.KObj(pvc))
		return nil
	}

	pv, err := ctrl.pvLister.Get(pvc.Spec.VolumeName)
	if err != nil {
		return fmt.Errorf("Get PV %q of pvc %q in PVInformer cache failed: %v", pvc.Spec.VolumeName, klog.KObj(pvc), err)
	}

	// Only trigger modify volume if the following conditions are met
	// 1. Non empty vac name
	// 2. PVC is in Bound state
	vacName := pvc.Spec.VolumeAttributesClassName
	if vacName != nil && *vacName != "" && pvc.Status.Phase == v1.ClaimBound {
		_, _, err, _ := ctrl.modify(pvc, pv)
		if err != nil {
			return err
		}
	} else {
		klog.V(4).InfoS("No need to modify PV", "PV", klog.KObj(pv))
	}

	return nil
}

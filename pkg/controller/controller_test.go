package controller

import (
	"context"
	"testing"
	"time"

	"github.com/kubernetes-csi/external-resizer/pkg/features"

	"k8s.io/client-go/util/workqueue"

	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	"github.com/kubernetes-csi/external-resizer/pkg/resizer"
	"github.com/kubernetes-csi/external-resizer/pkg/util"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
)

func TestController(t *testing.T) {
	blockVolumeMode := v1.PersistentVolumeBlock
	fsVolumeMode := v1.PersistentVolumeFilesystem

	for _, test := range []struct {
		Name string
		PVC  *v1.PersistentVolumeClaim
		PV   *v1.PersistentVolume

		CreateObjects          bool
		NodeResize             bool
		CallCSIExpand          bool
		expectBlockVolume      bool
		expectDeleteAnnotation bool

		// is PVC being expanded in-use
		pvcInUse bool
		// does PVC being expanded has Failed Precondition errors
		pvcHasInUseErrors              bool
		disableVolumeInUseErrorHandler bool
		enableFSResizeAnnotation       bool
	}{
		{
			Name:          "Invalid key",
			PVC:           invalidPVC(),
			CallCSIExpand: false,
		},
		{
			Name:          "PVC not found",
			PVC:           createPVC(1, 1),
			CallCSIExpand: false,
		},
		{
			Name:          "PVC doesn't need resize",
			PVC:           createPVC(1, 1),
			CreateObjects: true,
			CallCSIExpand: false,
		},
		{
			Name:          "PV not found",
			PVC:           createPVC(2, 1),
			CreateObjects: true,
			CallCSIExpand: false,
		},
		{
			Name:          "pv claimref does not have pvc UID",
			PVC:           createPVC(2, 1),
			PV:            createPV(1, "testPVC" /*pvcName*/, defaultNS, "foobaz" /*pvcUID*/, &fsVolumeMode),
			CallCSIExpand: false,
		},
		{
			Name:          "pv claimref does not have PVC namespace",
			PVC:           createPVC(2, 1),
			PV:            createPV(1, "testPVC" /*pvcName*/, "test1" /*pvcNamespace*/, "foobar" /*pvcUID*/, &fsVolumeMode),
			CallCSIExpand: false,
		},
		{
			Name:          "pv claimref is nil",
			PVC:           createPVC(2, 1),
			PV:            createPV(1, "" /*pvcName*/, "test1" /*pvcNamespace*/, "foobar" /*pvcUID*/, &fsVolumeMode),
			CallCSIExpand: false,
		},
		{
			Name:          "Resize PVC, no FS resize",
			PVC:           createPVC(2, 1),
			PV:            createPV(1, "testPVC", defaultNS, "foobar", &fsVolumeMode),
			CreateObjects: true,
			CallCSIExpand: true,
		},
		{
			Name:          "Resize PVC with FS resize",
			PVC:           createPVC(2, 1),
			PV:            createPV(1, "testPVC", defaultNS, "foobar", &fsVolumeMode),
			CreateObjects: true,
			NodeResize:    true,
			CallCSIExpand: true,
		},
		{
			Name:                     "PV nodeExpand Complete",
			PVC:                      createPVC(2, 1),
			PV:                       createPV(1, "testPVC", defaultNS, "foobar", &fsVolumeMode),
			CreateObjects:            true,
			NodeResize:               true,
			CallCSIExpand:            true,
			expectDeleteAnnotation:   true,
			enableFSResizeAnnotation: true,
		},
		{
			Name:              "Block Resize PVC with FS resize",
			PVC:               createPVC(2, 1),
			PV:                createPV(1, "testPVC", defaultNS, "foobar", &blockVolumeMode),
			CreateObjects:     true,
			NodeResize:        true,
			CallCSIExpand:     true,
			expectBlockVolume: true,
		},
		{
			Name:              "Resize PVC, no FS resize, pvc-inuse with failedprecondition",
			PVC:               createPVC(2, 1),
			PV:                createPV(1, "testPVC", defaultNS, "foobar", &fsVolumeMode),
			CreateObjects:     true,
			CallCSIExpand:     false,
			pvcHasInUseErrors: true,
			pvcInUse:          true,
		},
		{
			Name:              "Resize PVC, no FS resize, pvc-inuse but no failedprecondition error",
			PVC:               createPVC(2, 1),
			PV:                createPV(1, "testPVC", defaultNS, "foobar", &fsVolumeMode),
			CreateObjects:     true,
			CallCSIExpand:     true,
			pvcHasInUseErrors: false,
			pvcInUse:          true,
		},
		{
			Name:              "Resize PVC, no FS resize, pvc not in-use but has failedprecondition error",
			PVC:               createPVC(2, 1),
			PV:                createPV(1, "testPVC", defaultNS, "foobar", &fsVolumeMode),
			CreateObjects:     true,
			CallCSIExpand:     true,
			pvcHasInUseErrors: true,
			pvcInUse:          false,
		},
		// test cases with volume in use error handling disabled.
		{
			Name:                           "With volume-in-use error handler disabled, Resize PVC, no FS resize, pvc-inuse with failedprecondition",
			PVC:                            createPVC(2, 1),
			PV:                             createPV(1, "testPVC", defaultNS, "foobar", &fsVolumeMode),
			CreateObjects:                  true,
			CallCSIExpand:                  true,
			pvcHasInUseErrors:              true,
			pvcInUse:                       true,
			disableVolumeInUseErrorHandler: true,
		},
		{
			Name:                           "With volume-in-use error handler disabled, Resize PVC, no FS resize, pvc-inuse but no failedprecondition error",
			PVC:                            createPVC(2, 1),
			PV:                             createPV(1, "testPVC", defaultNS, "foobar", &fsVolumeMode),
			CreateObjects:                  true,
			CallCSIExpand:                  true,
			pvcHasInUseErrors:              false,
			pvcInUse:                       true,
			disableVolumeInUseErrorHandler: true,
		},
		{
			Name:                           "With volume-in-use error handler disabled, Resize PVC, no FS resize, pvc not in-use but has failedprecondition error",
			PVC:                            createPVC(2, 1),
			PV:                             createPV(1, "testPVC", defaultNS, "foobar", &fsVolumeMode),
			CreateObjects:                  true,
			CallCSIExpand:                  true,
			pvcHasInUseErrors:              true,
			pvcInUse:                       false,
			disableVolumeInUseErrorHandler: true,
		},
		{
			Name:                           "With volume-in-use error handler disabled, Block Resize PVC with FS resize",
			PVC:                            createPVC(2, 1),
			PV:                             createPV(1, "testPVC", defaultNS, "foobar", &blockVolumeMode),
			CreateObjects:                  true,
			NodeResize:                     true,
			CallCSIExpand:                  true,
			expectBlockVolume:              true,
			disableVolumeInUseErrorHandler: true,
		},
		{
			Name:                           "With volume-in-use error handler disabled, Resize PVC with FS resize",
			PVC:                            createPVC(2, 1),
			PV:                             createPV(1, "testPVC", defaultNS, "foobar", &fsVolumeMode),
			CreateObjects:                  true,
			NodeResize:                     true,
			CallCSIExpand:                  true,
			disableVolumeInUseErrorHandler: true,
		},
		{
			Name:                           "With volume-in-use error handler disabled, Resize PVC, no FS resize",
			PVC:                            createPVC(2, 1),
			PV:                             createPV(1, "testPVC", defaultNS, "foobar", &fsVolumeMode),
			CreateObjects:                  true,
			CallCSIExpand:                  true,
			disableVolumeInUseErrorHandler: true,
		},
	} {
		client := csi.NewMockClient("mock", test.NodeResize, true, false, true, true)
		driverName, _ := client.GetDriverName(context.TODO())

		var expectedCap resource.Quantity
		initialObjects := []runtime.Object{}
		if test.CreateObjects {
			if test.PVC != nil {
				initialObjects = append(initialObjects, test.PVC)
				expectedCap = test.PVC.Status.Capacity[v1.ResourceStorage]
			}
			if test.PV != nil {
				test.PV.Spec.PersistentVolumeSource.CSI.Driver = driverName
				initialObjects = append(initialObjects, test.PV)
			}
		}

		if test.pvcInUse {
			pod := withPVC(test.PVC.Name, pod())
			initialObjects = append(initialObjects, pod)
		}

		kubeClient, informerFactory := fakeK8s(initialObjects)
		pvInformer := informerFactory.Core().V1().PersistentVolumes()
		pvcInformer := informerFactory.Core().V1().PersistentVolumeClaims()
		podInformer := informerFactory.Core().V1().Pods()

		csiResizer, err := resizer.NewResizerFromClient(client, 15*time.Second, kubeClient, driverName)
		if err != nil {
			t.Fatalf("Test %s: Unable to create resizer: %v", test.Name, err)
		}

		defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.AnnotateFsResize, true)()
		controller := NewResizeController(driverName, csiResizer, kubeClient, time.Second, informerFactory, workqueue.DefaultControllerRateLimiter(), !test.disableVolumeInUseErrorHandler)

		ctrlInstance, _ := controller.(*resizeController)

		if test.pvcHasInUseErrors {
			ctrlInstance.usedPVCs.addPVCWithInUseError(test.PVC)
			if !ctrlInstance.usedPVCs.hasInUseErrors(test.PVC) {
				t.Fatalf("pvc %s does not have in-use errors", test.PVC.Name)
			}
		}

		stopCh := make(chan struct{})
		informerFactory.Start(stopCh)

		ctx := context.TODO()
		defer ctx.Done()
		go controller.Run(1, ctx)

		for _, obj := range initialObjects {
			switch obj.(type) {
			case *v1.PersistentVolume:
				pvInformer.Informer().GetStore().Add(obj)
			case *v1.PersistentVolumeClaim:
				pvcInformer.Informer().GetStore().Add(obj)
			case *v1.Pod:
				podInformer.Informer().GetStore().Add(obj)
			default:
				t.Fatalf("Test %s: Unknown initalObject type: %+v", test.Name, obj)
			}
		}

		time.Sleep(time.Second * 2)

		expandCallCount := client.GetExpandCount()
		if test.CallCSIExpand && expandCallCount == 0 {
			t.Fatalf("for %s: expected csi expand call, no csi expand call was made", test.Name)
		}

		if !test.CallCSIExpand && expandCallCount > 0 {
			t.Fatalf("for %s: expected no csi expand call, received csi expansion request", test.Name)
		}

		usedCapability := client.GetCapability()

		if test.CallCSIExpand && test.expectBlockVolume && usedCapability.GetBlock() == nil {
			t.Errorf("For %s: expected block accesstype got: %v", test.Name, usedCapability)
		}

		if test.CallCSIExpand && !test.expectBlockVolume && usedCapability.GetMount() == nil {
			t.Errorf("For %s: expected mount accesstype got: %v", test.Name, usedCapability)
		}

		// check if pre resize capacity annotation get properly populated after volume resize if node expand is required
		if test.enableFSResizeAnnotation && test.PV != nil && test.CreateObjects {
			volObj, _, _ := ctrlInstance.volumes.GetByKey("testPV")
			pv := volObj.(*v1.PersistentVolume)
			requestCap, statusCap := test.PVC.Spec.Resources.Requests[v1.ResourceStorage], test.PVC.Status.Capacity[v1.ResourceStorage]
			if requestCap.Cmp(statusCap) > 0 && test.NodeResize && !expectedCap.IsZero() {
				checkPreResizeCap(t, test.Name, pv, expectedCap.String())
			}
		}

		// check if pre resize capacity annotation gets properly deleted after node expand
		if test.enableFSResizeAnnotation && test.PVC != nil && test.CreateObjects && test.expectDeleteAnnotation {
			ctrlInstance.markPVCResizeFinished(test.PVC, test.PVC.Spec.Resources.Requests[v1.ResourceStorage])
			time.Sleep(time.Second * 2)
			volObj, _, _ := ctrlInstance.volumes.GetByKey("testPV")
			pv := volObj.(*v1.PersistentVolume)
			if pv.ObjectMeta.Annotations != nil {
				if _, exists := pv.ObjectMeta.Annotations[util.AnnPreResizeCapacity]; exists {
					t.Errorf("For %s: expected annotation %s to be empty, but received %s", test.Name, util.AnnPreResizeCapacity, pv.ObjectMeta.Annotations[util.AnnPreResizeCapacity])
				}
			}
		}
	}
}

func TestResizePVC(t *testing.T) {
	fsVolumeMode := v1.PersistentVolumeFilesystem

	for _, test := range []struct {
		Name string
		PVC  *v1.PersistentVolumeClaim
		PV   *v1.PersistentVolume

		NodeResize       bool
		expansionFailure bool
		expectFailure    bool
	}{
		{
			Name:       "Resize PVC with FS resize",
			PVC:        createPVC(2, 1),
			PV:         createPV(1, "testPVC", defaultNS, "foobar", &fsVolumeMode),
			NodeResize: true,
		},
		{
			Name:             "Resize PVC with FS resize failure",
			PVC:              createPVC(2, 1),
			PV:               createPV(1, "testPVC", defaultNS, "foobar", &fsVolumeMode),
			NodeResize:       true,
			expansionFailure: true,
			expectFailure:    true,
		},
	} {
		client := csi.NewMockClient("mock", test.NodeResize, true, false, true, true)
		if test.expansionFailure {
			client.SetExpansionFailed()
		}
		driverName, _ := client.GetDriverName(context.TODO())

		var expectedCap resource.Quantity
		initialObjects := []runtime.Object{}
		if test.PVC != nil {
			initialObjects = append(initialObjects, test.PVC)
			expectedCap = test.PVC.Status.Capacity[v1.ResourceStorage]
		}
		if test.PV != nil {
			test.PV.Spec.PersistentVolumeSource.CSI.Driver = driverName
			initialObjects = append(initialObjects, test.PV)
		}

		kubeClient, informerFactory := fakeK8s(initialObjects)
		pvInformer := informerFactory.Core().V1().PersistentVolumes()
		pvcInformer := informerFactory.Core().V1().PersistentVolumeClaims()
		podInformer := informerFactory.Core().V1().Pods()

		csiResizer, err := resizer.NewResizerFromClient(client, 15*time.Second, kubeClient, driverName)
		if err != nil {
			t.Fatalf("Test %s: Unable to create resizer: %v", test.Name, err)
		}

		defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.AnnotateFsResize, true)()
		controller := NewResizeController(driverName, csiResizer, kubeClient, time.Second, informerFactory, workqueue.DefaultControllerRateLimiter(), true /* disableVolumeInUseErrorHandler*/)

		ctrlInstance, _ := controller.(*resizeController)

		stopCh := make(chan struct{})
		informerFactory.Start(stopCh)

		for _, obj := range initialObjects {
			switch obj.(type) {
			case *v1.PersistentVolume:
				pvInformer.Informer().GetStore().Add(obj)
			case *v1.PersistentVolumeClaim:
				pvcInformer.Informer().GetStore().Add(obj)
			case *v1.Pod:
				podInformer.Informer().GetStore().Add(obj)
			default:
				t.Fatalf("Test %s: Unknown initalObject type: %+v", test.Name, obj)
			}
		}

		err = ctrlInstance.resizePVC(test.PVC, test.PV)
		if test.expectFailure && err == nil {
			t.Errorf("for %s expected error got nothing", test.Name)
			continue
		}
		if !test.expectFailure {
			if err != nil {
				t.Errorf("for %s, unexpected error: %v", test.Name, err)
				// check if pre resize capacity annotation gets properly populated after resize
			} else if test.PV != nil && test.NodeResize && !expectedCap.IsZero() {
				volObj, _, _ := ctrlInstance.volumes.GetByKey("testPV")
				pv := volObj.(*v1.PersistentVolume)
				checkPreResizeCap(t, test.Name, pv, expectedCap.String())
			}
		}
	}
}

func checkPreResizeCap(t *testing.T, testName string, pv *v1.PersistentVolume, expectedCap string) {
	if pv.ObjectMeta.Annotations == nil {
		t.Errorf("for %s, AnnpreResizeCapacity was not successfully updated, expected: %s, received: nil Annotations", testName, expectedCap)
	} else if pv.ObjectMeta.Annotations[util.AnnPreResizeCapacity] != expectedCap {
		t.Errorf("for %s, AnnpreResizeCapacity was not successfully updated, expected: %s, received: %s", testName, expectedCap, pv.ObjectMeta.Annotations[util.AnnPreResizeCapacity])
	}
}

func invalidPVC() *v1.PersistentVolumeClaim {
	pvc := createPVC(1, 1)
	pvc.ObjectMeta.Name = ""
	pvc.ObjectMeta.Namespace = ""

	return pvc
}

func quantityGB(i int) resource.Quantity {
	q := resource.NewQuantity(int64(i*1024*1024*1024), resource.BinarySI)
	return *q
}

func createPVC(requestGB, capacityGB int) *v1.PersistentVolumeClaim {
	request := quantityGB(requestGB)
	capacity := quantityGB(capacityGB)

	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testPVC",
			Namespace: defaultNS,
			UID:       "foobar",
		},
		Spec: v1.PersistentVolumeClaimSpec{
			Resources: v1.VolumeResourceRequirements{
				Requests: map[v1.ResourceName]resource.Quantity{
					v1.ResourceStorage: request,
				},
			},
			VolumeName: "testPV",
		},
		Status: v1.PersistentVolumeClaimStatus{
			Phase: v1.ClaimBound,
			Capacity: map[v1.ResourceName]resource.Quantity{
				v1.ResourceStorage: capacity,
			},
		},
	}
}

func createPV(capacityGB int, pvcName, pvcNamespace string, pvcUID types.UID, volumeMode *v1.PersistentVolumeMode) *v1.PersistentVolume {
	capacity := quantityGB(capacityGB)

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "testPV",
			Annotations: make(map[string]string),
		},
		Spec: v1.PersistentVolumeSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Capacity: map[v1.ResourceName]resource.Quantity{
				v1.ResourceStorage: capacity,
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				CSI: &v1.CSIPersistentVolumeSource{
					Driver:       "foo",
					VolumeHandle: "foo",
				},
			},
			VolumeMode: volumeMode,
		},
	}
	if len(pvcName) > 0 {
		pv.Spec.ClaimRef = &v1.ObjectReference{
			Namespace: pvcNamespace,
			Name:      pvcName,
			UID:       pvcUID,
		}
		pv.Status.Phase = v1.VolumeBound
	}
	return pv
}

func fakeK8s(objs []runtime.Object) (kubernetes.Interface, informers.SharedInformerFactory) {
	client := fake.NewSimpleClientset(objs...)
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	return client, informerFactory
}

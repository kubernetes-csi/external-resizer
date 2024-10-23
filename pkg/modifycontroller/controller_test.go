package modifycontroller

import (
	"context"
	"testing"
	"time"

	"github.com/kubernetes-csi/external-resizer/pkg/features"

	"k8s.io/client-go/util/workqueue"

	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	"github.com/kubernetes-csi/external-resizer/pkg/modifier"

	v1 "k8s.io/api/core/v1"
	storagev1beta1 "k8s.io/api/storage/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
)

func TestController(t *testing.T) {
	basePVC := createTestPVC(pvcName, testVac /*vacName*/, testVac /*curVacName*/, testVac /*targetVacName*/)
	basePV := createTestPV(1, pvcName, pvcNamespace, "foobaz" /*pvcUID*/, &fsVolumeMode, testVac)
	firstTimePV := basePV.DeepCopy()
	firstTimePV.Spec.VolumeAttributesClassName = nil
	firstTimePVC := basePVC.DeepCopy()
	firstTimePVC.Status.CurrentVolumeAttributesClassName = nil
	firstTimePVC.Status.ModifyVolumeStatus = nil

	tests := []struct {
		name          string
		pvc           *v1.PersistentVolumeClaim
		pv            *v1.PersistentVolume
		vacExists     bool
		callCSIModify bool
	}{
		{
			name:          "Modify called",
			pvc:           createTestPVC(pvcName, targetVac /*vacName*/, testVac /*curVacName*/, testVac /*targetVacName*/),
			pv:            basePV,
			vacExists:     true,
			callCSIModify: true,
		},
		{
			name:          "Nothing to modify",
			pvc:           basePVC,
			pv:            basePV,
			vacExists:     true,
			callCSIModify: false,
		},
		{
			name:          "First time modify",
			pvc:           firstTimePVC,
			pv:            firstTimePV,
			vacExists:     true,
			callCSIModify: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup
			client := csi.NewMockClient(testDriverName, true, true, true, true, true, false)

			initialObjects := []runtime.Object{test.pvc, test.pv, testVacObject, targetVacObject}
			ctrlInstance := setupFakeK8sEnvironment(t, client, initialObjects)

			_ = ctrlInstance.modifyPVC(test.pvc, test.pv)

			modifyCallCount := client.GetModifyCount()
			if test.callCSIModify && modifyCallCount == 0 {
				t.Fatalf("for %s: expected csi modify call, no csi modify call was made", test.name)
			}

			if !test.callCSIModify && modifyCallCount > 0 {
				t.Fatalf("for %s: expected no csi modify call, received csi modify request", test.name)
			}
		})
	}

}

func TestModifyPVC(t *testing.T) {
	basePV := createTestPV(1, pvcName, pvcNamespace, "foobaz" /*pvcUID*/, &fsVolumeMode, testVac)

	tests := []struct {
		name          string
		pvc           *v1.PersistentVolumeClaim
		pv            *v1.PersistentVolume
		modifyFailure bool
		expectFailure bool
	}{
		{
			name:          "Modify succeeded",
			pvc:           createTestPVC(pvcName, targetVac /*vacName*/, testVac /*curVacName*/, testVac /*targetVacName*/),
			pv:            basePV,
			modifyFailure: false,
			expectFailure: false,
		},
		{
			name:          "Modify failed",
			pvc:           createTestPVC(pvcName, targetVac /*vacName*/, testVac /*curVacName*/, testVac /*targetVacName*/),
			pv:            basePV,
			modifyFailure: true,
			expectFailure: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := csi.NewMockClient(testDriverName, true, true, true, true, true, false)
			if test.modifyFailure {
				client.SetModifyFailed()
			}

			initialObjects := []runtime.Object{test.pvc, test.pv, testVacObject, targetVacObject}
			ctrlInstance := setupFakeK8sEnvironment(t, client, initialObjects)

			_, _, err, _ := ctrlInstance.modify(test.pvc, test.pv)

			if test.expectFailure && err == nil {
				t.Errorf("for %s expected error got nothing", test.name)
			}

			if !test.expectFailure {
				if err != nil {
					t.Errorf("for %s, unexpected error: %v", test.name, err)
				}
			}
		})
	}
}

// setupFakeK8sEnvironment
func setupFakeK8sEnvironment(t *testing.T, client *csi.MockClient, initialObjects []runtime.Object) *modifyController {
	t.Helper()

	featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.VolumeAttributesClass, true)

	/* Create fake kubeClient, Informers, and ModifyController */
	kubeClient, informerFactory := fakeK8s(initialObjects)
	pvInformer := informerFactory.Core().V1().PersistentVolumes()
	pvcInformer := informerFactory.Core().V1().PersistentVolumeClaims()
	vacInformer := informerFactory.Storage().V1beta1().VolumeAttributesClasses()

	driverName, _ := client.GetDriverName(context.TODO())

	csiModifier, err := modifier.NewModifierFromClient(client, 15*time.Second, kubeClient, informerFactory, false, driverName)
	if err != nil {
		t.Fatalf("Test %s: Unable to create modifier: %v", t.Name(), err)
	}

	controller := NewModifyController(driverName,
		csiModifier, kubeClient,
		0, false, informerFactory,
		workqueue.DefaultControllerRateLimiter())

	/* Start informers and ModifyController*/
	stopCh := make(chan struct{})
	informerFactory.Start(stopCh)

	ctx := context.TODO()
	defer ctx.Done()
	go controller.Run(1, ctx)

	/* Add initial objects to informer caches (TODO Q confirm this is true/needed?) */
	for _, obj := range initialObjects {
		switch obj.(type) {
		case *v1.PersistentVolume:
			pvInformer.Informer().GetStore().Add(obj)
		case *v1.PersistentVolumeClaim:
			pvcInformer.Informer().GetStore().Add(obj)
		case *storagev1beta1.VolumeAttributesClass:
			vacInformer.Informer().GetStore().Add(obj)
		default:
			t.Fatalf("Test %s: Unknown initalObject type: %+v", t.Name(), obj)
		}
	}

	ctrlInstance, _ := controller.(*modifyController)

	return ctrlInstance
}

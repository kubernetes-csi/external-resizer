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
	storagev1alpha1 "k8s.io/api/storage/v1alpha1"
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
			pvc:           createTestPVC(pvcName, "target-vac" /*vacName*/, testVac /*curVacName*/, testVac /*targetVacName*/),
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
			client := csi.NewMockClient("foo", true, true, true, true, true)
			driverName, _ := client.GetDriverName(context.TODO())

			var initialObjects []runtime.Object
			initialObjects = append(initialObjects, test.pvc)
			initialObjects = append(initialObjects, test.pv)
			// existing vac set in the pvc and pv
			initialObjects = append(initialObjects, testVacObject)
			if test.vacExists {
				initialObjects = append(initialObjects, targetVacObject)
			}

			kubeClient, informerFactory := fakeK8s(initialObjects)
			pvInformer := informerFactory.Core().V1().PersistentVolumes()
			pvcInformer := informerFactory.Core().V1().PersistentVolumeClaims()
			vacInformer := informerFactory.Storage().V1alpha1().VolumeAttributesClasses()

			csiModifier, err := modifier.NewModifierFromClient(client, 15*time.Second, kubeClient, informerFactory, driverName)
			if err != nil {
				t.Fatalf("Test %s: Unable to create modifier: %v", test.name, err)
			}

			defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.VolumeAttributesClass, true)()
			controller := NewModifyController(driverName,
				csiModifier, kubeClient,
				time.Second, informerFactory,
				workqueue.DefaultControllerRateLimiter())

			ctrlInstance, _ := controller.(*modifyController)

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
				case *storagev1alpha1.VolumeAttributesClass:
					vacInformer.Informer().GetStore().Add(obj)
				default:
					t.Fatalf("Test %s: Unknown initalObject type: %+v", test.name, obj)
				}
			}
			time.Sleep(time.Second * 2)
			err = ctrlInstance.modifyPVC(test.pvc, test.pv)

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
			pvc:           createTestPVC(pvcName, "target-vac" /*vacName*/, testVac /*curVacName*/, testVac /*targetVacName*/),
			pv:            basePV,
			modifyFailure: false,
			expectFailure: false,
		},
		{
			name:          "Modify failed",
			pvc:           createTestPVC(pvcName, "target-vac" /*vacName*/, testVac /*curVacName*/, testVac /*targetVacName*/),
			pv:            basePV,
			modifyFailure: true,
			expectFailure: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := csi.NewMockClient("mock", true, true, true, true, true)
			if test.modifyFailure {
				client.SetModifyFailed()
			}
			driverName, _ := client.GetDriverName(context.TODO())

			initialObjects := []runtime.Object{}
			if test.pvc != nil {
				initialObjects = append(initialObjects, test.pvc)
			}
			if test.pv != nil {
				test.pv.Spec.PersistentVolumeSource.CSI.Driver = driverName
				initialObjects = append(initialObjects, test.pv)
			}

			// existing vac set in the pvc and pv
			initialObjects = append(initialObjects, testVacObject)
			// new vac used in modify volume
			initialObjects = append(initialObjects, targetVacObject)

			kubeClient, informerFactory := fakeK8s(initialObjects)
			pvInformer := informerFactory.Core().V1().PersistentVolumes()
			pvcInformer := informerFactory.Core().V1().PersistentVolumeClaims()
			vacInformer := informerFactory.Storage().V1alpha1().VolumeAttributesClasses()

			csiModifier, err := modifier.NewModifierFromClient(client, 15*time.Second, kubeClient, informerFactory, driverName)
			if err != nil {
				t.Fatalf("Test %s: Unable to create modifier: %v", test.name, err)
			}

			defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.VolumeAttributesClass, true)()
			controller := NewModifyController(driverName,
				csiModifier, kubeClient,
				time.Second, informerFactory,
				workqueue.DefaultControllerRateLimiter())

			ctrlInstance, _ := controller.(*modifyController)

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
				case *storagev1alpha1.VolumeAttributesClass:
					vacInformer.Informer().GetStore().Add(obj)
				default:
					t.Fatalf("Test %s: Unknown initalObject type: %+v", test.name, obj)
				}
			}

			time.Sleep(time.Second * 2)

			_, _, err, _ = ctrlInstance.modify(test.pvc, test.pv)

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

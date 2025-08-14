package modifycontroller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	"github.com/kubernetes-csi/external-resizer/pkg/features"
	"github.com/kubernetes-csi/external-resizer/pkg/modifier"
	"github.com/kubernetes-csi/external-resizer/pkg/testutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
)

const (
	pvcName      = "foo"
	pvcNamespace = "modify"
	pvName       = "testPV"
)

var (
	fsVolumeMode           = v1.PersistentVolumeFilesystem
	testVac                = "test-vac"
	targetVac              = "target-vac"
	testDriverName         = "mock"
	infeasibleErr          = status.Errorf(codes.InvalidArgument, "Parameters in VolumeAttributesClass is invalid")
	pvcConditionInProgress = v1.PersistentVolumeClaimCondition{
		Type:    v1.PersistentVolumeClaimVolumeModifyingVolume,
		Status:  v1.ConditionTrue,
		Message: "ModifyVolume operation in progress.",
	}

	pvcConditionInfeasible = v1.PersistentVolumeClaimCondition{
		Type:    v1.PersistentVolumeClaimVolumeModifyingVolume,
		Status:  v1.ConditionTrue,
		Message: "ModifyVolume failed with error: rpc error: code = InvalidArgument desc = Parameters in VolumeAttributesClass is invalid. Waiting for retry.",
	}

	pvcConditionError = v1.PersistentVolumeClaimCondition{
		Type:    v1.PersistentVolumeClaimVolumeModifyVolumeError,
		Status:  v1.ConditionTrue,
		Message: "ModifyVolume failed with error. Waiting for retry.",
	}
)

func TestMarkControllerModifyVolumeStatus(t *testing.T) {
	basePVC := testutil.MakeTestPVC([]v1.PersistentVolumeClaimCondition{})

	tests := []struct {
		name               string
		pvc                *v1.PersistentVolumeClaim
		expectedPVC        *v1.PersistentVolumeClaim
		expectedConditions []v1.PersistentVolumeClaimCondition
		expectedErr        error
		testFunc           func(*v1.PersistentVolumeClaim, *modifyController) (*v1.PersistentVolumeClaim, error)
	}{
		{
			name:               "mark modify volume as in progress",
			pvc:                basePVC.Get(),
			expectedPVC:        basePVC.WithModifyVolumeStatus(v1.PersistentVolumeClaimModifyVolumeInProgress).Get(),
			expectedConditions: []v1.PersistentVolumeClaimCondition{pvcConditionInProgress},
			expectedErr:        nil,
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *modifyController) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markControllerModifyVolumeStatus(pvc, v1.PersistentVolumeClaimModifyVolumeInProgress, nil)
			},
		},
		{
			name:               "mark modify volume as infeasible",
			pvc:                basePVC.Get(),
			expectedPVC:        basePVC.WithModifyVolumeStatus(v1.PersistentVolumeClaimModifyVolumeInfeasible).Get(),
			expectedConditions: []v1.PersistentVolumeClaimCondition{pvcConditionInfeasible},
			expectedErr:        nil,
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *modifyController) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markControllerModifyVolumeStatus(pvc, v1.PersistentVolumeClaimModifyVolumeInfeasible, infeasibleErr)
			},
		},
		{
			name:               "mark modify volume as pending",
			pvc:                basePVC.Get(),
			expectedPVC:        basePVC.WithModifyVolumeStatus(v1.PersistentVolumeClaimModifyVolumePending).Get(),
			expectedConditions: nil,
			expectedErr:        nil,
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *modifyController) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markControllerModifyVolumeStatus(pvc, v1.PersistentVolumeClaimModifyVolumePending, nil)
			},
		},
	}

	for _, test := range tests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.VolumeAttributesClass, true)
			client := csi.NewMockClient("foo", true, true, true, true, true)
			driverName, _ := client.GetDriverName(context.TODO())

			pvc := test.pvc

			var initialObjects []runtime.Object
			initialObjects = append(initialObjects, test.pvc)

			kubeClient, informerFactory := fakeK8s(initialObjects)

			csiModifier, err := modifier.NewModifierFromClient(client, 15*time.Second, kubeClient, informerFactory, false, driverName)
			if err != nil {
				t.Fatalf("Test %s: Unable to create modifier: %v", test.name, err)
			}
			controller := NewModifyController(driverName,
				csiModifier, kubeClient,
				time.Second, 2*time.Minute, false, informerFactory,
				workqueue.DefaultTypedControllerRateLimiter[string]())

			ctrlInstance, _ := controller.(*modifyController)

			pvc, err = tc.testFunc(pvc, ctrlInstance)
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error: %v, got: %v", tc.expectedErr, err)
			}

			realStatus := pvc.Status.ModifyVolumeStatus.Status
			expectedStatus := tc.expectedPVC.Status.ModifyVolumeStatus.Status
			if diff := cmp.Diff(expectedStatus, realStatus); diff != "" {
				t.Errorf("unexpected modify volume status (-want +got):\n%s", diff)
			}

			realConditions := pvc.Status.Conditions
			if !testutil.CompareConditions(realConditions, tc.expectedConditions) {
				t.Errorf("expected conditions %+v got %+v", tc.expectedConditions, realConditions)
			}
		})
	}
}

func TestUpdateConditionBasedOnError(t *testing.T) {
	basePVC := testutil.MakeTestPVC([]v1.PersistentVolumeClaimCondition{})

	tests := []struct {
		name               string
		pvc                *v1.PersistentVolumeClaim
		expectedConditions []v1.PersistentVolumeClaimCondition
		expectedErr        error
	}{
		{
			name:               "update condition based on error",
			pvc:                basePVC.Get(),
			expectedConditions: []v1.PersistentVolumeClaimCondition{pvcConditionError},
		},
	}

	for _, test := range tests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.VolumeAttributesClass, true)
			client := csi.NewMockClient("foo", true, true, true, true, true)
			driverName, _ := client.GetDriverName(context.TODO())

			var initialObjects []runtime.Object
			initialObjects = append(initialObjects, test.pvc)

			kubeClient, informerFactory := fakeK8s(initialObjects)

			csiModifier, err := modifier.NewModifierFromClient(client, 15*time.Second, kubeClient, informerFactory, false, driverName)
			if err != nil {
				t.Fatalf("Test %s: Unable to create modifier: %v", test.name, err)
			}
			controller := NewModifyController(driverName,
				csiModifier, kubeClient,
				time.Second, 2*time.Minute, false, informerFactory,
				workqueue.DefaultTypedControllerRateLimiter[string]())

			ctrlInstance, _ := controller.(*modifyController)

			pvc, err := ctrlInstance.updateConditionBasedOnError(tc.pvc, err)
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error: %v, got: %v", tc.expectedErr, err)
			}

			if !testutil.CompareConditions(pvc.Status.Conditions, tc.expectedConditions) {
				t.Errorf("expected conditions %+v got %+v", tc.expectedConditions, pvc.Status.Conditions)
			}
		})
	}
}

func TestMarkControllerModifyVolumeCompleted(t *testing.T) {
	basePVC := testutil.MakeTestPVC([]v1.PersistentVolumeClaimCondition{})
	basePV := createTestPV(1, pvcName, pvcNamespace, "foobaz" /*pvcUID*/, &fsVolumeMode, testVac)
	expectedPV := basePV.DeepCopy()
	expectedPV.Spec.VolumeAttributesClassName = &targetVac
	expectedPVC := basePVC.WithCurrentVolumeAttributesClassName(targetVac).Get()
	expectedPVC.Status.ModifyVolumeStatus = nil

	tests := []struct {
		name        string
		pvc         *v1.PersistentVolumeClaim
		pv          *v1.PersistentVolume
		expectedPVC *v1.PersistentVolumeClaim
		expectedPV  *v1.PersistentVolume
		expectedErr error
	}{
		{
			name:        "update modify volume status to completed",
			pvc:         basePVC.Get(),
			pv:          basePV,
			expectedPVC: expectedPVC,
			expectedPV:  expectedPV,
		},
		{
			name:        "update modify volume status to completed, and clear conditions",
			pvc:         basePVC.WithConditions([]v1.PersistentVolumeClaimCondition{pvcConditionInProgress}).Get(),
			pv:          basePV,
			expectedPVC: expectedPVC,
			expectedPV:  expectedPV,
		},
	}

	for _, test := range tests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.VolumeAttributesClass, true)
			client := csi.NewMockClient("foo", true, true, true, true, true)
			driverName, _ := client.GetDriverName(context.TODO())

			var initialObjects []runtime.Object
			initialObjects = append(initialObjects, test.pvc)
			initialObjects = append(initialObjects, test.pv)

			kubeClient, informerFactory := fakeK8s(initialObjects)

			csiModifier, err := modifier.NewModifierFromClient(client, 15*time.Second, kubeClient, informerFactory, false, driverName)
			if err != nil {
				t.Fatalf("Test %s: Unable to create modifier: %v", test.name, err)
			}
			controller := NewModifyController(driverName,
				csiModifier, kubeClient,
				time.Second, 2*time.Minute, false, informerFactory,
				workqueue.DefaultTypedControllerRateLimiter[string]())

			ctrlInstance, _ := controller.(*modifyController)

			actualPVC, pv, err := ctrlInstance.markControllerModifyVolumeCompleted(tc.pvc, tc.pv)
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error: %v, got: %v", tc.expectedErr, err)
			}

			if len(actualPVC.Status.Conditions) == 0 {
				actualPVC.Status.Conditions = []v1.PersistentVolumeClaimCondition{}
			}

			if diff := cmp.Diff(tc.expectedPVC, actualPVC); diff != "" {
				t.Errorf("unexpected pvc (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.expectedPV, pv); diff != "" {
				t.Errorf("unexpected pv (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRemovePVCFromModifyVolumeUncertainCache(t *testing.T) {
	basePVC := testutil.MakeTestPVC([]v1.PersistentVolumeClaimCondition{})
	basePVC.WithModifyVolumeStatus(v1.PersistentVolumeClaimModifyVolumeInProgress)
	secondPVC := testutil.GetTestPVC("test-vol0", "2G", "1G", "", "")
	secondPVC.Status.Phase = v1.ClaimBound
	secondPVC.Status.ModifyVolumeStatus = &v1.ModifyVolumeStatus{}
	secondPVC.Status.ModifyVolumeStatus.Status = v1.PersistentVolumeClaimModifyVolumeInfeasible

	tests := []struct {
		name string
		pvc  *v1.PersistentVolumeClaim
	}{
		{
			name: "should delete the target pvc but keep the others in the cache",
			pvc:  basePVC.Get(),
		},
	}

	for _, test := range tests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.VolumeAttributesClass, true)
			client := csi.NewMockClient("foo", true, true, true, true, true)
			driverName, _ := client.GetDriverName(context.TODO())

			var initialObjects []runtime.Object
			initialObjects = append(initialObjects, test.pvc)
			initialObjects = append(initialObjects, secondPVC)

			kubeClient, informerFactory := fakeK8s(initialObjects)
			pvInformer := informerFactory.Core().V1().PersistentVolumes()
			pvcInformer := informerFactory.Core().V1().PersistentVolumeClaims()
			podInformer := informerFactory.Core().V1().Pods()
			vacInformer := informerFactory.Storage().V1().VolumeAttributesClasses()

			csiModifier, err := modifier.NewModifierFromClient(client, 15*time.Second, kubeClient, informerFactory, false, driverName)
			if err != nil {
				t.Fatalf("Test %s: Unable to create modifier: %v", test.name, err)
			}
			controller := NewModifyController(driverName,
				csiModifier, kubeClient,
				time.Second, 2*time.Minute, false, informerFactory,
				workqueue.DefaultTypedControllerRateLimiter[string]())

			ctrlInstance, _ := controller.(*modifyController)

			stopCh := make(chan struct{})
			informerFactory.Start(stopCh)

			success := ctrlInstance.init(t.Context())
			if !success {
				t.Fatal("failed to init controller")
			}

			for _, obj := range initialObjects {
				switch obj.(type) {
				case *v1.PersistentVolume:
					pvInformer.Informer().GetStore().Add(obj)
				case *v1.PersistentVolumeClaim:
					pvcInformer.Informer().GetStore().Add(obj)
				case *v1.Pod:
					podInformer.Informer().GetStore().Add(obj)
				case *storagev1.VolumeAttributesClass:
					vacInformer.Informer().GetStore().Add(obj)
				default:
					t.Fatalf("Test %s: Unknown initalObject type: %+v", test.name, obj)
				}
			}

			time.Sleep(time.Second * 2)

			pvcKey, err := cache.MetaNamespaceKeyFunc(tc.pvc)
			if err != nil {
				t.Errorf("failed to extract pvc key from pvc %v", tc.pvc)
			}
			ctrlInstance.removePVCFromModifyVolumeUncertainCache(pvcKey)

			deletedPVCKey, err := cache.MetaNamespaceKeyFunc(tc.pvc)
			if err != nil {
				t.Errorf("failed to extract pvc key from pvc %v", tc.pvc)
			}
			_, ok := ctrlInstance.uncertainPVCs[deletedPVCKey]
			if ok {
				t.Errorf("pvc %v should be deleted but it is still in the uncertainPVCs cache", tc.pvc)
			}
			if err != nil {
				t.Errorf("err get pvc %v from uncertainPVCs: %v", tc.pvc, err)
			}

			notDeletedPVCKey, err := cache.MetaNamespaceKeyFunc(secondPVC)
			if err != nil {
				t.Errorf("failed to extract pvc key from secondPVC %v", secondPVC)
			}
			_, ok = ctrlInstance.uncertainPVCs[notDeletedPVCKey]
			if !ok {
				t.Errorf("pvc %v should not be deleted, uncertainPVCs list %v", secondPVC, ctrlInstance.uncertainPVCs)
			}
			if err != nil {
				t.Errorf("err get pvc %v from uncertainPVCs: %v", secondPVC, err)
			}
		})
	}
}

func createTestPV(capacityGB int, pvcName, pvcNamespace string, pvcUID types.UID, volumeMode *v1.PersistentVolumeMode, vacName string) *v1.PersistentVolume {
	capacity := testutil.QuantityGB(capacityGB)

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testPV",
		},
		Spec: v1.PersistentVolumeSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Capacity: map[v1.ResourceName]resource.Quantity{
				v1.ResourceStorage: capacity,
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				CSI: &v1.CSIPersistentVolumeSource{
					Driver:       testDriverName,
					VolumeHandle: "foo",
				},
			},
			VolumeMode:                volumeMode,
			VolumeAttributesClassName: &vacName,
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

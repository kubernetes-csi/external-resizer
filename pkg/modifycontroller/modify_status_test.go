package modifycontroller

import (
	"context"
	"reflect"
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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
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
	finalErr               = status.Errorf(codes.Internal, "Final error")
	nonFinalErr            = status.Errorf(codes.Aborted, "Non-final error")
	pvcConditionInProgress = v1.PersistentVolumeClaimCondition{
		Type:    v1.PersistentVolumeClaimVolumeModifyingVolume,
		Status:  v1.ConditionTrue,
		Message: "ModifyVolume operation in progress.",
	}

	pvcConditionInfeasible = v1.PersistentVolumeClaimCondition{
		Type:    v1.PersistentVolumeClaimVolumeModifyVolumeError,
		Status:  v1.ConditionTrue,
		Message: "ModifyVolume failed with error: rpc error: code = InvalidArgument desc = Parameters in VolumeAttributesClass is invalid. Waiting for retry.",
	}

	pvcConditionError = v1.PersistentVolumeClaimCondition{
		Type:    v1.PersistentVolumeClaimVolumeModifyVolumeError,
		Status:  v1.ConditionTrue,
		Message: "ModifyVolume failed with error: rpc error: code = Internal desc = Final error. Waiting for retry.",
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
			name:               "mark modify volume as failed",
			pvc:                basePVC.Get(),
			expectedPVC:        basePVC.WithModifyVolumeStatus(v1.PersistentVolumeClaimModifyVolumeInProgress).Get(),
			expectedConditions: []v1.PersistentVolumeClaimCondition{pvcConditionError},
			expectedErr:        nil,
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *modifyController) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markControllerModifyVolumeStatus(pvc, v1.PersistentVolumeClaimModifyVolumeInProgress, finalErr)
			},
		},
		{
			name:               "mark modify volume as infeasible",
			pvc:                basePVC.Get(),
			expectedPVC:        basePVC.WithModifyVolumeStatus(v1.PersistentVolumeClaimModifyVolumeInfeasible).Get(),
			expectedConditions: []v1.PersistentVolumeClaimCondition{pvcConditionInfeasible},
			expectedErr:        infeasibleErr,
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
			if err != nil && !reflect.DeepEqual(tc.expectedErr, err) {
				t.Errorf("Expected error to be %v but got %v", tc.expectedErr, err)
			}

			realStatus := pvc.Status.ModifyVolumeStatus.Status
			expectedStatus := tc.expectedPVC.Status.ModifyVolumeStatus.Status
			if !reflect.DeepEqual(realStatus, expectedStatus) {
				t.Errorf("expected modify volume status %+v got %+v", expectedStatus, realStatus)
			}

			realConditions := pvc.Status.Conditions
			if !testutil.CompareConditions(realConditions, tc.expectedConditions) {
				t.Errorf("expected conditions %+v got %+v", tc.expectedConditions, realConditions)
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
			if err != nil && !reflect.DeepEqual(tc.expectedErr, err) {
				t.Errorf("Expected error to be %v but got %v", tc.expectedErr, err)
			}

			if len(actualPVC.Status.Conditions) == 0 {
				actualPVC.Status.Conditions = []v1.PersistentVolumeClaimCondition{}
			}

			if diff := cmp.Diff(tc.expectedPVC, actualPVC); diff != "" {
				t.Errorf("expected pvc %+v got %+v, diff is: %v", tc.expectedPVC, actualPVC, diff)
			}

			if diff := cmp.Diff(tc.expectedPV, pv); diff != "" {
				t.Errorf("expected pvc %+v got %+v, diff is: %v", tc.expectedPV, pv, diff)
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

package modifycontroller

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	"github.com/kubernetes-csi/external-resizer/pkg/features"
	"github.com/kubernetes-csi/external-resizer/pkg/modifier"
	"github.com/kubernetes-csi/external-resizer/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
)

func TestController(t *testing.T) {
	basePVC := createTestPVC(pvcName, testVac /*vacName*/, testVac /*curVacName*/, "" /*targetVacName*/)
	basePVC.Status.ModifyVolumeStatus = nil
	basePV := createTestPV(1, pvcName, pvcNamespace, "foobaz" /*pvcUID*/, &fsVolumeMode, testVac)
	firstTimePV := basePV.DeepCopy()
	firstTimePV.Spec.VolumeAttributesClassName = nil
	firstTimePVC := basePVC.DeepCopy()
	firstTimePVC.Status.CurrentVolumeAttributesClassName = nil

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
			client := csi.NewMockClient(testDriverName, false, false, true, true, true)

			initialObjects := []runtime.Object{test.pvc, test.pv, testVacObject, targetVacObject}
			ctrlInstance := setupFakeK8sEnvironment(t, client, initialObjects)

			_, _, err, _ := ctrlInstance.modify(test.pvc, test.pv)
			if err != nil {
				t.Fatalf("for %s: unexpected error: %v", test.name, err)
			}

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
			client := csi.NewMockClient(testDriverName, true, true, true, true, true)
			if test.modifyFailure {
				client.SetModifyError(fmt.Errorf("fake modification error"))
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

func TestSyncPVC(t *testing.T) {
	basePVC := createTestPVC(pvcName, targetVac /*vacName*/, testVac /*curVacName*/, testVac /*targetVacName*/)
	basePV := createTestPV(1, pvcName, pvcNamespace, "foobaz" /*pvcUID*/, &fsVolumeMode, testVac)

	otherDriverPV := createTestPV(1, pvcName, pvcNamespace, "foobaz" /*pvcUID*/, &fsVolumeMode, testVac)
	otherDriverPV.Spec.PersistentVolumeSource.CSI.Driver = "some-other-driver"

	unboundPVC := createTestPVC(pvcName, targetVac /*vacName*/, testVac /*curVacName*/, testVac /*targetVacName*/)
	unboundPVC.Status.Phase = v1.ClaimPending

	pvcWithUncreatedPV := createTestPVC(pvcName, targetVac /*vacName*/, testVac /*curVacName*/, testVac /*targetVacName*/)
	pvcWithUncreatedPV.Spec.VolumeName = ""

	inprogressPVC := createTestPVC(pvcName, "" /*vacName*/, "" /*curVacName*/, testVac /*targetVacName*/)
	inprogressPVC.Status.ModifyVolumeStatus.Status = v1.PersistentVolumeClaimModifyVolumeInProgress

	nonCSIPVC := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: pvcName, Namespace: pvcNamespace},
		Spec: v1.PersistentVolumeClaimSpec{
			VolumeAttributesClassName: &targetVac,
			VolumeName:                pvName,
		},
		Status: v1.PersistentVolumeClaimStatus{
			Phase: v1.ClaimBound,
		},
	}
	nonCSIPV := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: pvName,
		},
		Spec: v1.PersistentVolumeSpec{
			VolumeAttributesClassName: nil,
		},
	}

	tests := []struct {
		name          string
		pvc           *v1.PersistentVolumeClaim
		pv            *v1.PersistentVolume
		callCSIModify bool
	}{
		{
			name:          "Should execute ModifyVolume operation when PVC's VAC changes",
			pvc:           basePVC,
			pv:            basePV,
			callCSIModify: true,
		},
		{
			name:          "Should NOT modify when rollback to empty VACName",
			pvc:           createTestPVC(pvcName, "" /*vacName*/, "" /*curVacName*/, testVac /*targetVacName*/),
			pv:            basePV,
			callCSIModify: false,
		},
		{
			name:          "Should NOT modify if PVC managed by another CSI Driver",
			pvc:           basePVC,
			pv:            otherDriverPV,
			callCSIModify: false,
		},
		{
			name:          "Should execute ModifyVolume for InProgress target if PVC has empty Spec.VACName",
			pvc:           inprogressPVC,
			pv:            basePV,
			callCSIModify: true,
		},
		{
			name:          "Should NOT modify if PVC deleted",
			pvc:           nil,
			callCSIModify: false,
		},
		{
			name:          "Should NOT modify if PVC not in bound state",
			pvc:           unboundPVC,
			pv:            basePV,
			callCSIModify: false,
		},
		{
			name:          "Should NOT modify if PVC's PV not created yet",
			pvc:           pvcWithUncreatedPV,
			pv:            basePV,
			callCSIModify: false,
		},
		{
			name:          "Should NOT modify if PV wasn't provisioned by CSI driver",
			pvc:           nonCSIPVC,
			pv:            nonCSIPV,
			callCSIModify: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := csi.NewMockClient(testDriverName, true, true, true, true, true)

			initialObjects := []runtime.Object{testVacObject, targetVacObject}
			if test.pvc != nil {
				initialObjects = append(initialObjects, test.pvc)
			}
			if test.pv != nil {
				initialObjects = append(initialObjects, test.pv)
			}
			ctrlInstance := setupFakeK8sEnvironment(t, client, initialObjects)

			err := ctrlInstance.syncPVC(pvcNamespace + "/" + pvcName)
			if err != nil {
				t.Errorf("for %s, unexpected error: %v", test.name, err)
			}

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

// TestInfeasibleRetry tests that sidecar doesn't spam plugin upon infeasible error code (e.g. invalid VAC parameter)
func TestInfeasibleRetry(t *testing.T) {
	basePVC := createTestPVC(pvcName, targetVac /*vacName*/, testVac /*curVacName*/, testVac /*targetVacName*/)
	basePV := createTestPV(1, pvcName, pvcNamespace, "foobaz" /*pvcUID*/, &fsVolumeMode, testVac)

	tests := []struct {
		name                        string
		pvc                         *v1.PersistentVolumeClaim
		expectedModifyCallCount     int
		csiModifyError              error
		eventuallyRemoveFromSlowSet bool
	}{
		{
			name:                        "Should retry non-infeasible error normally",
			pvc:                         basePVC,
			expectedModifyCallCount:     2,
			csiModifyError:              status.Errorf(codes.Internal, "fake non-infeasible error"),
			eventuallyRemoveFromSlowSet: false,
		},
		{
			name:                        "Should NOT retry infeasible error normally",
			pvc:                         basePVC,
			expectedModifyCallCount:     1,
			csiModifyError:              status.Errorf(codes.InvalidArgument, "fake infeasible error"),
			eventuallyRemoveFromSlowSet: false,
		},
		{
			name:                        "Should EVENTUALLY retry infeasible error",
			pvc:                         basePVC,
			expectedModifyCallCount:     2,
			csiModifyError:              status.Errorf(codes.InvalidArgument, "fake infeasible error"),
			eventuallyRemoveFromSlowSet: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup
			client := csi.NewMockClient(testDriverName, true, true, true, true, true)
			if test.csiModifyError != nil {
				client.SetModifyError(test.csiModifyError)
			}

			initialObjects := []runtime.Object{test.pvc, basePV, testVacObject, targetVacObject}
			ctrlInstance := setupFakeK8sEnvironment(t, client, initialObjects)

			// Attempt modification first time
			err := ctrlInstance.syncPVC(pvcNamespace + "/" + pvcName)
			if !errors.Is(err, test.csiModifyError) {
				t.Errorf("for %s, unexpected first syncPVC error: %v", test.name, err)
			}

			// Fake time passing by removing from SlowSet
			if test.eventuallyRemoveFromSlowSet {
				pvcKey, _ := cache.MetaNamespaceKeyFunc(test.pvc)
				ctrlInstance.slowSet.Remove(pvcKey)
			}

			// Attempt modification second time
			err2 := ctrlInstance.syncPVC(pvcNamespace + "/" + pvcName)
			switch test.expectedModifyCallCount {
			case 1:
				if !util.IsDelayRetryError(err2) {
					t.Errorf("for %s, unexpected second syncPVC error: %v", test.name, err)
				}
			case 2:
				if !errors.Is(err2, test.csiModifyError) {
					t.Errorf("for %s, unexpected second syncPVC error: %v", test.name, err)
				}
			default:
				t.Errorf("for %s, unexpected second syncPVC error: %v", test.name, err)
			}

			// Confirm CSI ModifyVolume was called desired amount of times
			modifyCallCount := client.GetModifyCount()
			if test.expectedModifyCallCount != modifyCallCount {
				t.Fatalf("for %s: expected %d csi modify calls, but got %d", test.name, test.expectedModifyCallCount, modifyCallCount)
			}
		})
	}
}

// Intended to catch any race conditions in the controller
func TestConcurrentSync(t *testing.T) {
	cases := []struct {
		name      string
		waitCount int
		err       error
	}{
		// TODO: This case is flaky due to fake client lacks resourceVersion support.
		// {
		// 	name:      "success",
		// 	waitCount: 10,
		// },
		{
			name:      "uncertain",
			waitCount: 30,
			err:       nonFinalErr,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := csi.NewMockClient(testDriverName, true, true, true, true, true)
			client.SetModifyError(tc.err)

			initialObjects := []runtime.Object{testVacObject, targetVacObject}
			for i := range 10 {
				initialObjects = append(initialObjects,
					&v1.PersistentVolumeClaim{
						ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("foo-%d", i), Namespace: pvcNamespace},
						Spec: v1.PersistentVolumeClaimSpec{
							VolumeAttributesClassName: &testVac,
							VolumeName:                fmt.Sprintf("testPV-%d", i),
						},
						Status: v1.PersistentVolumeClaimStatus{
							Phase: v1.ClaimBound,
						},
					},
					&v1.PersistentVolume{
						ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("testPV-%d", i)},
						Spec: v1.PersistentVolumeSpec{
							PersistentVolumeSource: v1.PersistentVolumeSource{
								CSI: &v1.CSIPersistentVolumeSource{
									Driver:       testDriverName,
									VolumeHandle: fmt.Sprintf("foo-%d", i),
								},
							},
						},
					},
				)
			}
			ctrlInstance := setupFakeK8sEnvironment(t, client, initialObjects)
			wg := sync.WaitGroup{}
			t.Cleanup(wg.Wait)
			go ctrlInstance.Run(3, t.Context(), &wg)

			for client.GetModifyCount() < tc.waitCount {
				time.Sleep(20 * time.Millisecond)
			}
		})
	}
}

// setupFakeK8sEnvironment creates fake K8s environment and starts Informers and ModifyController
func setupFakeK8sEnvironment(t *testing.T, client *csi.MockClient, initialObjects []runtime.Object) *modifyController {
	t.Helper()

	featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.VolumeAttributesClass, true)

	/* Create fake kubeClient, Informers, and ModifyController */
	kubeClient, informerFactory := fakeK8s(initialObjects)

	ctx := t.Context()
	driverName, _ := client.GetDriverName(ctx)

	csiModifier, err := modifier.NewModifierFromClient(client, 15*time.Second, kubeClient, informerFactory, false, driverName)
	if err != nil {
		t.Fatalf("Test %s: Unable to create modifier: %v", t.Name(), err)
	}

	controller := NewModifyController(driverName,
		csiModifier, kubeClient,
		0 /* resyncPeriod */, 2*time.Minute, false, informerFactory,
		workqueue.DefaultTypedControllerRateLimiter[string]())

	/* Start informers and ModifyController*/
	informerFactory.Start(ctx.Done())

	ctrlInstance := controller.(*modifyController)
	ctrlInstance.init(ctx)

	return ctrlInstance
}

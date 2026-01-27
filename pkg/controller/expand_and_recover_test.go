package controller

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	"github.com/kubernetes-csi/external-resizer/pkg/features"
	"github.com/kubernetes-csi/external-resizer/pkg/resizer"
	"github.com/kubernetes-csi/external-resizer/pkg/testutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
)

func TestExpandAndRecover(t *testing.T) {
	fsVolumeMode := v1.PersistentVolumeFilesystem
	var tests = []struct {
		name                       string
		pvc                        *v1.PersistentVolumeClaim
		pv                         *v1.PersistentVolume
		disableNodeExpansion       bool
		disableControllerExpansion bool
		// expectations of test
		expectedResizeStatus  v1.ClaimResourceStatus
		expectedAllocatedSize resource.Quantity
		pvcWithFinalErrors    sets.Set[string]
		expansionError        error
		expectResizeCall      bool
		expectedConditions    []v1.PersistentVolumeClaimConditionType
	}{
		{
			name:                  "pvc.spec.size > pv.spec.size, resize_status=node_expansion_inprogress",
			pvc:                   testutil.GetTestPVC("test-vol0", "2G", "1G", "", ""),
			pv:                    createPV(1, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			expectedResizeStatus:  v1.PersistentVolumeClaimNodeResizePending,
			expectedAllocatedSize: resource.MustParse("2G"),
			expectResizeCall:      true,
		},
		{
			name:                  "pvc.spec.size = pv.spec.size, resize_status=no_expansion_inprogress",
			pvc:                   testutil.GetTestPVC("test-vol0", "1G", "1G", "", ""),
			pv:                    createPV(1, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			expectedResizeStatus:  v1.PersistentVolumeClaimNodeResizePending,
			expectedAllocatedSize: resource.MustParse("1G"),
			expectResizeCall:      true,
		},
		{
			name:                  "pvc.spec.size > pv.spec.size, resize_status=controller_expansion_infeasible",
			pvc:                   testutil.GetTestPVC("test-vol0", "5G" /*specSize*/, "3G" /*statusSize*/, "10G" /*allocatedSize*/, v1.PersistentVolumeClaimControllerResizeInfeasible),
			pv:                    createPV(3, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			expectedResizeStatus:  v1.PersistentVolumeClaimNodeResizePending,
			expectedAllocatedSize: resource.MustParse("5G"),
			expectResizeCall:      true,
		},
		{
			name:                  "pvc.spec.size > pv.spec.size, resize_status=controller_expansion_fail(final, but feasible)",
			pvc:                   testutil.GetTestPVC("test-vol0", "5G" /*specSize*/, "3G" /*statusSize*/, "10G" /*allocatedSize*/, v1.PersistentVolumeClaimControllerResizeInProgress),
			pv:                    createPV(3, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			pvcWithFinalErrors:    sets.New(testutil.GetObjectKey("claim01")),
			expectedResizeStatus:  v1.PersistentVolumeClaimNodeResizePending,
			expectedAllocatedSize: resource.MustParse("5G"),
			expectResizeCall:      true,
		},
		{
			name:                  "pvc.spec.size > pv.spec.size, resize_status=controller_expansion_fail(final, but feasible), retry fail with final error",
			pvc:                   testutil.GetTestPVC("test-vol0", "5G" /*specSize*/, "3G" /*statusSize*/, "10G" /*allocatedSize*/, v1.PersistentVolumeClaimControllerResizeInProgress),
			pv:                    createPV(3, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			pvcWithFinalErrors:    sets.New(testutil.GetObjectKey("claim01")),
			expansionError:        status.Errorf(codes.FailedPrecondition, "something broke"),
			expectedResizeStatus:  v1.PersistentVolumeClaimControllerResizeInProgress,
			expectedAllocatedSize: resource.MustParse("5G"),
			expectResizeCall:      true,
		},
		{
			name:                  "pvc.spec.size > pv.spec.size, resize_status=controller_expansion_fail(final, but feasible), retry fail with infeasible error",
			pvc:                   testutil.GetTestPVC("test-vol0", "5G" /*specSize*/, "3G" /*statusSize*/, "10G" /*allocatedSize*/, v1.PersistentVolumeClaimControllerResizeInProgress),
			pv:                    createPV(3, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			pvcWithFinalErrors:    sets.New(testutil.GetObjectKey("claim01")),
			expansionError:        status.Errorf(codes.InvalidArgument, "something broke"),
			expectedResizeStatus:  v1.PersistentVolumeClaimControllerResizeInfeasible,
			expectedAllocatedSize: resource.MustParse("5G"),
			expectResizeCall:      true,
		},
		{
			name:                  "pvc.spec.size > pv.spec.size, resize_status=controller_expansion_inprogress",
			pvc:                   testutil.GetTestPVC("test-vol0", "5G" /*specSize*/, "3G" /*statusSize*/, "10G" /*allocatedSize*/, v1.PersistentVolumeClaimControllerResizeInProgress),
			pv:                    createPV(3, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			expectedResizeStatus:  v1.PersistentVolumeClaimNodeResizePending,
			expectedAllocatedSize: resource.MustParse("10G"),
			expectResizeCall:      true,
		},
		{
			name:                       "pvc.spec.size = pv.spec.size, resize_status=node_expansion_infeasible, disable_controller_expansion=true",
			pvc:                        testutil.GetTestPVC("test-vol0", "5G" /*specSize*/, "3G" /*statusSize*/, "10G" /*allocatedSize*/, v1.PersistentVolumeClaimNodeResizeInfeasible),
			pv:                         createPV(10, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			disableControllerExpansion: true,
			expectedResizeStatus:       v1.PersistentVolumeClaimNodeResizePending,
			expectedAllocatedSize:      resource.MustParse("5G"),
			expectResizeCall:           true,
		},
		{
			name:                  "pvc.spec.size = pv.spec.size, resize_status=node_expansion_pending",
			pvc:                   testutil.GetTestPVC("test-vol0", "1G", "1G", "1G", v1.PersistentVolumeClaimNodeResizePending),
			pv:                    createPV(1, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			expectedResizeStatus:  v1.PersistentVolumeClaimNodeResizePending,
			expectedAllocatedSize: resource.MustParse("1G"),
			expectResizeCall:      false,
		},
		{
			name:                  "pvc.spec.size > pv.spec.size, disable_node_expansion=true, resize_status=no_expansion_inprogress",
			pvc:                   testutil.GetTestPVC("test-vol0", "2G", "1G", "", ""),
			pv:                    createPV(1, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			disableNodeExpansion:  true,
			expectedResizeStatus:  "",
			expectedAllocatedSize: resource.MustParse("2G"),
			expectResizeCall:      true,
		},
		{
			name:                  "pv.spec.size >= pvc.spec.size, resize_status=node_expansion_failed",
			pvc:                   testutil.GetTestPVC("test-vol0", "2G", "1G", "2G", v1.PersistentVolumeClaimNodeResizeInfeasible),
			pv:                    createPV(2, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			expectedResizeStatus:  v1.PersistentVolumeClaimNodeResizePending,
			expectedAllocatedSize: resource.MustParse("2G"),
			expectResizeCall:      true,
		},
		{
			name:                  "pvc.spec.size > pv.spec.size, resize_status-node_expansion_pending",
			pvc:                   testutil.GetTestPVC("test-vol0", "10G", "1G", "3G", v1.PersistentVolumeClaimNodeResizePending),
			pv:                    createPV(3, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			expectedResizeStatus:  v1.PersistentVolumeClaimNodeResizePending,
			expectedAllocatedSize: resource.MustParse("3G"),
			expectResizeCall:      false,
		},
		{
			name:                  "pvc.spec.size > pv.spec.size, resize_status-node_expansion_inprogress",
			pvc:                   testutil.GetTestPVC("test-vol0", "10G", "1G", "3G", v1.PersistentVolumeClaimNodeResizeInProgress),
			pv:                    createPV(3, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			expectedResizeStatus:  v1.PersistentVolumeClaimNodeResizeInProgress,
			expectedAllocatedSize: resource.MustParse("3G"),
			expectResizeCall:      false,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.RecoverVolumeExpansionFailure, true)
			client := csi.NewMockClient("foo", !test.disableNodeExpansion, !test.disableControllerExpansion, false, true, true, false)
			driverName, _ := client.GetDriverName(context.TODO())
			if test.expansionError != nil {
				client.SetExpansionError(test.expansionError)
			}

			var initialObjects []runtime.Object
			initialObjects = append(initialObjects, test.pvc)
			initialObjects = append(initialObjects, test.pv)

			kubeClient, informerFactory := fakeK8s(initialObjects)

			csiResizer, err := resizer.NewResizerFromClient(client, 15*time.Second, kubeClient, driverName)
			if err != nil {
				t.Fatalf("Test %s: Unable to create resizer: %v", test.name, err)
			}

			controller := NewResizeController(driverName,
				csiResizer, kubeClient,
				time.Second, informerFactory,
				workqueue.DefaultTypedControllerRateLimiter[string](), true /*handleVolumeInUseError*/, 2*time.Minute /*maxRetryInterval*/)

			ctrlInstance, _ := controller.(*resizeController)
			recorder := record.NewFakeRecorder(10)
			ctrlInstance.eventRecorder = recorder
			ctrlInstance.finalErrorPVCs = test.pvcWithFinalErrors
			pvc, _, err, resizeCalled := ctrlInstance.expandAndRecover(test.pvc, test.pv)
			if test.expansionError == nil && err != nil {
				t.Fatalf("expansion failed with %v", err)
			}
			if test.expectResizeCall != resizeCalled {
				t.Fatalf("expansion failed: expected resize called %t, got %t", test.expectResizeCall, resizeCalled)
			}

			actualResizeStatus := pvc.Status.AllocatedResourceStatuses[v1.ResourceStorage]

			if actualResizeStatus != test.expectedResizeStatus {
				t.Fatalf("expected resize status to be %s, got %s", test.expectedResizeStatus, actualResizeStatus)
			}

			actualAllocatedSize := pvc.Status.AllocatedResources.Storage()

			if test.expectedAllocatedSize.Cmp(*actualAllocatedSize) != 0 {
				t.Fatalf("expansion failed: expected allocated size %s, got %s", test.expectedAllocatedSize.String(), actualAllocatedSize.String())
			}

		})
	}
}

// TestExpandAndRecoverConcurrent verifies that concurrent calls to expandAndRecover
// are thread-safe. This test exercises the synchronization of finalErrorPVCs access
// when multiple workers process failing PVC resize operations simultaneously.
func TestExpandAndRecoverConcurrent(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.RecoverVolumeExpansionFailure, true)

	fsVolumeMode := v1.PersistentVolumeFilesystem
	client := csi.NewMockClient("mock-driver", true, true, false, true, true, false)
	driverName, _ := client.GetDriverName(context.TODO())

	// Set expansion to fail with a final error - this triggers addFinalError()
	client.SetExpansionError(status.Errorf(codes.Internal, "simulated volume not found error"))

	numPVCs := 20
	pvcs := make([]*v1.PersistentVolumeClaim, numPVCs)
	pvs := make([]*v1.PersistentVolume, numPVCs)

	var initialObjects []runtime.Object
	for i := 0; i < numPVCs; i++ {
		pvcName := fmt.Sprintf("test-pvc-%d", i)
		pvName := fmt.Sprintf("test-pv-%d", i)
		pvcs[i] = createConcurrentTestPVC(pvcName, "2Gi", "1Gi")
		pvcs[i].Spec.VolumeName = pvName
		pvs[i] = createConcurrentTestPV(pvName, 1, pvcName, defaultNS, types.UID(pvcName+"-uid"), &fsVolumeMode)
		initialObjects = append(initialObjects, pvcs[i], pvs[i])
	}

	kubeClient, informerFactory := fakeK8s(initialObjects)

	csiResizer, err := resizer.NewResizerFromClient(client, 15*time.Second, kubeClient, driverName)
	if err != nil {
		t.Fatalf("Unable to create resizer: %v", err)
	}

	controller := NewResizeController(driverName,
		csiResizer, kubeClient,
		time.Second, informerFactory,
		workqueue.DefaultTypedControllerRateLimiter[string](), true, 2*time.Minute)

	ctrlInstance := controller.(*resizeController)
	ctrlInstance.eventRecorder = record.NewFakeRecorder(1000)

	var wg sync.WaitGroup
	numWorkers := 100

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			pvcIndex := workerID % numPVCs
			// This exercises the thread-safe finalErrorPVCs access
			_, _, _, _ = ctrlInstance.expandAndRecover(pvcs[pvcIndex], pvs[pvcIndex])
		}(i)
	}

	wg.Wait()
	t.Log("Concurrent expandAndRecover completed successfully")
}

// createConcurrentTestPVC creates a PVC for concurrent testing
func createConcurrentTestPVC(name string, specSize, statusSize string) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: defaultNS,
			UID:       types.UID(name + "-uid"),
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources:   v1.VolumeResourceRequirements{Requests: v1.ResourceList{v1.ResourceStorage: resource.MustParse(specSize)}},
			VolumeName:  name + "-pv",
		},
		Status: v1.PersistentVolumeClaimStatus{
			Phase:    v1.ClaimBound,
			Capacity: v1.ResourceList{v1.ResourceStorage: resource.MustParse(statusSize)},
		},
	}
}

// createConcurrentTestPV creates a PV for concurrent testing
func createConcurrentTestPV(name string, capacityGB int, pvcName, pvcNamespace string, pvcUID types.UID, volumeMode *v1.PersistentVolumeMode) *v1.PersistentVolume {
	capacity := resource.MustParse(fmt.Sprintf("%dGi", capacityGB))

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: make(map[string]string),
		},
		Spec: v1.PersistentVolumeSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Capacity: map[v1.ResourceName]resource.Quantity{
				v1.ResourceStorage: capacity,
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				CSI: &v1.CSIPersistentVolumeSource{
					Driver:       "mock-driver",
					VolumeHandle: name,
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

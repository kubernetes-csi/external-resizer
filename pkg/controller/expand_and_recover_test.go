package controller

import (
	"context"
	"testing"
	"time"

	"github.com/kubernetes-csi/external-resizer/v2/pkg/csi"
	"github.com/kubernetes-csi/external-resizer/v2/pkg/features"
	"github.com/kubernetes-csi/external-resizer/v2/pkg/resizer"
	"github.com/kubernetes-csi/external-resizer/v2/pkg/testutil"
	"github.com/kubernetes-csi/external-resizer/v2/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
		expectedResizeStatus                     v1.ClaimResourceStatus
		expectedAllocatedSize                    resource.Quantity
		expectNodeExpansionNotRequiredAnnotation bool
		pvcWithFinalErrors                       sets.Set[string]
		expansionError                           error
		expectResizeCall                         bool
		expectedConditions                       []v1.PersistentVolumeClaimConditionType
	}{
		{
			name:                                     "pvc.spec.size > pv.spec.size, resize_status=node_expansion_inprogress",
			pvc:                                      testutil.GetTestPVC("test-vol0", "2G", "1G", "", ""),
			pv:                                       createPV(1, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			expectedResizeStatus:                     v1.PersistentVolumeClaimNodeResizePending,
			expectNodeExpansionNotRequiredAnnotation: false,
			expectedAllocatedSize:                    resource.MustParse("2G"),
			expectResizeCall:                         true,
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
			name:                                     "pvc.spec.size > pv.spec.size, disable_node_expansion=true, resize_status=no_expansion_inprogress",
			pvc:                                      testutil.GetTestPVC("test-vol0", "2G", "1G", "", ""),
			pv:                                       createPV(1, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			disableNodeExpansion:                     true,
			expectedResizeStatus:                     "",
			expectNodeExpansionNotRequiredAnnotation: true,
			expectedAllocatedSize:                    resource.MustParse("2G"),
			expectResizeCall:                         true,
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
			client := csi.NewMockClient("foo", !test.disableNodeExpansion, !test.disableControllerExpansion, false, true, true)
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

			if test.expectNodeExpansionNotRequiredAnnotation != metav1.HasAnnotation(pvc.ObjectMeta, util.NodeExpansionNotRequired) {
				t.Fatalf("expected node expansion not required annotation to be %t, got %t", test.expectNodeExpansionNotRequiredAnnotation, metav1.HasAnnotation(pvc.ObjectMeta, util.NodeExpansionNotRequired))
			}

			actualAllocatedSize := pvc.Status.AllocatedResources.Storage()

			if test.expectedAllocatedSize.Cmp(*actualAllocatedSize) != 0 {
				t.Fatalf("expansion failed: expected allocated size %s, got %s", test.expectedAllocatedSize.String(), actualAllocatedSize.String())
			}

		})
	}
}

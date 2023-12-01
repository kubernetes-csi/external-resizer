package controller

import (
	"context"
	"testing"
	"time"

	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	"github.com/kubernetes-csi/external-resizer/pkg/features"
	"github.com/kubernetes-csi/external-resizer/pkg/resizer"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
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
		expectResizeCall      bool
	}{
		{
			name:                  "pvc.spec.size > pv.spec.size, resize_status=node_expansion_inprogress",
			pvc:                   getTestPVC("test-vol0", "2G", "1G", "", ""),
			pv:                    createPV(1, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			expectedResizeStatus:  v1.PersistentVolumeClaimNodeResizePending,
			expectedAllocatedSize: resource.MustParse("2G"),
			expectResizeCall:      true,
		},
		{
			name:                  "pvc.spec.size = pv.spec.size, resize_status=no_expansion_inprogress",
			pvc:                   getTestPVC("test-vol0", "1G", "1G", "", ""),
			pv:                    createPV(1, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			expectedResizeStatus:  v1.PersistentVolumeClaimNodeResizePending,
			expectedAllocatedSize: resource.MustParse("1G"),
			expectResizeCall:      true,
		},
		{
			name:                  "pvc.spec.size > pv.spec.size, resize_status=controller_expansion_failed",
			pvc:                   getTestPVC("test-vol0", "5G" /*specSize*/, "3G" /*statusSize*/, "10G" /*allocatedSize*/, v1.PersistentVolumeClaimControllerResizeFailed),
			pv:                    createPV(3, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			expectedResizeStatus:  v1.PersistentVolumeClaimNodeResizePending,
			expectedAllocatedSize: resource.MustParse("5G"),
			expectResizeCall:      true,
		},
		{
			name:                       "pvc.spec.size = pv.spec.size, resize_status=node_expansion_failed, disable_controller_expansion=true",
			pvc:                        getTestPVC("test-vol0", "5G" /*specSize*/, "3G" /*statusSize*/, "10G" /*allocatedSize*/, v1.PersistentVolumeClaimNodeResizeFailed),
			pv:                         createPV(10, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			disableControllerExpansion: true,
			expectedResizeStatus:       v1.PersistentVolumeClaimNodeResizePending,
			expectedAllocatedSize:      resource.MustParse("5G"),
			expectResizeCall:           true,
		},
		{
			name:                  "pvc.spec.size = pv.spec.size, resize_status=node_expansion_pending",
			pvc:                   getTestPVC("test-vol0", "1G", "1G", "1G", v1.PersistentVolumeClaimNodeResizePending),
			pv:                    createPV(1, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			expectedResizeStatus:  v1.PersistentVolumeClaimNodeResizePending,
			expectedAllocatedSize: resource.MustParse("1G"),
			expectResizeCall:      false,
		},
		{
			name:                  "pvc.spec.size > pv.spec.size, disable_node_expansion=true, resize_status=no_expansion_inprogress",
			pvc:                   getTestPVC("test-vol0", "2G", "1G", "", ""),
			pv:                    createPV(1, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			disableNodeExpansion:  true,
			expectedResizeStatus:  "",
			expectedAllocatedSize: resource.MustParse("2G"),
			expectResizeCall:      true,
		},
		{
			name:                  "pv.spec.size >= pvc.spec.size, resize_status=node_expansion_failed",
			pvc:                   getTestPVC("test-vol0", "2G", "1G", "2G", v1.PersistentVolumeClaimNodeResizeFailed),
			pv:                    createPV(2, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			expectedResizeStatus:  v1.PersistentVolumeClaimNodeResizePending,
			expectedAllocatedSize: resource.MustParse("2G"),
			expectResizeCall:      true,
		},
		{
			name:                  "pvc.spec.size > pv.spec.size, resize_status-node_expansion_pending",
			pvc:                   getTestPVC("test-vol0", "10G", "1G", "3G", v1.PersistentVolumeClaimNodeResizePending),
			pv:                    createPV(3, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			expectedResizeStatus:  v1.PersistentVolumeClaimNodeResizePending,
			expectedAllocatedSize: resource.MustParse("3G"),
			expectResizeCall:      false,
		},
		{
			name:                  "pvc.spec.size > pv.spec.size, resize_status-node_expansion_inprogress",
			pvc:                   getTestPVC("test-vol0", "10G", "1G", "3G", v1.PersistentVolumeClaimNodeResizeInProgress),
			pv:                    createPV(3, "claim01", defaultNS, "test-uid", &fsVolumeMode),
			expectedResizeStatus:  v1.PersistentVolumeClaimNodeResizeInProgress,
			expectedAllocatedSize: resource.MustParse("3G"),
			expectResizeCall:      false,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.RecoverVolumeExpansionFailure, true)()
			client := csi.NewMockClient("foo", !test.disableNodeExpansion, !test.disableControllerExpansion, true, true)
			driverName, _ := client.GetDriverName(context.TODO())

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
				workqueue.DefaultControllerRateLimiter(), true /*handleVolumeInUseError*/)

			ctrlInstance, _ := controller.(*resizeController)
			pvc, _, err, resizeCalled := ctrlInstance.expandAndRecover(test.pvc, test.pv)
			if err != nil {
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

func getTestPVC(volumeName string, specSize, statusSize, allocatedSize string, resizeStatus v1.ClaimResourceStatus) *v1.PersistentVolumeClaim {
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "claim01",
			Namespace: defaultNS,
			UID:       "test-uid",
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources:   v1.ResourceRequirements{Requests: v1.ResourceList{v1.ResourceStorage: resource.MustParse(specSize)}},
			VolumeName:  volumeName,
		},
		Status: v1.PersistentVolumeClaimStatus{
			Phase: v1.ClaimBound,
		},
	}
	if len(statusSize) > 0 {
		pvc.Status.Capacity = v1.ResourceList{v1.ResourceStorage: resource.MustParse(statusSize)}
	}
	if len(allocatedSize) > 0 {
		pvc.Status.AllocatedResources = v1.ResourceList{v1.ResourceStorage: resource.MustParse(allocatedSize)}
	}
	if len(resizeStatus) > 0 {
		pvc.Status.AllocatedResourceStatuses = map[v1.ResourceName]v1.ClaimResourceStatus{
			v1.ResourceStorage: resizeStatus,
		}
	}
	return pvc
}

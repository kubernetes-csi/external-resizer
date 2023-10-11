package controller

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	"github.com/kubernetes-csi/external-resizer/pkg/features"
	"github.com/kubernetes-csi/external-resizer/pkg/resizer"
	"github.com/kubernetes-csi/external-resizer/pkg/testutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/util/workqueue"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
)

func TestResizeFunctions(t *testing.T) {
	basePVC := testutil.MakePVC([]v1.PersistentVolumeClaimCondition{})

	tests := []struct {
		name        string
		pvc         *v1.PersistentVolumeClaim
		expectedPVC *v1.PersistentVolumeClaim
		testFunc    func(*v1.PersistentVolumeClaim, *resizeController, resource.Quantity) (*v1.PersistentVolumeClaim, error)
	}{
		{
			name:        "mark fs resize, with no other conditions",
			pvc:         basePVC.Get(),
			expectedPVC: basePVC.WithStorageResourceStatus(v1.PersistentVolumeClaimNodeResizePending).Get(),
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *resizeController, size resource.Quantity) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markForPendingNodeExpansion(pvc)
			},
		},
		{
			name: "mark fs resize, when other resource statuses are present",
			pvc:  basePVC.WithResourceStatus(v1.ResourceCPU, v1.PersistentVolumeClaimControllerResizeFailed).Get(),
			expectedPVC: basePVC.WithResourceStatus(v1.ResourceCPU, v1.PersistentVolumeClaimControllerResizeFailed).
				WithStorageResourceStatus(v1.PersistentVolumeClaimNodeResizePending).Get(),
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *resizeController, _ resource.Quantity) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markForPendingNodeExpansion(pvc)
			},
		},
		{
			name:        "mark controller resize in-progress",
			pvc:         basePVC.Get(),
			expectedPVC: basePVC.WithStorageResourceStatus(v1.PersistentVolumeClaimControllerResizeInProgress).Get(),
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *resizeController, q resource.Quantity) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markControllerResizeInProgress(pvc, q)
			},
		},
		{
			name:        "mark controller resize failed",
			pvc:         basePVC.Get(),
			expectedPVC: basePVC.WithStorageResourceStatus(v1.PersistentVolumeClaimControllerResizeFailed).Get(),
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *resizeController, q resource.Quantity) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markControllerExpansionFailed(pvc)
			},
		},
		{
			name: "mark resize finished",
			pvc: basePVC.WithResourceStatus(v1.ResourceCPU, v1.PersistentVolumeClaimControllerResizeFailed).
				WithStorageResourceStatus(v1.PersistentVolumeClaimNodeResizePending).Get(),
			expectedPVC: basePVC.WithResourceStatus(v1.ResourceCPU, v1.PersistentVolumeClaimControllerResizeFailed).
				WithStorageResourceStatus("").Get(),
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *resizeController, q resource.Quantity) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markOverallExpansionAsFinished(pvc, q)
			},
		},
	}

	for _, test := range tests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.RecoverVolumeExpansionFailure, true)()
			client := csi.NewMockClient("foo", true, true, false, true, true)
			driverName, _ := client.GetDriverName(context.TODO())

			pvc := test.pvc

			var initialObjects []runtime.Object
			initialObjects = append(initialObjects, test.pvc)

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

			pvc, err = tc.testFunc(pvc, ctrlInstance, resource.MustParse("10Gi"))
			if err != nil {
				t.Errorf("Expected no error but got %v", err)
			}

			realStatus := pvc.Status.AllocatedResourceStatuses
			expectedStatus := tc.expectedPVC.Status.AllocatedResourceStatuses
			if !reflect.DeepEqual(realStatus, expectedStatus) {
				t.Errorf("expected %+v got %+v", expectedStatus, realStatus)
			}
		})
	}

}

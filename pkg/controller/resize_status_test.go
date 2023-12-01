package controller

import (
	"context"
	"reflect"
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

func TestResizeFunctions(t *testing.T) {
	basePVC := makePVC([]v1.PersistentVolumeClaimCondition{})

	tests := []struct {
		name        string
		pvc         *v1.PersistentVolumeClaim
		expectedPVC *v1.PersistentVolumeClaim
		testFunc    func(*v1.PersistentVolumeClaim, *resizeController, resource.Quantity) (*v1.PersistentVolumeClaim, error)
	}{
		{
			name:        "mark fs resize, with no other conditions",
			pvc:         basePVC.get(),
			expectedPVC: basePVC.withStorageResourceStatus(v1.PersistentVolumeClaimNodeResizePending).get(),
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *resizeController, size resource.Quantity) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markForPendingNodeExpansion(pvc)
			},
		},
		{
			name: "mark fs resize, when other resource statuses are present",
			pvc:  basePVC.withResourceStatus(v1.ResourceCPU, v1.PersistentVolumeClaimControllerResizeFailed).get(),
			expectedPVC: basePVC.withResourceStatus(v1.ResourceCPU, v1.PersistentVolumeClaimControllerResizeFailed).
				withStorageResourceStatus(v1.PersistentVolumeClaimNodeResizePending).get(),
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *resizeController, _ resource.Quantity) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markForPendingNodeExpansion(pvc)
			},
		},
		{
			name:        "mark controller resize in-progress",
			pvc:         basePVC.get(),
			expectedPVC: basePVC.withStorageResourceStatus(v1.PersistentVolumeClaimControllerResizeInProgress).get(),
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *resizeController, q resource.Quantity) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markControllerResizeInProgress(pvc, q)
			},
		},
		{
			name:        "mark controller resize failed",
			pvc:         basePVC.get(),
			expectedPVC: basePVC.withStorageResourceStatus(v1.PersistentVolumeClaimControllerResizeFailed).get(),
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *resizeController, q resource.Quantity) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markControllerExpansionFailed(pvc)
			},
		},
		{
			name: "mark resize finished",
			pvc: basePVC.withResourceStatus(v1.ResourceCPU, v1.PersistentVolumeClaimControllerResizeFailed).
				withStorageResourceStatus(v1.PersistentVolumeClaimNodeResizePending).get(),
			expectedPVC: basePVC.withResourceStatus(v1.ResourceCPU, v1.PersistentVolumeClaimControllerResizeFailed).
				withStorageResourceStatus("").get(),
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *resizeController, q resource.Quantity) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markOverallExpansionAsFinished(pvc, q)
			},
		},
	}

	for _, test := range tests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.RecoverVolumeExpansionFailure, true)()
			client := csi.NewMockClient("foo", true, true, true, true)
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

type pvcModifier struct {
	pvc *v1.PersistentVolumeClaim
}

func (m pvcModifier) get() *v1.PersistentVolumeClaim {
	return m.pvc.DeepCopy()
}

func makePVC(conditions []v1.PersistentVolumeClaimCondition) pvcModifier {
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "resize"},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
				v1.ReadOnlyMany,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("2Gi"),
				},
			},
		},
		Status: v1.PersistentVolumeClaimStatus{
			Phase:      v1.ClaimBound,
			Conditions: conditions,
			Capacity: v1.ResourceList{
				v1.ResourceStorage: resource.MustParse("2Gi"),
			},
		},
	}
	return pvcModifier{pvc}
}

func (m pvcModifier) withStorageResourceStatus(status v1.ClaimResourceStatus) pvcModifier {
	return m.withResourceStatus(v1.ResourceStorage, status)
}

func (m pvcModifier) withResourceStatus(resource v1.ResourceName, status v1.ClaimResourceStatus) pvcModifier {
	if m.pvc.Status.AllocatedResourceStatuses != nil && status == "" {
		delete(m.pvc.Status.AllocatedResourceStatuses, resource)
		return m
	}
	if m.pvc.Status.AllocatedResourceStatuses != nil {
		m.pvc.Status.AllocatedResourceStatuses[resource] = status
	} else {
		m.pvc.Status.AllocatedResourceStatuses = map[v1.ResourceName]v1.ClaimResourceStatus{
			resource: status,
		}
	}
	return m
}

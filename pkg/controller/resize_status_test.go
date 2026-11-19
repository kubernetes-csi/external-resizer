package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/kubernetes-csi/external-resizer/v2/pkg/csi"
	"github.com/kubernetes-csi/external-resizer/v2/pkg/resizer"
	"github.com/kubernetes-csi/external-resizer/v2/pkg/testutil"
	"github.com/kubernetes-csi/external-resizer/v2/pkg/util"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
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
			pvc:         basePVC().Get(),
			expectedPVC: basePVC().WithStorageResourceStatus(v1.PersistentVolumeClaimNodeResizePending).Get(),
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *resizeController, size resource.Quantity) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markForPendingNodeExpansion(pvc)
			},
		},
		{
			name: "mark fs resize, when other resource statuses are present",
			pvc:  basePVC().WithResourceStatus(v1.ResourceCPU, v1.PersistentVolumeClaimControllerResizeInfeasible).Get(),
			expectedPVC: basePVC().WithResourceStatus(v1.ResourceCPU, v1.PersistentVolumeClaimControllerResizeInfeasible).
				WithStorageResourceStatus(v1.PersistentVolumeClaimNodeResizePending).Get(),
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *resizeController, _ resource.Quantity) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markForPendingNodeExpansion(pvc)
			},
		},
		{
			name:        "mark controller resize in-progress",
			pvc:         basePVC().Get(),
			expectedPVC: basePVC().WithStorageResourceStatus(v1.PersistentVolumeClaimControllerResizeInProgress).Get(),
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *resizeController, q resource.Quantity) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markControllerResizeInProgress(pvc, q, true)
			},
		},
		{
			name:        "mark controller resize failed",
			pvc:         basePVC().Get(),
			expectedPVC: basePVC().WithStorageResourceStatus(v1.PersistentVolumeClaimControllerResizeInfeasible).Get(),
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *resizeController, q resource.Quantity) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markControllerExpansionInfeasible(pvc, fmt.Errorf("things failed"))
			},
		},
		{
			name: "mark resize finished",
			pvc: basePVC().WithResourceStatus(v1.ResourceCPU, v1.PersistentVolumeClaimControllerResizeInfeasible).
				WithStorageResourceStatus(v1.PersistentVolumeClaimNodeResizePending).Get(),
			expectedPVC: basePVC().WithResourceStatus(v1.ResourceCPU, v1.PersistentVolumeClaimControllerResizeInfeasible).
				WithStorageResourceStatus("").Get(),
			testFunc: func(pvc *v1.PersistentVolumeClaim, ctrl *resizeController, q resource.Quantity) (*v1.PersistentVolumeClaim, error) {
				return ctrl.markOverallExpansionAsFinished(pvc, q)
			},
		},
	}

	for _, test := range tests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
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
				workqueue.DefaultTypedControllerRateLimiter[string](),
				true, /*handleVolumeInUseError*/
				2*time.Minute /*maxRetryInterval*/)

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

func TestMarkOverallExpansionAsFinishedConflict(t *testing.T) {
	size := resource.MustParse("10Gi")
	unused := v1.PersistentVolumeClaimCondition{
		Type: v1.PersistentVolumeClaimUnused, Status: v1.ConditionTrue,
		Reason: "NoPodsUsingPVC", LastTransitionTime: metav1.NewTime(time.Unix(123, 0)),
	}

	for _, test := range []struct {
		name        string
		change      func(*v1.PersistentVolumeClaim)
		wantPatches int
	}{
		{
			name: "unused condition changed",
			change: func(pvc *v1.PersistentVolumeClaim) {
				pvc.Status.Conditions = append(pvc.Status.Conditions, unused)
			},
			wantPatches: 2,
		},
		{
			name: "request changed",
			change: func(pvc *v1.PersistentVolumeClaim) {
				pvc.Spec.Resources.Requests[v1.ResourceStorage] = resource.MustParse("20Gi")
			},
			wantPatches: 1,
		},
		{
			name: "claim replaced",
			change: func(pvc *v1.PersistentVolumeClaim) {
				pvc.UID = "replacement"
			},
			wantPatches: 1,
		},
		{
			name: "resize state advanced",
			change: func(pvc *v1.PersistentVolumeClaim) {
				pvc.Status.AllocatedResourceStatuses[v1.ResourceStorage] = v1.PersistentVolumeClaimNodeResizePending
			},
			wantPatches: 1,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			stale := testutil.MakePVC([]v1.PersistentVolumeClaimCondition{
				{Type: v1.PersistentVolumeClaimResizing, Status: v1.ConditionTrue},
			})().WithStorageResourceStatus(v1.PersistentVolumeClaimControllerResizeInProgress).Get()
			stale.Spec.Resources.Requests[v1.ResourceStorage] = size
			stale.UID = "initial"
			stale.ResourceVersion = "1"
			current := stale.DeepCopy()
			current.ResourceVersion = "2"
			test.change(current)

			client := fake.NewSimpleClientset(current)
			patches := 0
			client.PrependReactor("patch", "persistentvolumeclaims", func(action clienttesting.Action) (bool, runtime.Object, error) {
				patches++
				var patch struct {
					Metadata metav1.ObjectMeta
				}
				if err := json.Unmarshal(action.(clienttesting.PatchAction).GetPatch(), &patch); err != nil {
					t.Fatalf("invalid PVC patch: %v", err)
				}
				if patch.Metadata.ResourceVersion != current.ResourceVersion {
					return true, nil, apierrors.NewConflict(schema.GroupResource{Resource: "persistentvolumeclaims"}, stale.Name, fmt.Errorf("object has been modified"))
				}
				return false, nil, nil
			})
			events := record.NewFakeRecorder(1)
			ctrl := &resizeController{
				kubeClient:    client,
				claims:        cache.NewStore(cache.MetaNamespaceKeyFunc),
				eventRecorder: events,
			}

			result, err := ctrl.markOverallExpansionAsFinished(stale, size)
			if patches != test.wantPatches {
				t.Errorf("got %d PVC patches, want %d", patches, test.wantPatches)
			}
			if test.wantPatches == 1 {
				if !apierrors.IsConflict(err) {
					t.Fatalf("expected conflict, got %v", err)
				}
				if len(events.Events) != 0 {
					t.Fatal("unexpected success event after conflict")
				}
				return
			}
			if err != nil {
				t.Fatalf("finalizing resize: %v", err)
			}
			capacity := result.Status.Capacity[v1.ResourceStorage]
			if capacity.Cmp(size) != 0 || result.Status.AllocatedResourceStatuses[v1.ResourceStorage] != "" {
				t.Errorf("resize was not finalized: %+v", result.Status)
			}
			conditions := result.Status.Conditions
			if len(conditions) != 1 || conditions[0].Type != unused.Type ||
				conditions[0].Status != unused.Status || conditions[0].Reason != unused.Reason ||
				!conditions[0].LastTransitionTime.Equal(&unused.LastTransitionTime) {
				t.Errorf("expected only the concurrent Unused condition, got %+v", conditions)
			}
			if result.Annotations[util.NodeExpansionNotRequired] != "true" || len(events.Events) != 1 {
				t.Error("expected node expansion annotation and one success event")
			}
		})
	}
}

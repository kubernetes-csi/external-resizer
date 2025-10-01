package util

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	now = metav1.Now()
	old = metav1.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)

	pvcWithResizePendingCondition = v1.PersistentVolumeClaimCondition{
		Type:               v1.PersistentVolumeClaimFileSystemResizePending,
		Status:             v1.ConditionTrue,
		LastTransitionTime: now,
		Message:            "Waiting for user to (re-)start a pod to finish file system resize of volume on node.",
	}
	pvcWithControllerResizeError = v1.PersistentVolumeClaimCondition{
		Type:               v1.PersistentVolumeClaimControllerResizeError,
		Status:             v1.ConditionTrue,
		LastTransitionTime: now,
		Message:            "controller resize failed",
	}

	pvcWithControllerResizeError2 = v1.PersistentVolumeClaimCondition{
		Type:               v1.PersistentVolumeClaimControllerResizeError,
		Status:             v1.ConditionTrue,
		LastTransitionTime: now,
		Message:            "controller resize failed with error2",
	}

	pvcWithModifyVolumeProgressCondition = v1.PersistentVolumeClaimCondition{
		Type:          v1.PersistentVolumeClaimVolumeModifyingVolume,
		Status:        v1.ConditionTrue,
		LastProbeTime: now,
		Message:       "ModifyVolume operation in progress.",
	}

	pvcConditionInfeasible = v1.PersistentVolumeClaimCondition{
		Type:          v1.PersistentVolumeClaimVolumeModifyVolumeError,
		Status:        v1.ConditionTrue,
		LastProbeTime: now,
		Reason:        "InvalidParameter",
		Message:       "invalid parameter",
	}

	pvcConditionTimeout = v1.PersistentVolumeClaimCondition{
		Type:          v1.PersistentVolumeClaimVolumeModifyVolumeError,
		Status:        v1.ConditionTrue,
		LastProbeTime: now,
		Reason:        "DeadlineExceeded",
		Message:       "Progress: 10%",
	}
)

func TestGetPVCPatchData(t *testing.T) {
	for i, c := range []struct {
		OldPVC *v1.PersistentVolumeClaim
	}{
		{&v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "1"}}},
		{&v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "2"}}},
	} {
		newPVC := c.OldPVC.DeepCopy()
		newPVC.Status.Conditions = append(newPVC.Status.Conditions,
			v1.PersistentVolumeClaimCondition{Type: VolumeResizing, Status: v1.ConditionTrue})
		patchBytes, err := GetPVCPatchData(c.OldPVC, newPVC, true /*addResourceVersionCheck*/)
		if err != nil {
			t.Errorf("Case %d: Get patch data failed: %v", i, err)
		}

		var patchMap map[string]interface{}
		err = json.Unmarshal(patchBytes, &patchMap)
		if err != nil {
			t.Errorf("Case %d: unmarshalling json patch failed: %v", i, err)
		}

		metadata, exist := patchMap["metadata"].(map[string]interface{})
		if !exist {
			t.Errorf("Case %d: ResourceVersion should exist in patch data", i)
		}
		resourceVersion := metadata["resourceVersion"].(string)
		if resourceVersion != c.OldPVC.ResourceVersion {
			t.Errorf("Case %d: ResourceVersion should be %s, got %s",
				i, c.OldPVC.ResourceVersion, resourceVersion)
		}
	}
}

func TestMergeResizeConditionsOfPVC(t *testing.T) {
	tests := []struct {
		name                   string
		oldConditions          []v1.PersistentVolumeClaimCondition
		newConditions          []v1.PersistentVolumeClaimCondition
		expectedConditions     []v1.PersistentVolumeClaimCondition
		keepOldResizeCondition bool
	}{
		{
			name:                   "should not remove previous non-resize conditions",
			oldConditions:          []v1.PersistentVolumeClaimCondition{pvcWithModifyVolumeProgressCondition},
			newConditions:          []v1.PersistentVolumeClaimCondition{pvcWithResizePendingCondition},
			expectedConditions:     []v1.PersistentVolumeClaimCondition{pvcWithModifyVolumeProgressCondition, pvcWithResizePendingCondition},
			keepOldResizeCondition: false,
		},
		{
			name:                   "should not remove previous resize conditions if requested",
			oldConditions:          []v1.PersistentVolumeClaimCondition{pvcWithModifyVolumeProgressCondition},
			newConditions:          []v1.PersistentVolumeClaimCondition{pvcWithControllerResizeError},
			expectedConditions:     []v1.PersistentVolumeClaimCondition{pvcWithModifyVolumeProgressCondition, pvcWithControllerResizeError},
			keepOldResizeCondition: true,
		},
		{
			name:                   "should not keep previous resize conditions of same type",
			oldConditions:          []v1.PersistentVolumeClaimCondition{pvcWithControllerResizeError},
			newConditions:          []v1.PersistentVolumeClaimCondition{pvcWithControllerResizeError2},
			expectedConditions:     []v1.PersistentVolumeClaimCondition{pvcWithControllerResizeError2},
			keepOldResizeCondition: true,
		},
	}

	for _, test := range tests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			resultConditions := MergeResizeConditionsOfPVC(tc.oldConditions, tc.newConditions, tc.keepOldResizeCondition)
			if !reflect.DeepEqual(resultConditions, tc.expectedConditions) {
				t.Errorf("expected conditions %+v got %+v", tc.expectedConditions, resultConditions)
			}
		})
	}
}

func transit(condition v1.PersistentVolumeClaimCondition, t metav1.Time) v1.PersistentVolumeClaimCondition {
	condition.LastTransitionTime = t
	return condition
}

func TestMergePVCConditions(t *testing.T) {
	tests := []struct {
		name               string
		oldConditions      []v1.PersistentVolumeClaimCondition
		newConditions      []v1.PersistentVolumeClaimCondition
		expectedConditions []v1.PersistentVolumeClaimCondition
	}{
		{
			name:          "merge new into irrelevant",
			oldConditions: []v1.PersistentVolumeClaimCondition{pvcWithResizePendingCondition},
			newConditions: []v1.PersistentVolumeClaimCondition{pvcWithModifyVolumeProgressCondition},
			expectedConditions: []v1.PersistentVolumeClaimCondition{
				pvcWithResizePendingCondition,
				transit(pvcWithModifyVolumeProgressCondition, now),
			},
		},
		{
			name: "no transition",
			oldConditions: []v1.PersistentVolumeClaimCondition{
				transit(pvcWithModifyVolumeProgressCondition, old),
				transit(pvcConditionInfeasible, old),
			},
			newConditions: []v1.PersistentVolumeClaimCondition{pvcConditionInfeasible},
			expectedConditions: []v1.PersistentVolumeClaimCondition{
				transit(pvcWithModifyVolumeProgressCondition, old),
				transit(pvcConditionInfeasible, old),
			},
		},
		{
			name: "transit",
			oldConditions: []v1.PersistentVolumeClaimCondition{
				transit(pvcWithModifyVolumeProgressCondition, old),
				transit(pvcConditionInfeasible, old),
			},
			newConditions: []v1.PersistentVolumeClaimCondition{pvcConditionTimeout},
			expectedConditions: []v1.PersistentVolumeClaimCondition{
				transit(pvcWithModifyVolumeProgressCondition, old),
				transit(pvcConditionTimeout, now),
			},
		},
		{
			name:               "merge empty",
			oldConditions:      []v1.PersistentVolumeClaimCondition{pvcWithModifyVolumeProgressCondition},
			newConditions:      []v1.PersistentVolumeClaimCondition{},
			expectedConditions: []v1.PersistentVolumeClaimCondition{pvcWithModifyVolumeProgressCondition},
		},
		{
			name:          "merge new into empty",
			oldConditions: []v1.PersistentVolumeClaimCondition{},
			newConditions: []v1.PersistentVolumeClaimCondition{
				pvcWithModifyVolumeProgressCondition,
				pvcConditionInfeasible,
			},
			expectedConditions: []v1.PersistentVolumeClaimCondition{
				transit(pvcWithModifyVolumeProgressCondition, now),
				transit(pvcConditionInfeasible, now),
			},
		},
		{
			name:               "noop",
			oldConditions:      []v1.PersistentVolumeClaimCondition{},
			newConditions:      []v1.PersistentVolumeClaimCondition{},
			expectedConditions: []v1.PersistentVolumeClaimCondition{},
		},
	}

	for _, test := range tests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			resultConditions := MergePVCConditions(tc.oldConditions, tc.newConditions)
			if diff := cmp.Diff(tc.expectedConditions, resultConditions); diff != "" {
				t.Errorf("conditions diff (-expected, +got): %v", diff)
			}
		})
	}
}

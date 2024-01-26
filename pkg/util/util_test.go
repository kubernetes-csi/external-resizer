package util

import (
	"encoding/json"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	pvcOldCondition = v1.PersistentVolumeClaimCondition{
		Type:               v1.PersistentVolumeClaimFileSystemResizePending,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            "Waiting for user to (re-)start a pod to finish file system resize of volume on node.",
	}
	pvcConditionInProgress = v1.PersistentVolumeClaimCondition{
		Type:               v1.PersistentVolumeClaimVolumeModifyingVolume,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            "ModifyVolume operation in progress.",
	}

	pvcConditionInfeasible = v1.PersistentVolumeClaimCondition{
		Type:               v1.PersistentVolumeClaimVolumeModifyingVolume,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            "ModifyVolume operation in progress.",
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

func TestMergeModifyVolumeConditionsOfPVC(t *testing.T) {
	tests := []struct {
		name               string
		oldConditions      []v1.PersistentVolumeClaimCondition
		newConditions      []v1.PersistentVolumeClaimCondition
		expectedConditions []v1.PersistentVolumeClaimCondition
	}{
		{
			name:               "merge new modify volume condition with old resize condition",
			oldConditions:      []v1.PersistentVolumeClaimCondition{pvcOldCondition},
			newConditions:      []v1.PersistentVolumeClaimCondition{pvcConditionInProgress},
			expectedConditions: []v1.PersistentVolumeClaimCondition{pvcOldCondition, pvcConditionInProgress},
		},
		{
			name:               "merge new modify volume condition with old modify volume condition",
			oldConditions:      []v1.PersistentVolumeClaimCondition{pvcConditionInProgress},
			newConditions:      []v1.PersistentVolumeClaimCondition{pvcConditionInfeasible},
			expectedConditions: []v1.PersistentVolumeClaimCondition{pvcConditionInfeasible},
		},
		{
			name:               "merge empty condition with old modify volume condition",
			oldConditions:      []v1.PersistentVolumeClaimCondition{pvcConditionInProgress},
			newConditions:      []v1.PersistentVolumeClaimCondition{},
			expectedConditions: []v1.PersistentVolumeClaimCondition{},
		},
		{
			name:               "merge new condition with old empty volume condition",
			oldConditions:      []v1.PersistentVolumeClaimCondition{},
			newConditions:      []v1.PersistentVolumeClaimCondition{pvcConditionInProgress},
			expectedConditions: []v1.PersistentVolumeClaimCondition{pvcConditionInProgress},
		},
	}

	for _, test := range tests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			resultConditions := MergeModifyVolumeConditionsOfPVC(tc.oldConditions, tc.newConditions)
			if !reflect.DeepEqual(resultConditions, tc.expectedConditions) {
				t.Errorf("expected conditions %+v got %+v", tc.expectedConditions, resultConditions)
			}
		})
	}
}

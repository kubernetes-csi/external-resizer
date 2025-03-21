package util

import (
	"encoding/json"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	pvcWithResizePendingCondition = v1.PersistentVolumeClaimCondition{
		Type:               v1.PersistentVolumeClaimFileSystemResizePending,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            "Waiting for user to (re-)start a pod to finish file system resize of volume on node.",
	}
	pvcWithControllerResizeError = v1.PersistentVolumeClaimCondition{
		Type:               v1.PersistentVolumeClaimControllerResizeError,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            "controller resize failed",
	}

	pvcWithControllerResizeError2 = v1.PersistentVolumeClaimCondition{
		Type:               v1.PersistentVolumeClaimControllerResizeError,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            "controller resize failed with error2",
	}

	pvcWithModifyVolumeProgressCondition = v1.PersistentVolumeClaimCondition{
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

func TestMergeModifyVolumeConditionsOfPVC(t *testing.T) {
	tests := []struct {
		name               string
		oldConditions      []v1.PersistentVolumeClaimCondition
		newConditions      []v1.PersistentVolumeClaimCondition
		expectedConditions []v1.PersistentVolumeClaimCondition
	}{
		{
			name:               "merge new modify volume condition with old resize condition",
			oldConditions:      []v1.PersistentVolumeClaimCondition{pvcWithResizePendingCondition},
			newConditions:      []v1.PersistentVolumeClaimCondition{pvcWithModifyVolumeProgressCondition},
			expectedConditions: []v1.PersistentVolumeClaimCondition{pvcWithResizePendingCondition, pvcWithModifyVolumeProgressCondition},
		},
		{
			name:               "merge new modify volume condition with old modify volume condition",
			oldConditions:      []v1.PersistentVolumeClaimCondition{pvcWithModifyVolumeProgressCondition},
			newConditions:      []v1.PersistentVolumeClaimCondition{pvcConditionInfeasible},
			expectedConditions: []v1.PersistentVolumeClaimCondition{pvcConditionInfeasible},
		},
		{
			name:               "merge empty condition with old modify volume condition",
			oldConditions:      []v1.PersistentVolumeClaimCondition{pvcWithModifyVolumeProgressCondition},
			newConditions:      []v1.PersistentVolumeClaimCondition{},
			expectedConditions: []v1.PersistentVolumeClaimCondition{},
		},
		{
			name:               "merge new condition with old empty volume condition",
			oldConditions:      []v1.PersistentVolumeClaimCondition{},
			newConditions:      []v1.PersistentVolumeClaimCondition{pvcWithModifyVolumeProgressCondition},
			expectedConditions: []v1.PersistentVolumeClaimCondition{pvcWithModifyVolumeProgressCondition},
		},
		{
			name:               "should not remove previous non-resize conditions",
			oldConditions:      []v1.PersistentVolumeClaimCondition{},
			newConditions:      []v1.PersistentVolumeClaimCondition{},
			expectedConditions: []v1.PersistentVolumeClaimCondition{},
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

func TestIsVacRolledBack(t *testing.T) {
	emptyString := ""
	originalVacName := "original-vac"
	targetVacName := "test"
	noRollbackVacName := "no-rollback-vac"
	tests := []struct {
		name           string
		pvc            *v1.PersistentVolumeClaim
		expectedOutput bool
	}{
		{
			name: "rollback to nil",
			pvc: &v1.PersistentVolumeClaim{
				Spec: v1.PersistentVolumeClaimSpec{
					VolumeAttributesClassName: nil,
				},
				Status: v1.PersistentVolumeClaimStatus{
					CurrentVolumeAttributesClassName: nil,
					ModifyVolumeStatus: &v1.ModifyVolumeStatus{
						Status:                          v1.PersistentVolumeClaimModifyVolumeInfeasible,
						TargetVolumeAttributesClassName: targetVacName,
					},
				},
			},
			expectedOutput: true,
		},
		{
			name: "rollback to empty string",
			pvc: &v1.PersistentVolumeClaim{
				Spec: v1.PersistentVolumeClaimSpec{
					VolumeAttributesClassName: &emptyString,
				},
				Status: v1.PersistentVolumeClaimStatus{
					CurrentVolumeAttributesClassName: nil,
					ModifyVolumeStatus: &v1.ModifyVolumeStatus{
						Status:                          v1.PersistentVolumeClaimModifyVolumeInfeasible,
						TargetVolumeAttributesClassName: targetVacName,
					},
				},
			},
			expectedOutput: true,
		},
		{
			name: "rollback from VAC B to VAC A",
			pvc: &v1.PersistentVolumeClaim{
				Spec: v1.PersistentVolumeClaimSpec{
					VolumeAttributesClassName: &originalVacName,
				},
				Status: v1.PersistentVolumeClaimStatus{
					CurrentVolumeAttributesClassName: &originalVacName,
					ModifyVolumeStatus: &v1.ModifyVolumeStatus{
						Status:                          v1.PersistentVolumeClaimModifyVolumeInfeasible,
						TargetVolumeAttributesClassName: targetVacName,
					},
				},
			},
			expectedOutput: true,
		},
		{
			name: "no rollback",
			pvc: &v1.PersistentVolumeClaim{
				Spec: v1.PersistentVolumeClaimSpec{
					VolumeAttributesClassName: &noRollbackVacName,
				},
				Status: v1.PersistentVolumeClaimStatus{
					CurrentVolumeAttributesClassName: &originalVacName,
					ModifyVolumeStatus: &v1.ModifyVolumeStatus{
						Status:                          v1.PersistentVolumeClaimModifyVolumeInfeasible,
						TargetVolumeAttributesClassName: targetVacName,
					},
				},
			},
			expectedOutput: false,
		},
	}

	for _, test := range tests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			actualOutput := IsVacRolledBack(tc.pvc)
			if actualOutput != tc.expectedOutput {
				t.Errorf("expected output %+v got %+v", tc.expectedOutput, actualOutput)
			}
		})
	}

}

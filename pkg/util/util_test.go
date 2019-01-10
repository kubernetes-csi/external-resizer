package util

import (
	"encoding/json"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		patchBytes, err := getPVCPatchData(c.OldPVC, newPVC)
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

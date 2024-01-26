package testutil

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	pvcName   = "foo"
	defaultNS = "default"
)

var (
	testVac   = "test-vac"
	targetVac = "target-vac"
)

func GetTestPVC(volumeName string, specSize, statusSize, allocatedSize string, resizeStatus v1.ClaimResourceStatus) *v1.PersistentVolumeClaim {
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "claim01",
			Namespace: defaultNS,
			UID:       "test-uid",
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources:   v1.VolumeResourceRequirements{Requests: v1.ResourceList{v1.ResourceStorage: resource.MustParse(specSize)}},
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

type pvcModifier struct {
	pvc *v1.PersistentVolumeClaim
}

func MakePVC(conditions []v1.PersistentVolumeClaimCondition) pvcModifier {
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "resize"},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
				v1.ReadOnlyMany,
			},
			Resources: v1.VolumeResourceRequirements{
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

func MakeTestPVC(conditions []v1.PersistentVolumeClaimCondition) pvcModifier {
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: pvcName, Namespace: "modify"},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
				v1.ReadOnlyMany,
			},
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("2Gi"),
				},
			},
			VolumeAttributesClassName: &targetVac,
		},
		Status: v1.PersistentVolumeClaimStatus{
			Phase:      v1.ClaimBound,
			Conditions: conditions,
			Capacity: v1.ResourceList{
				v1.ResourceStorage: resource.MustParse("2Gi"),
			},
			CurrentVolumeAttributesClassName: &testVac,
			ModifyVolumeStatus: &v1.ModifyVolumeStatus{
				TargetVolumeAttributesClassName: targetVac,
			},
		},
	}
	return pvcModifier{pvc}
}

func (m pvcModifier) WithModifyVolumeStatus(status v1.PersistentVolumeClaimModifyVolumeStatus) pvcModifier {
	if m.pvc.Status.ModifyVolumeStatus == nil {
		m.pvc.Status.ModifyVolumeStatus = &v1.ModifyVolumeStatus{}
	}
	m.pvc.Status.ModifyVolumeStatus.Status = status
	return m
}

func (m pvcModifier) WithCurrentVolumeAttributesClassName(currentVacName string) pvcModifier {
	if m.pvc.Status.ModifyVolumeStatus == nil {
		m.pvc.Status.ModifyVolumeStatus = &v1.ModifyVolumeStatus{}
	}
	m.pvc.Status.CurrentVolumeAttributesClassName = &currentVacName
	return m
}

func (m pvcModifier) WithConditions(conditions []v1.PersistentVolumeClaimCondition) pvcModifier {
	m.pvc.Status.Conditions = conditions
	return m
}

func CompareConditions(realConditions, expectedConditions []v1.PersistentVolumeClaimCondition) bool {
	if realConditions == nil && expectedConditions == nil {
		return true
	}
	if (realConditions == nil || expectedConditions == nil) || len(realConditions) != len(expectedConditions) {
		return false
	}

	for i, condition := range realConditions {
		if condition.Type != expectedConditions[i].Type || condition.Message != expectedConditions[i].Message || condition.Status != expectedConditions[i].Status {
			return false
		}
	}
	return true
}

func (m pvcModifier) Get() *v1.PersistentVolumeClaim {
	return m.pvc.DeepCopy()
}

func (m pvcModifier) WithStorageResourceStatus(status v1.ClaimResourceStatus) pvcModifier {
	return m.WithResourceStatus(v1.ResourceStorage, status)
}

func (m pvcModifier) WithResourceStatus(resource v1.ResourceName, status v1.ClaimResourceStatus) pvcModifier {
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

func QuantityGB(i int) resource.Quantity {
	q := resource.NewQuantity(int64(i*1024*1024*1024), resource.BinarySI)
	return *q
}

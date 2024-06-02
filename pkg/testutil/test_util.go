package testutil

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	pvcName   = "foo"
	defaultNS = "default"
)

var (
	testVac   = "test-vac"
	targetVac = "target-vac"
)

// PVCWrapper wraps a PVC inside.
type PVCWrapper struct {
	pvc *v1.PersistentVolumeClaim
}

// MakePVC builds a PVC wrapper.
func MakePVC(name string) *PVCWrapper {
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return &PVCWrapper{pvc}
}

// Get returns the inner PVC.
func (m *PVCWrapper) Get() *v1.PersistentVolumeClaim {
	return m.pvc.DeepCopy()
}

// WithNamespace sets namespace of inner PVC.
func (m *PVCWrapper) WithNamespace(namespace string) *PVCWrapper {
	m.pvc.ObjectMeta.Namespace = namespace
	return m
}

// WithUID sets UID of inner PVC.
func (m *PVCWrapper) WithUID(uid string) *PVCWrapper {
	m.pvc.ObjectMeta.UID = types.UID(uid)
	return m
}

// WithVolumeName sets `name` as .Spec.VolumeName of inner PVC.
func (m *PVCWrapper) WithVolumeName(name string) *PVCWrapper {
	m.pvc.Spec.VolumeName = name
	return m
}

// WithAccessModes sets `modes` as .Spec.AccessModes of inner PVC.
func (m *PVCWrapper) WithAccessModes(modes []v1.PersistentVolumeAccessMode) *PVCWrapper {
	m.pvc.Spec.AccessModes = modes
	return m
}

// WithRequest sets the resource storage request of inner PVC's Spec. Accepts strings like `2Gi`.
func (m *PVCWrapper) WithRequest(storage string) *PVCWrapper {
	m.pvc.Spec.Resources = v1.VolumeResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceStorage: resource.MustParse(storage),
		},
	}
	return m
}

// WithVolumeAttributesClassName sets `vacName` as .Spec.VolumeAttributeClass of inner PVC.
func (m *PVCWrapper) WithVolumeAttributesClassName(vacName string) *PVCWrapper {
	m.pvc.Spec.VolumeAttributesClassName = &vacName
	return m
}

// WithStatus sets status of inner PVC.
func (m *PVCWrapper) WithStatus(status v1.PersistentVolumeClaimStatus) *PVCWrapper {
	m.pvc.Status = status
	return m
}

// WithPhase sets `phase` as .status.Phase of inner PVC.
func (m *PVCWrapper) WithPhase(phase v1.PersistentVolumeClaimPhase) *PVCWrapper {
	m.pvc.Status.Phase = phase
	return m
}

// WithConditions sets `conditions` as .Status.Conditions of inner PVC's.
func (m *PVCWrapper) WithConditions(conditions []v1.PersistentVolumeClaimCondition) *PVCWrapper {
	m.pvc.Status.Conditions = conditions
	return m
}

// WithCapacity `capacity` of .Status.Capacity of inner PVC. Accepts strings like `2Gi`.
func (m *PVCWrapper) WithCapacity(capacity string) *PVCWrapper {
	m.pvc.Status.Capacity = v1.ResourceList{
		v1.ResourceStorage: resource.MustParse(capacity),
	}
	return m
}

// WithModifyVolumeStatus sets `status` as .Status.ModifyVolumeStatus.Status of inner PVC.
func (m *PVCWrapper) WithModifyVolumeStatus(status v1.PersistentVolumeClaimModifyVolumeStatus) *PVCWrapper {
	if m.pvc.Status.ModifyVolumeStatus == nil {
		m.pvc.Status.ModifyVolumeStatus = &v1.ModifyVolumeStatus{}
	}
	m.pvc.Status.ModifyVolumeStatus.Status = status
	return m
}

// WithTargetVolumeAttributeClassName sets `targetVacName` as .Status.ModifyVolumeStatus.TargetVolumeAttributesClassName
// of inner PVC.
func (m *PVCWrapper) WithTargetVolumeAttributeClassName(targetVacName string) *PVCWrapper {
	if m.pvc.Status.ModifyVolumeStatus == nil {
		m.pvc.Status.ModifyVolumeStatus = &v1.ModifyVolumeStatus{}
	}
	m.pvc.Status.ModifyVolumeStatus.TargetVolumeAttributesClassName = targetVacName
	return m
}

// WithCurrentVolumeAttributesClassName sets `currentVacName`
// as .Status.ModifyVolumeStatus.CurrentVolumeAttributesClassName of inner PVC.
func (m *PVCWrapper) WithCurrentVolumeAttributesClassName(currentVacName string) *PVCWrapper {
	if m.pvc.Status.ModifyVolumeStatus == nil {
		m.pvc.Status.ModifyVolumeStatus = &v1.ModifyVolumeStatus{}
	}
	m.pvc.Status.CurrentVolumeAttributesClassName = &currentVacName
	return m
}

// WithStorageResource adds `allocatedSize` as a `ResourceStorage` to .Status.AllocatedResources of inner PVC.
// Accepts strings like `2Gi`.
func (m *PVCWrapper) WithStorageResource(allocatedSize string) *PVCWrapper {
	return m.WithResource(v1.ResourceStorage, resource.MustParse(allocatedSize))
}

// WithResource adds `quantity` as .Status.AllocatedResources[`resource`] of inner PVC.
func (m *PVCWrapper) WithResource(resource v1.ResourceName, quantity resource.Quantity) *PVCWrapper {
	if m.pvc.Status.AllocatedResources != nil && &quantity == nil {
		delete(m.pvc.Status.AllocatedResources, resource)
		return m
	}
	if m.pvc.Status.AllocatedResources != nil {
		m.pvc.Status.AllocatedResources[resource] = quantity
	} else {
		m.pvc.Status.AllocatedResources = v1.ResourceList{v1.ResourceStorage: quantity}
	}
	return m
}

// WithStorageResourceStatus adds `status` as .Status.AllocatedResourceStatuses[v1.ResourceStorage] of inner PVC.
func (m *PVCWrapper) WithStorageResourceStatus(status v1.ClaimResourceStatus) *PVCWrapper {
	return m.WithResourceStatus(v1.ResourceStorage, status)
}

// WithResourceStatus adds `status` as .Status.AllocatedResourceStatuses[`resource`] of inner PVC.
func (m *PVCWrapper) WithResourceStatus(resource v1.ResourceName, status v1.ClaimResourceStatus) *PVCWrapper {
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

func OldMakePVC(conditions []v1.PersistentVolumeClaimCondition) pvcModifier {
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

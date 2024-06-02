package testutil

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
func (p *PVCWrapper) Get() *v1.PersistentVolumeClaim {
	return p.pvc.DeepCopy()
}

// WithName sets name of inner PVC.
func (p *PVCWrapper) WithName(name string) *PVCWrapper {
	p.pvc.ObjectMeta.Name = name
	return p
}

// WithNamespace sets namespace of inner PVC.
func (p *PVCWrapper) WithNamespace(namespace string) *PVCWrapper {
	p.pvc.ObjectMeta.Namespace = namespace
	return p
}

// WithUID sets UID of inner PVC.
func (p *PVCWrapper) WithUID(uid string) *PVCWrapper {
	p.pvc.ObjectMeta.UID = types.UID(uid)
	return p
}

// WithAnnotations sets `annotations` as .Annotations of inner PVC.
func (p *PVCWrapper) WithAnnotations(annotations map[string]string) *PVCWrapper {
	p.pvc.Annotations = annotations
	return p
}

// WithVolumeName sets `name` as .Spec.VolumeName of inner PVC.
func (p *PVCWrapper) WithVolumeName(name string) *PVCWrapper {
	p.pvc.Spec.VolumeName = name
	return p
}

// WithAccessModes sets `modes` as .Spec.AccessModes of inner PVC.
func (p *PVCWrapper) WithAccessModes(modes []v1.PersistentVolumeAccessMode) *PVCWrapper {
	p.pvc.Spec.AccessModes = modes
	return p
}

// WithRequest sets the resource storage request of inner PVC's Spec. Accepts strings like `2Gi`.
func (p *PVCWrapper) WithRequest(storage string) *PVCWrapper {
	p.pvc.Spec.Resources = v1.VolumeResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceStorage: resource.MustParse(storage),
		},
	}
	return p
}

// WithVolumeAttributesClassName sets `vacName` as .Spec.VolumeAttributeClass of inner PVC.
func (p *PVCWrapper) WithVolumeAttributesClassName(vacName string) *PVCWrapper {
	p.pvc.Spec.VolumeAttributesClassName = &vacName
	return p
}

// WithStatus sets status of inner PVC.
func (p *PVCWrapper) WithStatus(status v1.PersistentVolumeClaimStatus) *PVCWrapper {
	p.pvc.Status = status
	return p
}

// WithPhase sets `phase` as .status.Phase of inner PVC.
func (p *PVCWrapper) WithPhase(phase v1.PersistentVolumeClaimPhase) *PVCWrapper {
	p.pvc.Status.Phase = phase
	return p
}

// WithConditions sets `conditions` as .Status.Conditions of inner PVC's.
func (p *PVCWrapper) WithConditions(conditions []v1.PersistentVolumeClaimCondition) *PVCWrapper {
	p.pvc.Status.Conditions = conditions
	return p
}

// WithCapacity `capacity` of .Status.Capacity of inner PVC. Accepts strings like `2Gi`.
func (p *PVCWrapper) WithCapacity(capacity string) *PVCWrapper {
	p.pvc.Status.Capacity = v1.ResourceList{
		v1.ResourceStorage: resource.MustParse(capacity),
	}
	return p
}

// WithModifyVolumeStatus sets `status` as .Status.ModifyVolumeStatus.Status of inner PVC.
func (p *PVCWrapper) WithModifyVolumeStatus(status v1.PersistentVolumeClaimModifyVolumeStatus) *PVCWrapper {
	if p.pvc.Status.ModifyVolumeStatus == nil {
		p.pvc.Status.ModifyVolumeStatus = &v1.ModifyVolumeStatus{}
	}
	p.pvc.Status.ModifyVolumeStatus.Status = status
	return p
}

// WithTargetVolumeAttributeClassName sets `targetVacName` as .Status.ModifyVolumeStatus.TargetVolumeAttributesClassName
// of inner PVC.
func (p *PVCWrapper) WithTargetVolumeAttributeClassName(targetVacName string) *PVCWrapper {
	if p.pvc.Status.ModifyVolumeStatus == nil {
		p.pvc.Status.ModifyVolumeStatus = &v1.ModifyVolumeStatus{}
	}
	p.pvc.Status.ModifyVolumeStatus.TargetVolumeAttributesClassName = targetVacName
	return p
}

// WithCurrentVolumeAttributesClassName sets `currentVacName`
// as .Status.ModifyVolumeStatus.CurrentVolumeAttributesClassName of inner PVC.
func (p *PVCWrapper) WithCurrentVolumeAttributesClassName(currentVacName string) *PVCWrapper {
	if p.pvc.Status.ModifyVolumeStatus == nil {
		p.pvc.Status.ModifyVolumeStatus = &v1.ModifyVolumeStatus{}
	}
	p.pvc.Status.CurrentVolumeAttributesClassName = &currentVacName
	return p
}

// WithStorageResource adds `allocatedSize` as a `ResourceStorage` to .Status.AllocatedResources of inner PVC.
// Accepts strings like `2Gi`.
func (p *PVCWrapper) WithStorageResource(allocatedSize string) *PVCWrapper {
	return p.WithResource(v1.ResourceStorage, resource.MustParse(allocatedSize))
}

// WithResource adds `quantity` as .Status.AllocatedResources[`resource`] of inner PVC.
func (p *PVCWrapper) WithResource(resource v1.ResourceName, quantity resource.Quantity) *PVCWrapper {
	if p.pvc.Status.AllocatedResources != nil && &quantity == nil {
		delete(p.pvc.Status.AllocatedResources, resource)
		return p
	}
	if p.pvc.Status.AllocatedResources != nil {
		p.pvc.Status.AllocatedResources[resource] = quantity
	} else {
		p.pvc.Status.AllocatedResources = v1.ResourceList{v1.ResourceStorage: quantity}
	}
	return p
}

// WithStorageResourceStatus adds `status` as .Status.AllocatedResourceStatuses[v1.ResourceStorage] of inner PVC.
func (p *PVCWrapper) WithStorageResourceStatus(status v1.ClaimResourceStatus) *PVCWrapper {
	return p.WithResourceStatus(v1.ResourceStorage, status)
}

// WithResourceStatus adds `status` as .Status.AllocatedResourceStatuses[`resource`] of inner PVC.
func (p *PVCWrapper) WithResourceStatus(resource v1.ResourceName, status v1.ClaimResourceStatus) *PVCWrapper {
	if p.pvc.Status.AllocatedResourceStatuses != nil && status == "" {
		delete(p.pvc.Status.AllocatedResourceStatuses, resource)
		return p
	}
	if p.pvc.Status.AllocatedResourceStatuses != nil {
		p.pvc.Status.AllocatedResourceStatuses[resource] = status
	} else {
		p.pvc.Status.AllocatedResourceStatuses = map[v1.ResourceName]v1.ClaimResourceStatus{
			resource: status,
		}
	}
	return p
}

// CompareConditions returns true if `realConditions` and `expectedConditions` are equivalent.
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

func QuantityGB(i int) resource.Quantity {
	q := resource.NewQuantity(int64(i*1024*1024*1024), resource.BinarySI)
	return *q
}

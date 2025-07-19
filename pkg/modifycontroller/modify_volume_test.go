package modifycontroller

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	testVacObject = &storagev1.VolumeAttributesClass{
		ObjectMeta: metav1.ObjectMeta{Name: testVac},
		DriverName: testDriverName,
		Parameters: map[string]string{"iops": "3000"},
	}

	targetVacObject = &storagev1.VolumeAttributesClass{
		ObjectMeta: metav1.ObjectMeta{Name: targetVac},
		DriverName: testDriverName,
		Parameters: map[string]string{
			"iops":                             "4567",
			"csi.storage.k8s.io/pvc/name":      pvcName,
			"csi.storage.k8s.io/pvc/namespace": pvcNamespace,
			"csi.storage.k8s.io/pv/name":       pvName,
		},
	}
)

func TestModify(t *testing.T) {
	basePVC := createTestPVC(pvcName, testVac /*vacName*/, testVac /*curVacName*/, testVac /*targetVacName*/)
	basePV := createTestPV(1, pvcName, pvcNamespace, "foobaz" /*pvcUID*/, &fsVolumeMode, testVac)

	var tests = []struct {
		name                                     string
		pvc                                      *v1.PersistentVolumeClaim
		pv                                       *v1.PersistentVolume
		vacExists                                bool
		expectModifyCall                         bool
		expectedModifyVolumeStatus               *v1.ModifyVolumeStatus
		expectedCurrentVolumeAttributesClassName *string
		expectedPVVolumeAttributesClassName      *string
		withExtraMetadata                        bool
		expectedVacParams                        map[string]string
	}{
		{
			name:                                     "nothing to modify",
			pvc:                                      basePVC,
			pv:                                       basePV,
			expectModifyCall:                         false,
			expectedModifyVolumeStatus:               basePVC.Status.ModifyVolumeStatus,
			expectedCurrentVolumeAttributesClassName: &testVac,
			expectedPVVolumeAttributesClassName:      &testVac,
		},
		{
			name:             "vac does not exist, no modification and set ModifyVolumeStatus to pending",
			pvc:              createTestPVC(pvcName, targetVac /*vacName*/, testVac /*curVacName*/, "" /*targetVacName*/),
			pv:               basePV,
			expectModifyCall: false,
			expectedModifyVolumeStatus: &v1.ModifyVolumeStatus{
				TargetVolumeAttributesClassName: targetVac,
				Status:                          v1.PersistentVolumeClaimModifyVolumePending,
			},
			expectedCurrentVolumeAttributesClassName: &testVac,
			expectedPVVolumeAttributesClassName:      &testVac,
		},
		{
			name:                                     "modify volume success",
			pvc:                                      createTestPVC(pvcName, targetVac /*vacName*/, testVac /*curVacName*/, testVac /*targetVacName*/),
			pv:                                       basePV,
			vacExists:                                true,
			expectModifyCall:                         true,
			expectedModifyVolumeStatus:               nil,
			expectedCurrentVolumeAttributesClassName: &targetVac,
			expectedPVVolumeAttributesClassName:      &targetVac,
		},
		{
			name:                                     "modify volume success with extra metadata",
			pvc:                                      createTestPVC(pvcName, targetVac /*vacName*/, testVac /*curVacName*/, testVac /*targetVacName*/),
			pv:                                       basePV,
			vacExists:                                true,
			expectModifyCall:                         true,
			expectedModifyVolumeStatus:               nil,
			expectedCurrentVolumeAttributesClassName: &targetVac,
			expectedPVVolumeAttributesClassName:      &targetVac,
			withExtraMetadata:                        true,
			expectedVacParams: map[string]string{
				"iops":                             "4567",
				"csi.storage.k8s.io/pvc/name":      basePVC.GetName(),
				"csi.storage.k8s.io/pvc/namespace": basePVC.GetNamespace(),
				"csi.storage.k8s.io/pv/name":       "testPV",
			},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			// Setup
			client := csi.NewMockClient(testDriverName, true, true, true, true, true, test.withExtraMetadata)
			initialObjects := []runtime.Object{test.pvc, test.pv, testVacObject}
			if test.vacExists {
				initialObjects = append(initialObjects, targetVacObject)
			}
			ctrlInstance := setupFakeK8sEnvironment(t, client, initialObjects)

			// Action
			pvc, pv, err, modifyCalled := ctrlInstance.modify(test.pvc, test.pv)
			// Verify
			if err != nil {
				t.Errorf("modify failed with %v", err)
			}
			if test.expectModifyCall != modifyCalled {
				t.Errorf("modify volume failed: expected modify called %t, got %t", test.expectModifyCall, modifyCalled)
			}

			actualModifyVolumeStatus := pvc.Status.ModifyVolumeStatus

			if diff := cmp.Diff(test.expectedModifyVolumeStatus, actualModifyVolumeStatus); diff != "" {
				t.Errorf("expected modify volume status to be %v, got %v", test.expectedModifyVolumeStatus, actualModifyVolumeStatus)
			}

			actualCurrentVolumeAttributesClassName := pvc.Status.CurrentVolumeAttributesClassName

			if diff := cmp.Diff(*test.expectedCurrentVolumeAttributesClassName, *actualCurrentVolumeAttributesClassName); diff != "" {
				t.Errorf("expected CurrentVolumeAttributesClassName to be %v, got %v", *test.expectedCurrentVolumeAttributesClassName, *actualCurrentVolumeAttributesClassName)
			}

			actualPVVolumeAttributesClassName := pv.Spec.VolumeAttributesClassName
			if diff := cmp.Diff(*test.expectedPVVolumeAttributesClassName, *actualPVVolumeAttributesClassName); diff != "" {
				t.Errorf("expected VolumeAttributesClassName of pv to be %v, got %v", *test.expectedPVVolumeAttributesClassName, *actualPVVolumeAttributesClassName)
			}

			if test.withExtraMetadata {
				vacObj, err := ctrlInstance.vacLister.Get(*test.expectedPVVolumeAttributesClassName)
				if err != nil {
					t.Errorf("failed to get VAC: %v", err)
				} else {
					vacParams := vacObj.Parameters
					if diff := cmp.Diff(test.expectedVacParams, vacParams); diff != "" {
						t.Errorf("expected VAC parameters to be %v, got %v", test.expectedVacParams, vacParams)
					}
				}
			}
		})
	}
}

func TestModifyUncertain(t *testing.T) {
	basePVC := createTestPVC(pvcName, targetVac /*vacName*/, testVac /*curVacName*/, targetVac /*targetVacName*/)
	basePVC.Status.ModifyVolumeStatus.Status = v1.PersistentVolumeClaimModifyVolumeInProgress
	basePV := createTestPV(1, pvcName, pvcNamespace, "foobaz" /*pvcUID*/, &fsVolumeMode, testVac)

	client := csi.NewMockClient(testDriverName, true, true, true, true, true, false)
	initialObjects := []runtime.Object{testVacObject, targetVacObject, basePVC, basePV}
	ctrlInstance := setupFakeK8sEnvironment(t, client, initialObjects)

	pvcKey := fmt.Sprintf("%s/%s", pvcNamespace, pvcName)
	assertUncertain := func(uncertain bool) {
		t.Helper()
		_, ok := ctrlInstance.uncertainPVCs.Load(pvcKey)
		if ok != uncertain {
			t.Fatalf("expected uncertain state to be %v, got %v", uncertain, ok)
		}
	}

	// initialized to uncertain
	assertUncertain(true)

	client.SetModifyError(finalErr)
	pvc, pv, err, _ := ctrlInstance.modify(basePVC, basePV)
	if !errors.Is(err, finalErr) {
		t.Fatalf("expected error to be %v, got %v", finalErr, err)
	}
	// should clear uncertain state
	assertUncertain(false)

	client.SetModifyError(nonFinalErr)
	_, _, err, _ = ctrlInstance.modify(pvc, pv)
	if !errors.Is(err, nonFinalErr) {
		t.Fatalf("expected error to be %v, got %v", nonFinalErr, err)
	}
	// should enter uncertain state again
	assertUncertain(true)
}

func createTestPVC(pvcName string, vacName string, curVacName string, targetVacName string) *v1.PersistentVolumeClaim {
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: pvcName, Namespace: pvcNamespace},
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
			VolumeAttributesClassName: &vacName,
			VolumeName:                pvName,
		},
		Status: v1.PersistentVolumeClaimStatus{
			Phase: v1.ClaimBound,
			Capacity: v1.ResourceList{
				v1.ResourceStorage: resource.MustParse("2Gi"),
			},
			CurrentVolumeAttributesClassName: &curVacName,
			ModifyVolumeStatus: &v1.ModifyVolumeStatus{
				TargetVolumeAttributesClassName: targetVacName,
				Status:                          "",
			},
		},
	}
	return pvc
}

func fakeK8s(objs []runtime.Object) (kubernetes.Interface, informers.SharedInformerFactory) {
	client := fake.NewSimpleClientset(objs...)
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	return client, informerFactory
}

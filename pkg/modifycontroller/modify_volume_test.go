package modifycontroller

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	v1 "k8s.io/api/core/v1"
	storagev1beta1 "k8s.io/api/storage/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	testVacObject = &storagev1beta1.VolumeAttributesClass{
		ObjectMeta: metav1.ObjectMeta{Name: testVac},
		DriverName: testDriverName,
		Parameters: map[string]string{"iops": "3000"},
	}

	targetVacObject = &storagev1beta1.VolumeAttributesClass{
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
	basePVC := createTestPVC(pvcName, &testVac /*vacName*/, &testVac /*curVacName*/, testVac /*targetVacName*/, "" /*modifyVolumeStatus*/)
	basePV := createTestPV(1, pvcName, pvcNamespace, "foobaz" /*pvcUID*/, &fsVolumeMode, &testVac)

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
		addPVCtoUncertainCache                   bool
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
			pvc:              createTestPVC(pvcName, &targetVac /*vacName*/, &testVac /*curVacName*/, "" /*targetVacName*/, "" /*modifyVolumeStatus*/),
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
			pvc:                                      createTestPVC(pvcName, &targetVac /*vacName*/, &testVac /*curVacName*/, testVac /*targetVacName*/, "" /*modifyVolumeStatus*/),
			pv:                                       basePV,
			vacExists:                                true,
			expectModifyCall:                         true,
			expectedModifyVolumeStatus:               nil,
			expectedCurrentVolumeAttributesClassName: &targetVac,
			expectedPVVolumeAttributesClassName:      &targetVac,
		},
		{
			name:                                     "modify volume success with extra metadata",
			pvc:                                      createTestPVC(pvcName, &targetVac /*vacName*/, &testVac /*curVacName*/, testVac /*targetVacName*/, "" /*modifyVolumeStatus*/),
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
		{
			name:                                     "modify volume rollback succeeds for infeasible errors",
			pvc:                                      createTestPVC(pvcName, &testVac /*vacName*/, &testVac /*curVacName*/, targetVac /*targetVacName*/, v1.PersistentVolumeClaimModifyVolumeInfeasible),
			pv:                                       basePV,
			vacExists:                                true,
			expectModifyCall:                         false,
			expectedModifyVolumeStatus:               nil,
			expectedCurrentVolumeAttributesClassName: &testVac,
			expectedPVVolumeAttributesClassName:      &testVac,
		},
		{
			name:                                     "modify volume rollback to nil succeeds for infeasible errors",
			pvc:                                      createTestPVC(pvcName, nil /*vacName*/, nil /*curVacName*/, targetVac /*targetVacName*/, v1.PersistentVolumeClaimModifyVolumeInfeasible),
			pv:                                       createTestPV(1, pvcName, pvcNamespace, "foobaz" /*pvcUID*/, &fsVolumeMode, nil),
			vacExists:                                true,
			expectModifyCall:                         false,
			expectedModifyVolumeStatus:               nil,
			expectedCurrentVolumeAttributesClassName: nil,
			expectedPVVolumeAttributesClassName:      nil,
		},
		{
			name:                                     "modify volume rollback to empty string succeeds for infeasible errors",
			pvc:                                      createTestPVC(pvcName, &emptyString /*vacName*/, nil /*curVacName*/, targetVac /*targetVacName*/, v1.PersistentVolumeClaimModifyVolumeInfeasible),
			pv:                                       createTestPV(1, pvcName, pvcNamespace, "foobaz" /*pvcUID*/, &fsVolumeMode, nil),
			vacExists:                                true,
			expectModifyCall:                         false,
			expectedModifyVolumeStatus:               nil,
			expectedCurrentVolumeAttributesClassName: nil,
			expectedPVVolumeAttributesClassName:      nil,
		},
		{
			name: "pvc in uncertain cache, no modification",
			pvc:  createTestPVC(pvcName, &targetVac, &testVac, targetVac, v1.PersistentVolumeClaimModifyVolumeInProgress),
			pv:   basePV,
			addPVCtoUncertainCache: true,
			expectModifyCall:       false,
			expectedModifyVolumeStatus: &v1.ModifyVolumeStatus{
				TargetVolumeAttributesClassName: targetVac,
				Status:                          v1.PersistentVolumeClaimModifyVolumeInProgress,
			},
			expectedCurrentVolumeAttributesClassName: &testVac,
			expectedPVVolumeAttributesClassName:      &testVac,
		},
		{
			name: "pvc not in uncertain cache and no error, no modification",
			pvc:  createTestPVC(pvcName, &targetVac, &testVac, targetVac, v1.PersistentVolumeClaimModifyVolumeInProgress),
			pv:   basePV,
			expectModifyCall: false,
			expectedModifyVolumeStatus: &v1.ModifyVolumeStatus{
				TargetVolumeAttributesClassName: targetVac,
				Status:                          v1.PersistentVolumeClaimModifyVolumeInProgress,
			},
			expectedCurrentVolumeAttributesClassName: &testVac,
			expectedPVVolumeAttributesClassName:      &testVac,
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

			if test.addPVCtoUncertainCache {
				ctrlInstance.uncertainPVCs[pvcNamespace+"/"+pvcName] = *test.pvc
			}

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

			if test.expectedCurrentVolumeAttributesClassName != nil && actualCurrentVolumeAttributesClassName != nil {
				if diff := cmp.Diff(*test.expectedCurrentVolumeAttributesClassName, *actualCurrentVolumeAttributesClassName); diff != "" {
					t.Errorf("expected CurrentVolumeAttributesClassName to be %v, got %v", *test.expectedCurrentVolumeAttributesClassName, *actualCurrentVolumeAttributesClassName)
				}
			}

			actualPVVolumeAttributesClassName := pv.Spec.VolumeAttributesClassName
			if test.expectedPVVolumeAttributesClassName != nil && actualPVVolumeAttributesClassName != nil {
				if diff := cmp.Diff(*test.expectedPVVolumeAttributesClassName, *actualPVVolumeAttributesClassName); diff != "" {
					t.Errorf("expected VolumeAttributesClassName of pv to be %v, got %v", *test.expectedPVVolumeAttributesClassName, *actualPVVolumeAttributesClassName)
				}
			}

			if test.withExtraMetadata && test.expectedPVVolumeAttributesClassName != nil {
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

func createTestPVC(pvcName string, vacName *string, curVacName *string, targetVacName string, modifyVolumeStatus v1.PersistentVolumeClaimModifyVolumeStatus) *v1.PersistentVolumeClaim {
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: pvcName, Namespace: pvcNamespace},
		Spec: v1.PersistentVolumeClaimSpec{
			VolumeAttributesClassName: vacName,
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
				v1.ReadOnlyMany,
			},
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("2Gi"),
				},
			},
			VolumeName: pvName,
		},
		Status: v1.PersistentVolumeClaimStatus{
			CurrentVolumeAttributesClassName: curVacName,
			Phase:                            v1.ClaimBound,
			Capacity: v1.ResourceList{
				v1.ResourceStorage: resource.MustParse("2Gi"),
			},
			ModifyVolumeStatus: &v1.ModifyVolumeStatus{
				TargetVolumeAttributesClassName: targetVacName,
				Status:                          modifyVolumeStatus,
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

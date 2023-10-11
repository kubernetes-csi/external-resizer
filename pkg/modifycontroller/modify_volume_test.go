package modifycontroller

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	"github.com/kubernetes-csi/external-resizer/pkg/features"
	"github.com/kubernetes-csi/external-resizer/pkg/modifier"
	v1 "k8s.io/api/core/v1"
	storagev1alpha1 "k8s.io/api/storage/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/util/workqueue"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
)

var (
	testVacObject = &storagev1alpha1.VolumeAttributesClass{
		ObjectMeta: metav1.ObjectMeta{Name: testVac},
		DriverName: "test-driver",
		Parameters: map[string]string{"iops": "3000"},
	}

	targetVacObject = &storagev1alpha1.VolumeAttributesClass{
		ObjectMeta: metav1.ObjectMeta{Name: targetVac},
		DriverName: "test-driver",
		Parameters: map[string]string{"iops": "4567"},
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
			pvc:              createTestPVC(pvcName, targetVac /*vacName*/, testVac /*curVacName*/, testVac /*targetVacName*/),
			pv:               basePV,
			expectModifyCall: false,
			expectedModifyVolumeStatus: &v1.ModifyVolumeStatus{
				TargetVolumeAttributesClassName: testVac,
				Status:                          v1.PersistentVolumeClaimModifyVolumePending,
			},
			expectedCurrentVolumeAttributesClassName: &testVac,
			expectedPVVolumeAttributesClassName:      &testVac,
		},
		{
			name:             "modify volume success",
			pvc:              createTestPVC(pvcName, targetVac /*vacName*/, testVac /*curVacName*/, testVac /*targetVacName*/),
			pv:               basePV,
			vacExists:        true,
			expectModifyCall: true,
			expectedModifyVolumeStatus: &v1.ModifyVolumeStatus{
				TargetVolumeAttributesClassName: targetVac,
				Status:                          "",
			},
			expectedCurrentVolumeAttributesClassName: &targetVac,
			expectedPVVolumeAttributesClassName:      &targetVac,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			// Setup
			defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.VolumeAttributesClass, true)()
			client := csi.NewMockClient("foo", true, true, true, true, true)
			driverName, _ := client.GetDriverName(context.TODO())

			var initialObjects []runtime.Object
			initialObjects = append(initialObjects, test.pvc)
			initialObjects = append(initialObjects, test.pv)
			// existing vac set in the pvc and pv
			initialObjects = append(initialObjects, testVacObject)
			if test.vacExists {
				initialObjects = append(initialObjects, targetVacObject)
			}

			kubeClient, informerFactory := fakeK8s(initialObjects)
			pvInformer := informerFactory.Core().V1().PersistentVolumes()
			pvcInformer := informerFactory.Core().V1().PersistentVolumeClaims()
			vacInformer := informerFactory.Storage().V1alpha1().VolumeAttributesClasses()

			csiModifier, err := modifier.NewModifierFromClient(client, 15*time.Second, kubeClient, informerFactory, driverName)
			if err != nil {
				t.Fatalf("Test %s: Unable to create modifier: %v", test.name, err)
			}
			controller := NewModifyController(driverName,
				csiModifier, kubeClient,
				time.Second, informerFactory,
				workqueue.DefaultControllerRateLimiter())

			ctrlInstance, _ := controller.(*modifyController)

			stopCh := make(chan struct{})
			informerFactory.Start(stopCh)

			for _, obj := range initialObjects {
				switch obj.(type) {
				case *v1.PersistentVolume:
					pvInformer.Informer().GetStore().Add(obj)
				case *v1.PersistentVolumeClaim:
					pvcInformer.Informer().GetStore().Add(obj)
				case *storagev1alpha1.VolumeAttributesClass:
					vacInformer.Informer().GetStore().Add(obj)
				default:
					t.Fatalf("Test %s: Unknown initalObject type: %+v", test.name, obj)
				}
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

			if diff := cmp.Diff(*test.expectedCurrentVolumeAttributesClassName, *actualCurrentVolumeAttributesClassName); diff != "" {
				t.Errorf("expected CurrentVolumeAttributesClassName to be %v, got %v", *test.expectedCurrentVolumeAttributesClassName, *actualCurrentVolumeAttributesClassName)
			}

			actualPVVolumeAttributesClassName := pv.Spec.VolumeAttributesClassName
			if diff := cmp.Diff(*test.expectedPVVolumeAttributesClassName, *actualPVVolumeAttributesClassName); diff != "" {
				t.Errorf("expected VolumeAttributesClassName of pv to be %v, got %v", *test.expectedPVVolumeAttributesClassName, *actualPVVolumeAttributesClassName)
			}
		})
	}
}

func createTestPVC(pvcName string, vacName string, curVacName string, targetVacName string) *v1.PersistentVolumeClaim {
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
			VolumeAttributesClassName: &vacName,
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

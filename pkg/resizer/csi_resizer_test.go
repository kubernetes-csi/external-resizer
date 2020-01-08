package resizer

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kubernetes-csi/csi-lib-utils/metrics"
	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	"github.com/kubernetes-csi/external-resizer/pkg/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewResizer(t *testing.T) {
	for i, c := range []struct {
		SupportsNodeResize              bool
		SupportsControllerResize        bool
		SupportsPluginControllerService bool

		Error   error
		Trivial bool
	}{
		// Create succeeded.
		{
			SupportsNodeResize:              true,
			SupportsControllerResize:        true,
			SupportsPluginControllerService: true,

			Trivial: false,
		},
		// Controller service not supported.
		{
			SupportsNodeResize:              true,
			SupportsControllerResize:        true,
			SupportsPluginControllerService: false,

			Error: controllerServiceNotSupportErr,
		},
		// Only node resize supported.
		{
			SupportsNodeResize:              true,
			SupportsControllerResize:        false,
			SupportsPluginControllerService: true,

			Trivial: true,
		},
		// Both controller and node resize not supported.
		{
			SupportsNodeResize:              false,
			SupportsControllerResize:        false,
			SupportsPluginControllerService: true,

			Error: resizeNotSupportErr,
		},
	} {
		client := csi.NewMockClient("mock", c.SupportsNodeResize, c.SupportsControllerResize, c.SupportsPluginControllerService)
		metricsManager := metrics.NewCSIMetricsManager("" /* driverName */)
		metricsAddress := ""
		metricsPath := ""
		k8sClient, informerFactory := fakeK8s()
		resizer, err := NewResizerFromClient(client, 0, k8sClient, informerFactory, metricsManager, metricsAddress, metricsPath)
		if err != c.Error {
			t.Errorf("Case %d: Unexpected error: wanted %v, got %v", i, c.Error, err)
		}
		if c.Error == nil {
			_, isTrivialResizer := resizer.(*trivialResizer)
			if isTrivialResizer != c.Trivial {
				t.Errorf("Case %d: Wrong trivial atrribute: wanted %t, got %t", i, c.Trivial, isTrivialResizer)
			}
		}
	}
}

func TestResizeWithSecret(t *testing.T) {
	tests := []struct {
		name               string
		hasExpansionSecret bool
		expectSecrets      bool
	}{
		{
			name:               "when CSI source has expansion secret",
			hasExpansionSecret: true,
			expectSecrets:      true,
		},
		{
			name:               "when CSI source has no secret",
			hasExpansionSecret: false,
			expectSecrets:      false,
		},
	}
	for _, tc := range tests {
		client := csi.NewMockClient("mock", true, true, true)
		secret := makeSecret("some-secret", "secret-namespace")
		k8sClient := fake.NewSimpleClientset(secret)
		pv := makeTestPV("test-csi", 2, "ebs-csi", "vol-abcde", tc.hasExpansionSecret)
		csiResizer := &csiResizer{
			name:      "ebs-csi",
			client:    client,
			timeout:   10 * time.Second,
			k8sClient: k8sClient,
		}
		_, _, err := csiResizer.Resize(pv, resource.MustParse("10Gi"))
		if err != nil {
			t.Errorf("unexpected error while expansion : %v", err)
		}
		usedSecrets := client.GetSecrets()
		if !tc.expectSecrets && len(usedSecrets) > 0 {
			t.Errorf("expected no secrets, got : %+v", usedSecrets)
		}

		if tc.expectSecrets && len(usedSecrets) == 0 {
			t.Errorf("expected secrets got none")
		}
	}

}

func TestResizeMigratedPV(t *testing.T) {
	testCases := []struct {
		name               string
		driverName         string
		pv                 *v1.PersistentVolume
		nodeResizeRequired bool
		err                error
	}{
		{
			name:               "Test AWS EBS CSI Driver",
			driverName:         "ebs.csi.aws.com",
			pv:                 createInTreeEBSPV(1),
			nodeResizeRequired: true,
		},
		{
			name:               "Test GCE PD Driver",
			driverName:         "pd.csi.storage.gke.io",
			pv:                 createInTreeGCEPDPV(1),
			nodeResizeRequired: true,
		},
		{
			name:               "Test unknonwn driver",
			driverName:         "unknown",
			pv:                 createInTreeEBSPV(1),
			nodeResizeRequired: true,
			err:                errors.New("volume testEBSPV is not migrated to CSI"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			driverName := tc.driverName
			client := csi.NewMockClient(driverName, true, true, true)
			k8sClient, informerFactory := fakeK8s()
			metricsManager := metrics.NewCSIMetricsManager("" /* driverName */)
			metricsAddress := ""
			metricsPath := ""
			resizer, err := NewResizerFromClient(client, 0, k8sClient, informerFactory, metricsManager, metricsAddress, metricsPath)
			if err != nil {
				t.Fatalf("Failed to create resizer: %v", err)
			}

			pv := tc.pv
			expectedSize := quantityGB(2)
			newSize, nodeResizeRequired, err := resizer.Resize(pv, expectedSize)

			if tc.err != nil {
				if err == nil {
					t.Fatalf("Got wrong error, wanted: %v, got: %v", tc.err, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Failed to resize the PV: %v", err)
				}

				if newSize != expectedSize {
					t.Fatalf("newSize mismatches, wanted: %v, got: %v", expectedSize, newSize)
				}
				if nodeResizeRequired != tc.nodeResizeRequired {
					t.Fatalf("nodeResizeRequired mismatches, wanted: %v, got: %v", tc.nodeResizeRequired, nodeResizeRequired)
				}
			}
		})
	}
}

func TestCanSupport(t *testing.T) {
	testCases := []struct {
		name       string
		driverName string
		pv         *v1.PersistentVolume
		pvc        *v1.PersistentVolumeClaim
		canSupport bool
	}{
		{
			name:       "EBS PV/PVC is supported",
			driverName: "ebs.csi.aws.com",
			pv:         createInTreeEBSPV(1),
			pvc:        createPVC("ebs.csi.aws.com"),
			canSupport: true,
		},
		{
			name:       "EBS PV/PVC is not supported when migartion is disabled",
			driverName: "ebs.csi.aws.com",
			pv:         createInTreeEBSPV(1),
			pvc:        createPVC("kubernetes.io/aws-ebs"),
			canSupport: false,
		},
		{
			name:       "PD PV/PVC is supported",
			driverName: "pd.csi.storage.gke.io",
			pv:         createInTreeGCEPDPV(1),
			pvc:        createPVC("pd.csi.storage.gke.io"),
			canSupport: true,
		},
		{
			name:       "unknown PV/PVC is not supported",
			driverName: "ebs.csi.aws.com",
			pv:         createInTreeEBSPV(1),
			pvc:        createPVC("unknown"),
			canSupport: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			driverName := tc.driverName
			client := csi.NewMockClient(driverName, true, true, true)
			k8sClient, informerFactory := fakeK8s()
			metricsManager := metrics.NewCSIMetricsManager("" /* driverName */)
			metricsAddress := ""
			metricsPath := ""
			resizer, err := NewResizerFromClient(client, 0, k8sClient, informerFactory, metricsManager, metricsAddress, metricsPath)
			if err != nil {
				t.Fatalf("Failed to create resizer: %v", err)
			}

			canSupport := resizer.CanSupport(tc.pv, tc.pvc)
			if canSupport != tc.canSupport {
				t.Fatalf("Wrong canSupport, wanted: %v got: %v", tc.canSupport, canSupport)
			}
		})
	}
}

func quantityGB(i int) resource.Quantity {
	q := resource.NewQuantity(int64(i*1024*1024), resource.BinarySI)
	return *q
}

func createPVC(resizerName string) *v1.PersistentVolumeClaim {
	request := quantityGB(2)
	capacity := quantityGB(1)

	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testPVC",
			Namespace: "test",
			Annotations: map[string]string{
				util.VolumeResizerKey: resizerName,
			},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			Resources: v1.ResourceRequirements{
				Requests: map[v1.ResourceName]resource.Quantity{
					v1.ResourceStorage: request,
				},
			},
			VolumeName: "testPV",
		},
		Status: v1.PersistentVolumeClaimStatus{
			Phase: v1.ClaimBound,
			Capacity: map[v1.ResourceName]resource.Quantity{
				v1.ResourceStorage: capacity,
			},
		},
	}
}

func makeTestPV(name string, sizeGig int, driverName, volID string, withSecret bool) *v1.PersistentVolume {
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.PersistentVolumeSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): resource.MustParse(
					fmt.Sprintf("%dGi", sizeGig),
				),
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				CSI: &v1.CSIPersistentVolumeSource{
					Driver:       driverName,
					VolumeHandle: volID,
					ReadOnly:     false,
				},
			},
		},
	}
	if withSecret {
		pv.Spec.CSI.ControllerExpandSecretRef = &v1.SecretReference{
			Name:      "some-secret",
			Namespace: "secret-namespace",
		}
	}
	return pv
}

func fakeK8s() (kubernetes.Interface, informers.SharedInformerFactory) {
	client := fake.NewSimpleClientset()
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	return client, informerFactory
}

func createInTreeEBSPV(capacityGB int) *v1.PersistentVolume {
	capacity := quantityGB(capacityGB)

	return &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testEBSPV",
		},
		Spec: v1.PersistentVolumeSpec{
			Capacity: map[v1.ResourceName]resource.Quantity{
				v1.ResourceStorage: capacity,
			},
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				AWSElasticBlockStore: &v1.AWSElasticBlockStoreVolumeSource{
					VolumeID: "testVolumeId",
				},
			},
		},
	}
}

func createInTreeGCEPDPV(capacityGB int) *v1.PersistentVolume {
	capacity := quantityGB(capacityGB)

	return &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testPDPV",
		},
		Spec: v1.PersistentVolumeSpec{
			Capacity: map[v1.ResourceName]resource.Quantity{
				v1.ResourceStorage: capacity,
			},
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				GCEPersistentDisk: &v1.GCEPersistentDiskVolumeSource{},
			},
		},
	}
}
func makeSecret(name string, namespace string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			UID:             "23456",
			ResourceVersion: "1",
		},
		Type: "Opaque",
		Data: map[string][]byte{
			"mykey": []byte("mydata"),
		},
	}
}

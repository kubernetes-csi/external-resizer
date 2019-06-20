package resizer

import (
	"fmt"
	"testing"
	"time"

	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		client := csi.NewMockClient(c.SupportsNodeResize, c.SupportsControllerResize, c.SupportsPluginControllerService)
		k8sClient, informerFactory := fakeK8s()
		resizer, err := NewResizerFromClient(client, 0, k8sClient, informerFactory)
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
		client := csi.NewMockClient(true, true, true)
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

func makeSecret(name string, namespace string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: meta.ObjectMeta{
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

func makeTestPV(name string, sizeGig int, driverName, volID string, withSecret bool) *v1.PersistentVolume {
	pv := &v1.PersistentVolume{
		ObjectMeta: meta.ObjectMeta{
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

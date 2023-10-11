package modifier

import (
	"errors"
	"testing"

	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	controllerServiceNotSupportErr = errors.New("CSI driver does not support controller service")
)

func TestNewModifier(t *testing.T) {
	for i, c := range []struct {
		SupportsControllerModify bool
		Error                    error
	}{
		// Create succeeded.
		{
			SupportsControllerModify: true,
		},
		// Controller modify not supported.
		{
			SupportsControllerModify: false,
		},
	} {
		client := csi.NewMockClient("mock", false, false, c.SupportsControllerModify, false, false)
		driverName := "mock-driver"
		k8sClient, informerFactory := fakeK8s()
		_, err := NewModifierFromClient(client, 0, k8sClient, informerFactory, driverName)
		if err != c.Error {
			t.Errorf("Case %d: Unexpected error: wanted %v, got %v", i, c.Error, err)
		}
	}
}

func fakeK8s() (kubernetes.Interface, informers.SharedInformerFactory) {
	client := fake.NewSimpleClientset()
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	return client, informerFactory
}

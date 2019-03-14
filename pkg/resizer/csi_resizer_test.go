package resizer

import (
	"testing"
	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/informers"
)

func TestNewResizer(t *testing.T) {
	for i, c := range []struct{
		SupportsNodeResize              bool
		SupportsControllerResize        bool
		SupportsPluginControllerService bool

		Error error
		Trivial bool
	}{
		// Create succeeded.
		{
			SupportsNodeResize: true,
			SupportsControllerResize: true,
			SupportsPluginControllerService: true,

			Trivial: false,
		},
		// Controller service not supported.
		{
			SupportsNodeResize: true,
			SupportsControllerResize: true,
			SupportsPluginControllerService: false,

			Error: controllerServiceNotSupportErr,
		},
		// Only node resize supported.
		{
			SupportsNodeResize: true,
			SupportsControllerResize: false,
			SupportsPluginControllerService: true,

			Trivial: true,
		},
		// Both controller and node resize not supported.
		{
			SupportsNodeResize: false,
			SupportsControllerResize: false,
			SupportsPluginControllerService: true,

			Error: resizeNotSupportErr,
		},
	} {
		client := csi.NewMockClient(c.SupportsNodeResize, c.SupportsControllerResize, c.SupportsPluginControllerService)
		k8sClient, informerFactory := fakeK8s()
		resizer, err := newResizer(client, 0, k8sClient, informerFactory)
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

func fakeK8s() (kubernetes.Interface, informers.SharedInformerFactory) {
	client := fake.NewSimpleClientset()
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	return client, informerFactory
}

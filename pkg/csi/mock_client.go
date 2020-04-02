package csi

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

func NewMockClient(
	name string,
	supportsNodeResize bool,
	supportsControllerResize bool,
	supportsPluginControllerService bool) *MockClient {
	return &MockClient{
		name:                            name,
		supportsNodeResize:              supportsNodeResize,
		supportsControllerResize:        supportsControllerResize,
		expandCalled:                    0,
		supportsPluginControllerService: supportsPluginControllerService,
	}
}

type MockClient struct {
	name                            string
	supportsNodeResize              bool
	supportsControllerResize        bool
	supportsPluginControllerService bool
	expandCalled                    int
	usedSecrets                     map[string]string
	usedCapability                  *csi.VolumeCapability
}

func (c *MockClient) GetDriverName(context.Context) (string, error) {
	return c.name, nil
}

func (c *MockClient) SupportsPluginControllerService(context.Context) (bool, error) {
	return c.supportsPluginControllerService, nil
}

func (c *MockClient) SupportsControllerResize(context.Context) (bool, error) {
	return c.supportsControllerResize, nil
}

func (c *MockClient) SupportsNodeResize(context.Context) (bool, error) {
	return c.supportsNodeResize, nil
}

func (c *MockClient) Expand(
	ctx context.Context,
	volumeID string,
	requestBytes int64,
	secrets map[string]string,
	capability *csi.VolumeCapability) (int64, bool, error) {
	// TODO: Determine whether the operation succeeds or fails by parameters.
	c.expandCalled++
	c.usedSecrets = secrets
	c.usedCapability = capability
	return requestBytes, c.supportsNodeResize, nil
}

func (c *MockClient) GetExpandCount() int {
	return c.expandCalled
}

func (c *MockClient) GetCapability() *csi.VolumeCapability {
	return c.usedCapability
}

// GetSecrets returns secrets used for volume expansion
func (c *MockClient) GetSecrets() map[string]string {
	return c.usedSecrets
}

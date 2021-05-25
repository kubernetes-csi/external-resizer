package csi

import (
	"context"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/connection"
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
	expansionFailed                 bool
	checkMigratedLabel              bool
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

func (c *MockClient) SetExpansionFailed() {
	c.expansionFailed = true
}

func (c *MockClient) SupportsControllerGetVolume(ctx context.Context) (bool, error) {
	return false, nil
}

func (c *MockClient) GetVolume(ctx context.Context, volumeID string) (int64, error) {
	return 0, nil
}

func (c *MockClient) SetCheckMigratedLabel() {
	c.checkMigratedLabel = true
}

func (c *MockClient) Expand(
	ctx context.Context,
	volumeID string,
	requestBytes int64,
	secrets map[string]string,
	capability *csi.VolumeCapability) (int64, bool, error) {
	// TODO: Determine whether the operation succeeds or fails by parameters.
	if c.expansionFailed {
		c.expandCalled++
		return requestBytes, c.supportsNodeResize, fmt.Errorf("expansion failed")
	}
	if c.checkMigratedLabel {
		additionalInfo := ctx.Value(connection.AdditionalInfoKey)
		additionalInfoVal := additionalInfo.(connection.AdditionalInfo)
		migrated := additionalInfoVal.Migrated
		if migrated != "true" {
			err := fmt.Errorf("Expected value of migrated label: true, Actual value: %s", migrated)
			return requestBytes, c.supportsNodeResize, err
		}
	}
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

func (c *MockClient) CloseConnection() {

}

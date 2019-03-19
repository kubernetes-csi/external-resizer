package csi

import "context"

func NewMockClient(
	supportsNodeResize bool,
	supportsControllerResize bool,
	supportsPluginControllerService bool) *MockClient {
	return &MockClient{
		name:                            "mock",
		supportsNodeResize:              supportsNodeResize,
		supportsControllerResize:        supportsControllerResize,
		supportsPluginControllerService: supportsPluginControllerService,
	}
}

type MockClient struct {
	name                            string
	supportsNodeResize              bool
	supportsControllerResize        bool
	supportsPluginControllerService bool
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
	secrets map[string]string) (int64, bool, error) {
	// TODO: Determine whether the operation succeeds or fails by parameters.
	return requestBytes, c.supportsNodeResize, nil
}

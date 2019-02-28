/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
)

// Client is a gRPC client connect to remote CSI driver and abstracts all CSI calls.
type Client interface {
	// Expand expands the volume to a new size at least as big as requestBytes.
	// It returns the new size and whether the volume need expand operation on the node.
	Expand(ctx context.Context, volumeID string, requestBytes int64, secrets map[string]string) (int64, bool, error)
}

// New creates a new CSI client.
func New(conn *grpc.ClientConn) Client {
	return &client{
		ctrlClient: csi.NewControllerClient(conn),
	}
}

type client struct {
	ctrlClient csi.ControllerClient
}

func (c *client) Expand(
	ctx context.Context,
	volumeID string,
	requestBytes int64,
	secrets map[string]string) (int64, bool, error) {
	req := &csi.ControllerExpandVolumeRequest{
		Secrets:       secrets,
		VolumeId:      volumeID,
		CapacityRange: &csi.CapacityRange{RequiredBytes: requestBytes},
	}
	resp, err := c.ctrlClient.ControllerExpandVolume(ctx, req)
	if err != nil {
		return 0, false, err
	}
	return resp.CapacityBytes, resp.NodeExpansionRequired, nil
}

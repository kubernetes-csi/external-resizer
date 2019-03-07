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

package csi

import (
	"context"
	"fmt"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/connection"
	csirpc "github.com/kubernetes-csi/csi-lib-utils/rpc"
	"google.golang.org/grpc"
)

// Client is a gRPC client connect to remote CSI driver and abstracts all CSI calls.
type Client interface {
	// GetDriverName returns driver name as discovered by GetPluginInfo()
	// gRPC call.
	GetDriverName(ctx context.Context) (string, error)

	// SupportsPluginControllerService return true if the CSI driver reports
	// CONTROLLER_SERVICE in GetPluginCapabilities() gRPC call.
	SupportsPluginControllerService(ctx context.Context) (bool, error)

	// SupportsControllerResize returns whether the CSI driver reports EXPAND_VOLUME
	// in ControllerGetCapabilities() gRPC call.
	SupportsControllerResize(ctx context.Context) (bool, error)

	// Expand expands the volume to a new size at least as big as requestBytes.
	// It returns the new size and whether the volume need expand operation on the node.
	Expand(ctx context.Context, volumeID string, requestBytes int64, secrets map[string]string) (int64, bool, error)
}

// New creates a new CSI client.
func New(address string, timeout time.Duration) (Client, error) {
	conn, err := connection.Connect(address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to CSI driver: %v", err)
	}

	err = csirpc.ProbeForever(conn, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed probing CSI driver: %v", err)
	}

	return &client{
		conn:       conn,
		ctrlClient: csi.NewControllerClient(conn),
	}, nil
}

type client struct {
	conn       *grpc.ClientConn
	ctrlClient csi.ControllerClient
}

func (c *client) GetDriverName(ctx context.Context) (string, error) {
	return csirpc.GetDriverName(ctx, c.conn)
}

func (c *client) SupportsPluginControllerService(ctx context.Context) (bool, error) {
	caps, err := csirpc.GetPluginCapabilities(ctx, c.conn)
	if err != nil {
		return false, fmt.Errorf("error getting controller capabilities: %v", err)
	}
	return caps[csi.PluginCapability_Service_CONTROLLER_SERVICE], nil
}

func (c *client) SupportsControllerResize(ctx context.Context) (bool, error) {
	caps, err := csirpc.GetControllerCapabilities(ctx, c.conn)
	if err != nil {
		return false, fmt.Errorf("error getting controller capabilities: %v", err)
	}
	return caps[csi.ControllerServiceCapability_RPC_EXPAND_VOLUME], nil
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

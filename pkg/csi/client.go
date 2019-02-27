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
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"k8s.io/klog"
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

	// Probe checks that the CSI driver is ready to process requests
	Probe(ctx context.Context) error
}

// New creates a new CSI client.
func New(address string, timeout time.Duration) (Client, error) {
	conn, err := newGRPCConnection(address, timeout)
	if err != nil {
		return nil, err
	}
	return &client{
		idClient:   csi.NewIdentityClient(conn),
		ctrlClient: csi.NewControllerClient(conn),
	}, nil
}

type client struct {
	idClient   csi.IdentityClient
	ctrlClient csi.ControllerClient
}

func (c *client) GetDriverName(ctx context.Context) (string, error) {
	req := csi.GetPluginInfoRequest{}

	resp, err := c.idClient.GetPluginInfo(ctx, &req)
	if err != nil {
		return "", err
	}

	name := resp.GetName()
	if name == "" {
		return "", errors.New("driver name is empty")
	}

	return name, nil
}

func (c *client) SupportsPluginControllerService(ctx context.Context) (bool, error) {
	rsp, err := c.idClient.GetPluginCapabilities(ctx, &csi.GetPluginCapabilitiesRequest{})
	if err != nil {
		return false, err
	}
	caps := rsp.GetCapabilities()
	for _, capability := range caps {
		if capability == nil {
			continue
		}
		service := capability.GetService()
		if service == nil {
			continue
		}
		if service.GetType() == csi.PluginCapability_Service_CONTROLLER_SERVICE {
			return true, nil
		}
	}
	return false, nil
}

func (c *client) SupportsControllerResize(ctx context.Context) (bool, error) {
	rsp, err := c.ctrlClient.ControllerGetCapabilities(ctx, &csi.ControllerGetCapabilitiesRequest{})
	if err != nil {
		return false, err
	}
	caps := rsp.GetCapabilities()
	for _, capability := range caps {
		if capability == nil {
			continue
		}
		rpc := capability.GetRpc()
		if rpc == nil {
			continue
		}
		if rpc.GetType() == csi.ControllerServiceCapability_RPC_EXPAND_VOLUME {
			return true, nil
		}
	}
	return false, nil
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

func (c *client) Probe(ctx context.Context) error {
	resp, err := c.idClient.Probe(ctx, &csi.ProbeRequest{})
	if err != nil {
		return err
	}
	if resp.Ready == nil || !resp.Ready.Value {
		return errors.New("driver is still initializing")
	}
	return nil
}

func newGRPCConnection(address string, timeout time.Duration) (*grpc.ClientConn, error) {
	klog.V(2).Infof("Connecting to %s", address)
	dialOptions := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBackoffMaxDelay(time.Second),
		grpc.WithUnaryInterceptor(logGRPC),
	}
	if strings.HasPrefix(address, "/") {
		dialOptions = append(dialOptions, grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}))
	}
	conn, err := grpc.Dial(address, dialOptions...)

	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		if !conn.WaitForStateChange(ctx, conn.GetState()) {
			klog.V(4).Infof("Connection timed out")
			return conn, fmt.Errorf("Connection timed out")
		}
		if conn.GetState() == connectivity.Ready {
			klog.V(3).Infof("Connected")
			return conn, nil
		}
		klog.V(4).Infof("Still trying, connection is %s", conn.GetState())
	}
}

func logGRPC(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	klog.V(5).Infof("GRPC call: %s", method)
	klog.V(5).Infof("GRPC request: %s", protosanitizer.StripSecrets(req))
	err := invoker(ctx, method, req, reply, cc, opts...)
	klog.V(5).Infof("GRPC response: %s", protosanitizer.StripSecrets(reply))
	klog.V(5).Infof("GRPC error: %v", err)
	return err
}

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

package resizer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kubernetes-csi/external-resizer/pkg/csi"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

var (
	controllerServiceNotSupportErr = errors.New("CSI driver does not support controller service")
	resizeNotSupportErr            = errors.New("CSI driver neither supports controller resize nor node resize")
)

// NewResizer creates a new resizer responsible for resizing CSI volumes.
func NewResizer(
	address string,
	timeout time.Duration,
	k8sClient kubernetes.Interface,
	informerFactory informers.SharedInformerFactory) (Resizer, error) {
	csiClient, err := csi.New(address, timeout)
	if err != nil {
		return nil, err
	}
	return NewResizerFromClient(csiClient, timeout, k8sClient, informerFactory)
}

func NewResizerFromClient(
	csiClient csi.Client,
	timeout time.Duration,
	k8sClient kubernetes.Interface,
	informerFactory informers.SharedInformerFactory) (Resizer, error) {
	driverName, err := getDriverName(csiClient, timeout)
	if err != nil {
		return nil, fmt.Errorf("get driver name failed: %v", err)
	}

	supportControllerService, err := supportsPluginControllerService(csiClient, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to check if plugin supports controller service: %v", err)
	}

	if !supportControllerService {
		return nil, controllerServiceNotSupportErr
	}

	supportControllerResize, err := supportsControllerResize(csiClient, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to check if plugin supports controller resize: %v", err)
	}

	if !supportControllerResize {
		supportsNodeResize, err := supportsNodeResize(csiClient, timeout)
		if err != nil {
			return nil, fmt.Errorf("failed to check if plugin supports node resize: %v", err)
		}
		if supportsNodeResize {
			klog.Info("The CSI driver supports node resize only, using trivial resizer to handle resize requests")
			return newTrivialResizer(driverName), nil
		}
		return nil, resizeNotSupportErr
	}

	return &csiResizer{
		name:    driverName,
		client:  csiClient,
		timeout: timeout,

		k8sClient: k8sClient,
	}, nil
}

type csiResizer struct {
	name    string
	client  csi.Client
	timeout time.Duration

	k8sClient kubernetes.Interface
}

func (r *csiResizer) Name() string {
	return r.name
}

func (r *csiResizer) CanSupport(pv *v1.PersistentVolume) bool {
	source := pv.Spec.CSI
	if source == nil {
		klog.V(4).Infof("PV %s is not a CSI volume, skip it", pv.Name)
		return false
	}
	if source.Driver != r.name {
		klog.V(4).Infof("Skip resize PV %s for resizer %s", pv.Name, source.Driver)
		return false
	}
	return true
}

func (r *csiResizer) Resize(pv *v1.PersistentVolume, requestSize resource.Quantity) (resource.Quantity, bool, error) {
	oldSize := pv.Spec.Capacity[v1.ResourceStorage]

	source := pv.Spec.CSI
	if source == nil {
		return oldSize, false, errors.New("not a CSI volume")
	}
	volumeID := source.VolumeHandle
	if len(volumeID) == 0 {
		return oldSize, false, errors.New("empty volume handle")
	}

	var secrets map[string]string
	secreRef := source.ControllerExpandSecretRef
	if secreRef != nil {
		var err error
		secrets, err = getCredentials(r.k8sClient, secreRef)
		if err != nil {
			return oldSize, false, err
		}
	}

	ctx, cancel := timeoutCtx(r.timeout)
	defer cancel()
	newSizeBytes, nodeResizeRequired, err := r.client.Expand(ctx, volumeID, requestSize.Value(), secrets)
	if err != nil {
		return oldSize, nodeResizeRequired, err
	}
	return *resource.NewQuantity(newSizeBytes, resource.BinarySI), nodeResizeRequired, err
}

func getDriverName(client csi.Client, timeout time.Duration) (string, error) {
	ctx, cancel := timeoutCtx(timeout)
	defer cancel()
	return client.GetDriverName(ctx)
}

func supportsPluginControllerService(client csi.Client, timeout time.Duration) (bool, error) {
	ctx, cancel := timeoutCtx(timeout)
	defer cancel()
	return client.SupportsPluginControllerService(ctx)
}

func supportsControllerResize(client csi.Client, timeout time.Duration) (bool, error) {
	ctx, cancel := timeoutCtx(timeout)
	defer cancel()
	return client.SupportsControllerResize(ctx)
}

func supportsNodeResize(client csi.Client, timeout time.Duration) (bool, error) {
	ctx, cancel := timeoutCtx(timeout)
	defer cancel()
	return client.SupportsNodeResize(ctx)
}

func timeoutCtx(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

func getCredentials(k8sClient kubernetes.Interface, ref *v1.SecretReference) (map[string]string, error) {
	if ref == nil {
		return nil, nil
	}

	secret, err := k8sClient.CoreV1().Secrets(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting secret %s in namespace %s: %v", ref.Name, ref.Namespace, err)
	}

	credentials := map[string]string{}
	for key, value := range secret.Data {
		credentials[key] = string(value)
	}
	return credentials, nil
}

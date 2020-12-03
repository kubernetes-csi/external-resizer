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

	csilib "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/metrics"
	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	"github.com/kubernetes-csi/external-resizer/pkg/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	csitrans "k8s.io/csi-translation-lib"
	"k8s.io/klog/v2"
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
	informerFactory informers.SharedInformerFactory,
	metricsAddress, metricsPath string) (Resizer, error) {
	metricsManager := metrics.NewCSIMetricsManager("" /* driverName */)
	csiClient, err := csi.New(address, timeout, metricsManager)
	if err != nil {
		return nil, err
	}
	return NewResizerFromClient(
		csiClient,
		timeout,
		k8sClient,
		informerFactory,
		metricsManager,
		metricsAddress,
		metricsPath)
}

func NewResizerFromClient(
	csiClient csi.Client,
	timeout time.Duration,
	k8sClient kubernetes.Interface,
	informerFactory informers.SharedInformerFactory,
	metricsManager metrics.CSIMetricsManager,
	metricsAddress, metricsPath string) (Resizer, error) {
	driverName, err := getDriverName(csiClient, timeout)
	if err != nil {
		return nil, fmt.Errorf("get driver name failed: %v", err)
	}

	klog.V(2).Infof("CSI driver name: %q", driverName)
	metricsManager.SetDriverName(driverName)
	metricsManager.StartMetricsEndpoint(metricsAddress, metricsPath)

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

// CanSupport returns whether the PV is supported by resizer
// Resizer will resize the volume if it is CSI volume or is migration enabled in-tree volume
func (r *csiResizer) CanSupport(pv *v1.PersistentVolume, pvc *v1.PersistentVolumeClaim) bool {
	resizerName := pvc.Annotations[util.VolumeResizerKey]
	// resizerName will be CSI driver name when CSI migration is enabled
	// otherwise, it will be in-tree plugin name
	// r.name is the CSI driver name, return true only when they match
	// and the CSI driver is migrated
	translator := csitrans.New()
	if translator.IsMigratedCSIDriverByName(r.name) && resizerName == r.name {
		return true
	}
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

// Resize resizes the persistence volume given request size
// It supports both CSI volume and migrated in-tree volume
func (r *csiResizer) Resize(pv *v1.PersistentVolume, requestSize resource.Quantity) (resource.Quantity, bool, error) {
	oldSize := pv.Spec.Capacity[v1.ResourceStorage]

	var volumeID string
	var source *v1.CSIPersistentVolumeSource
	var pvSpec v1.PersistentVolumeSpec
	if pv.Spec.CSI != nil {
		// handle CSI volume
		source = pv.Spec.CSI
		volumeID = source.VolumeHandle
		pvSpec = pv.Spec
	} else {
		translator := csitrans.New()
		if translator.IsMigratedCSIDriverByName(r.name) {
			// handle migrated in-tree volume
			csiPV, err := translator.TranslateInTreePVToCSI(pv)
			if err != nil {
				return oldSize, false, fmt.Errorf("failed to translate persistent volume: %v", err)
			}
			source = csiPV.Spec.CSI
			pvSpec = csiPV.Spec
			volumeID = source.VolumeHandle
		} else {
			// non-migrated in-tree volume
			return oldSize, false, fmt.Errorf("volume %v is not migrated to CSI", pv.Name)
		}
	}

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

	capability, err := GetVolumeCapabilities(pvSpec)
	if err != nil {
		return oldSize, false, fmt.Errorf("failed to get capabilities of volume %s with %v", pv.Name, err)
	}

	ctx, cancel := timeoutCtx(r.timeout)
	defer cancel()
	newSizeBytes, nodeResizeRequired, err := r.client.Expand(ctx, volumeID, requestSize.Value(), secrets, capability)
	if err != nil {
		return oldSize, nodeResizeRequired, err
	}

	return *resource.NewQuantity(newSizeBytes, resource.BinarySI), nodeResizeRequired, err
}

// GetVolumeCapabilities returns volumecapability from PV spec
func GetVolumeCapabilities(pvSpec v1.PersistentVolumeSpec) (*csilib.VolumeCapability, error) {
	m := map[v1.PersistentVolumeAccessMode]bool{}
	for _, mode := range pvSpec.AccessModes {
		m[mode] = true
	}

	if pvSpec.CSI == nil {
		return nil, errors.New("CSI volume source was nil")
	}

	var cap *csilib.VolumeCapability
	if pvSpec.VolumeMode != nil && *pvSpec.VolumeMode == v1.PersistentVolumeBlock {
		cap = &csilib.VolumeCapability{
			AccessType: &csilib.VolumeCapability_Block{
				Block: &csilib.VolumeCapability_BlockVolume{},
			},
			AccessMode: &csilib.VolumeCapability_AccessMode{},
		}

	} else {
		fsType := pvSpec.CSI.FSType

		cap = &csilib.VolumeCapability{
			AccessType: &csilib.VolumeCapability_Mount{
				Mount: &csilib.VolumeCapability_MountVolume{
					FsType:     fsType,
					MountFlags: pvSpec.MountOptions,
				},
			},
			AccessMode: &csilib.VolumeCapability_AccessMode{},
		}
	}

	// Translate array of modes into single VolumeCapability
	switch {
	case m[v1.ReadWriteMany]:
		// ReadWriteMany trumps everything, regardless what other modes are set
		cap.AccessMode.Mode = csilib.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER

	case m[v1.ReadOnlyMany] && m[v1.ReadWriteOnce]:
		// This is no way how to translate this to CSI...
		return nil, fmt.Errorf("CSI does not support ReadOnlyMany and ReadWriteOnce on the same PersistentVolume")

	case m[v1.ReadOnlyMany]:
		// There is only ReadOnlyMany set
		cap.AccessMode.Mode = csilib.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY

	case m[v1.ReadWriteOnce]:
		// There is only ReadWriteOnce set
		cap.AccessMode.Mode = csilib.VolumeCapability_AccessMode_SINGLE_NODE_WRITER

	default:
		return nil, fmt.Errorf("unsupported AccessMode combination: %+v", pvSpec.AccessModes)
	}
	return cap, nil
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

	secret, err := k8sClient.CoreV1().Secrets(ref.Namespace).Get(context.TODO(), ref.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting secret %s in namespace %s: %v", ref.Name, ref.Namespace, err)
	}

	credentials := map[string]string{}
	for key, value := range secret.Data {
		credentials[key] = string(value)
	}
	return credentials, nil
}

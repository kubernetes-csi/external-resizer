/*
Copyright 2023 The Kubernetes Authors.

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

package modifier

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

func NewModifierFromClient(
	csiClient csi.Client,
	timeout time.Duration,
	k8sClient kubernetes.Interface,
	informerFactory informers.SharedInformerFactory,
	driverName string) (Modifier, error) {

	_, err := supportsControllerModify(csiClient, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to check if plugin supports controller modify: %v", err)
	}

	return &csiModifier{
		name:    driverName,
		client:  csiClient,
		timeout: timeout,

		k8sClient: k8sClient,
	}, nil
}

type csiModifier struct {
	name    string
	client  csi.Client
	timeout time.Duration

	k8sClient kubernetes.Interface
}

func (r *csiModifier) Name() string {
	return r.name
}

func (r *csiModifier) Modify(pv *v1.PersistentVolume, mutableParameters map[string]string) error {

	var volumeID string
	var source *v1.CSIPersistentVolumeSource

	if pv.Spec.CSI != nil {
		// handle CSI volume
		source = pv.Spec.CSI
		volumeID = source.VolumeHandle
	} else {
		return fmt.Errorf("volume %v is not a CSI volumes, modify volume feature only supports CSI volumes", pv.Name)
	}

	if len(volumeID) == 0 {
		return errors.New("empty volume handle")
	}

	var secrets map[string]string

	ctx, cancel := timeoutCtx(r.timeout)

	defer cancel()
	err := r.client.Modify(ctx, volumeID, secrets, mutableParameters)
	if err != nil {
		return err
	}
	return nil
}

func supportsControllerModify(client csi.Client, timeout time.Duration) (bool, error) {
	ctx, cancel := timeoutCtx(timeout)
	defer cancel()
	return client.SupportsControllerModify(ctx)
}

func timeoutCtx(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

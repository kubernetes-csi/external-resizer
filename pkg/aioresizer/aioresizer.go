/*
Copyright 2018 The Kubernetes Authors.

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

// TODO: Obviously "aioresizer" a horrible name but I couldn't think of a better one
package aioresizer

import (
	"context"
	"time"

	"github.com/kubernetes-csi/external-resizer/pkg/controller"
	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	"github.com/kubernetes-csi/external-resizer/pkg/features"
	"github.com/kubernetes-csi/external-resizer/pkg/modifier"
	"github.com/kubernetes-csi/external-resizer/pkg/modifycontroller"
	extresizer "github.com/kubernetes-csi/external-resizer/pkg/resizer"
	"k8s.io/apimachinery/pkg/util/wait"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
	csitrans "k8s.io/csi-translation-lib"
	"k8s.io/klog/v2"
)

const version = "unknown"

type ResizerOptions struct {
	// CLI options
	OperationTimeout       time.Duration
	RetryIntervalStart     time.Duration
	RetryIntervalMax       time.Duration
	ResyncPeriod           time.Duration
	HandleVolumeInUseError bool
	Workers                int

	// Derived options (Kubernetes config, etc)
	DriverName string

	// Shared objects (clients, informers, etc)
	Client     kubernetes.Interface
	Factory    informers.SharedInformerFactory
	CSIClient  csi.Client
	Translator csitrans.CSITranslator
}

type AIOResizer interface {
	Run(ctx context.Context)
	GetVersion() string
}

type resizer struct {
	options          ResizerOptions
	csiResizer       extresizer.Resizer
	csiModifier      modifier.Modifier
	resizeController controller.ResizeController
	modifyController modifycontroller.ModifyController
}

func (r *resizer) Run(ctx context.Context) {
	r.options.Factory.Start(wait.NeverStop)
	go r.resizeController.Run(r.options.Workers, ctx)
	if utilfeature.DefaultFeatureGate.Enabled(features.VolumeAttributesClass) {
		go r.modifyController.Run(r.options.Workers, ctx)
	}
	<-ctx.Done()
}

func (r *resizer) GetVersion() string {
	return version
}

func NewAIOResizer(options ResizerOptions) (AIOResizer, string) {
	var err error
	r := &resizer{
		options: options,
	}

	r.csiResizer, err = extresizer.NewResizerFromClient(
		options.CSIClient,
		options.OperationTimeout,
		options.Client,
		options.DriverName)
	if err != nil {
		klog.ErrorS(err, "Failed to create CSI resizer")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	r.csiModifier, err = modifier.NewModifierFromClient(
		options.CSIClient,
		options.OperationTimeout,
		options.Client,
		options.Factory,
		options.DriverName)
	if err != nil {
		klog.ErrorS(err, "Failed to create CSI modifier")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	resizerName := r.csiResizer.Name()
	r.resizeController = controller.NewResizeController(resizerName, r.csiResizer, options.Client, options.ResyncPeriod, options.Factory,
		workqueue.NewItemExponentialFailureRateLimiter(options.RetryIntervalStart, options.RetryIntervalMax),
		options.HandleVolumeInUseError)
	modifierName := r.csiModifier.Name()
	// Add modify controller only if the feature gate is enabled
	if utilfeature.DefaultFeatureGate.Enabled(features.VolumeAttributesClass) {
		r.modifyController = modifycontroller.NewModifyController(modifierName, r.csiModifier, options.Client, options.ResyncPeriod, options.Factory,
			workqueue.NewItemExponentialFailureRateLimiter(options.RetryIntervalStart, options.RetryIntervalMax))
	}

	return r, resizerName
}

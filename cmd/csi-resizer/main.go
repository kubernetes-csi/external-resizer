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

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kubernetes-csi/csi-lib-utils/metrics"
	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/client-go/util/workqueue"

	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	"github.com/kubernetes-csi/csi-lib-utils/standardflags"
	"github.com/kubernetes-csi/external-resizer/pkg/controller"
	"github.com/kubernetes-csi/external-resizer/pkg/features"
	"github.com/kubernetes-csi/external-resizer/pkg/modifier"
	"github.com/kubernetes-csi/external-resizer/pkg/modifycontroller"
	"github.com/kubernetes-csi/external-resizer/pkg/resizer"
	"github.com/kubernetes-csi/external-resizer/pkg/util"
	csitrans "k8s.io/csi-translation-lib"

	"k8s.io/apimachinery/pkg/runtime"
	server "k8s.io/apiserver/pkg/server"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/informers"
	cflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"

	"k8s.io/component-base/featuregate"
	"k8s.io/component-base/logs"
	logsapi "k8s.io/component-base/logs/api/v1"
	_ "k8s.io/component-base/logs/json/register"
	"k8s.io/component-base/metrics/legacyregistry"
	_ "k8s.io/component-base/metrics/prometheus/clientgo/leaderelection" // register leader election in the default legacy registry
	_ "k8s.io/component-base/metrics/prometheus/workqueue"               // register work queues in the default legacy registry
)

var (
	master       = flag.String("master", "", "Master URL to build a client config from. Either this or kubeconfig needs to be set if the provisioner is being run out of cluster.")
	resyncPeriod = flag.Duration("resync-period", time.Minute*10, "Resync period for cache")
	workers      = flag.Int("workers", 10, "Concurrency to process multiple resize requests")

	extraModifyMetadata = flag.Bool("extra-modify-metadata", false, "If set, add pv/pvc metadata to plugin modify requests as parameters.")

	timeout = flag.Duration("timeout", 10*time.Second, "Timeout for waiting for CSI driver socket.")

	retryIntervalStart = flag.Duration("retry-interval-start", time.Second, "Initial retry interval of failed volume resize. It exponentially increases with each failure, up to retry-interval-max.")
	retryIntervalMax   = flag.Duration("retry-interval-max", 5*time.Minute, "Maximum retry interval of failed volume resize.")

	handleVolumeInUseError = flag.Bool("handle-volume-inuse-error", true, "Flag to turn on/off capability to handle volume in use error in resizer controller. Defaults to true if not set.")

	featureGates map[string]bool

	version = "unknown"
)

func main() {
	flag.Var(cflag.NewMapStringBool(&featureGates), "feature-gates", "A set of key=value paris that describe feature gates for alpha/experimental features for csi external resizer."+"Options are:\n"+strings.Join(utilfeature.DefaultFeatureGate.KnownFeatures(), "\n"))
	fg := featuregate.NewFeatureGate()
	logsapi.AddFeatureGates(fg)
	c := logsapi.NewLoggingConfiguration()
	logsapi.AddGoFlags(c, flag.CommandLine)
	logs.InitLogs()
	standardflags.RegisterCommonFlags(flag.CommandLine)
	standardflags.AddAutomaxprocs(klog.Infof)
	flag.Parse()
	if err := logsapi.ValidateAndApply(c, fg); err != nil {
		klog.ErrorS(err, "LoggingConfiguration is invalid")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	if standardflags.Configuration.ShowVersion {
		fmt.Println(os.Args[0], version)
		os.Exit(0)
	}
	klog.InfoS("Version", "version", version)

	if standardflags.Configuration.MetricsAddress != "" && standardflags.Configuration.HttpEndpoint != "" {
		klog.ErrorS(nil, "Only one of `--metrics-address` and `--http-endpoint` can be set.")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	addr := standardflags.Configuration.MetricsAddress
	if addr == "" {
		addr = standardflags.Configuration.HttpEndpoint
	}
	if err := utilfeature.DefaultMutableFeatureGate.SetFromMap(featureGates); err != nil {
		klog.ErrorS(err, "Failed to set feature gates")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	var config *rest.Config
	var err error
	if *master != "" || standardflags.Configuration.KubeConfig != "" {
		config, err = clientcmd.BuildConfigFromFlags(*master, standardflags.Configuration.KubeConfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		klog.ErrorS(err, "Failed to create cluster config")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	config.QPS = float32(standardflags.Configuration.KubeAPIQPS)
	config.Burst = standardflags.Configuration.KubeAPIBurst
	config.ContentType = runtime.ContentTypeProtobuf

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.ErrorS(err, "Failed to create kube client")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	// if feature gate is not explicitly set, probe if we have VAC API available
	if !utilfeature.DefaultMutableFeatureGate.ExplicitlySet(features.VolumeAttributesClass) {
		enabled, err := features.IsVolumeAttributesClassV1Enabled(kubeClient.Discovery())
		switch {
		case err != nil:
			klog.ErrorS(err, "Failed to check VolumeAttributesClass V1 API availability")
		case enabled:
			klog.InfoS("VolumeAttributesClass v1 API is available")
		default:
			klog.InfoS("Disabling VolumeAttributesClass feature gate because the VolumeAttributesClass v1 API is not available")
			if err := utilfeature.DefaultMutableFeatureGate.OverrideDefault(features.VolumeAttributesClass, false); err != nil {
				klog.Fatalf("Failed to disable VolumeAttributesClass feature gate: %v", err)
			}
		}
	}

	informerFactory := informers.NewSharedInformerFactory(kubeClient, *resyncPeriod)

	mux := http.NewServeMux()

	metricsManager := metrics.NewCSIMetricsManager("" /* driverName */)

	ctx := context.Background()
	csiClient, err := csi.New(ctx, standardflags.Configuration.CSIAddress, *timeout, metricsManager)
	if err != nil {
		klog.ErrorS(err, "Failed to create CSI client")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	driverName, err := getDriverName(csiClient, *timeout)
	if err != nil {
		klog.ErrorS(err, "Get driver name failed")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	klog.V(2).InfoS("CSI driver name", "driverName", driverName)

	translator := csitrans.New()
	if translator.IsMigratedCSIDriverByName(driverName) {
		metricsManager = metrics.NewCSIMetricsManagerWithOptions(driverName, metrics.WithMigration())
		migratedCsiClient, err := csi.New(ctx, standardflags.Configuration.CSIAddress, *timeout, metricsManager)
		if err != nil {
			klog.ErrorS(err, "Failed to create MigratedCSI client")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
		csiClient.CloseConnection()
		csiClient = migratedCsiClient
	}

	// Add default legacy registry so that metrics manager serves Go runtime and process metrics.
	// Also registers the `k8s.io/component-base/` work queue and leader election metrics we anonymously import.
	metricsManager.WithAdditionalRegistry(legacyregistry.DefaultGatherer)

	csiResizer, err := resizer.NewResizerFromClient(
		csiClient,
		*timeout,
		kubeClient,
		driverName)
	if err != nil && errors.Is(err, resizer.ResizeNotSupportErr) {
		klog.InfoS("Resize not supported", "message", err)
	} else if err != nil {
		klog.ErrorS(err, "Failed to create CSI resizer")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	csiModifier, err := modifier.NewModifierFromClient(
		csiClient,
		*timeout,
		kubeClient,
		informerFactory,
		*extraModifyMetadata,
		driverName)
	if err != nil && errors.Is(err, modifier.ModifyNotSupportErr) {
		klog.InfoS("Modify not supported", "message", err)
	} else if err != nil {
		klog.ErrorS(err, "Failed to create CSI modifier")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	if csiResizer == nil && csiModifier == nil {
		klog.Fatalf("CSI driver does not support resize nor modify")
	}

	// Start HTTP server for metrics + leader election healthz
	if addr != "" {
		metricsManager.RegisterToServer(mux, standardflags.Configuration.MetricsPath)
		metricsManager.SetDriverName(driverName)
		go func() {
			klog.InfoS("ServeMux listening", "address", addr)
			err := http.ListenAndServe(addr, mux)
			if err != nil {
				klog.ErrorS(err, "Failed to start HTTP server", "address", addr, "metricsPath", standardflags.Configuration.MetricsPath)
				klog.FlushAndExit(klog.ExitFlushTimeout, 1)
			}
		}()
	}

	leaseHolder := ""
	var rc controller.ResizeController
	if csiResizer != nil {
		resizerName := csiResizer.Name()
		rc = controller.NewResizeController(resizerName, csiResizer, kubeClient, *resyncPeriod, informerFactory,
			workqueue.NewTypedItemExponentialFailureRateLimiter[string](*retryIntervalStart, *retryIntervalMax),
			*handleVolumeInUseError, *retryIntervalMax)

		leaseHolder = resizerName
	}

	var mc modifycontroller.ModifyController
	if csiModifier != nil {
		modifierName := csiModifier.Name()
		// Add modify controller only if the feature gate is enabled
		if utilfeature.DefaultFeatureGate.Enabled(features.VolumeAttributesClass) {
			mc = modifycontroller.NewModifyController(modifierName, csiModifier, kubeClient, *resyncPeriod,
				*retryIntervalMax, *extraModifyMetadata, informerFactory,
				workqueue.NewTypedItemExponentialFailureRateLimiter[string](*retryIntervalStart, *retryIntervalMax))
		}

		if leaseHolder == "" {
			leaseHolder = modifierName
		}
	}

	// handle SIGTERM and SIGINT by cancelling the context.
	var (
		terminate       func()          // called when all controllers are finished
		controllerCtx   context.Context // shuts down all controllers on a signal
		shutdownHandler <-chan struct{} // called when the signal is received
	)

	if utilfeature.DefaultFeatureGate.Enabled(features.ReleaseLeaderElectionOnExit) {
		// ctx waits for all controllers to finish, then shuts down the whole process, incl. leader election
		ctx, terminate = context.WithCancel(ctx)
		var cancelControllerCtx context.CancelFunc
		controllerCtx, cancelControllerCtx = context.WithCancel(ctx)
		shutdownHandler = server.SetupSignalHandler()

		defer terminate()

		go func() {
			defer cancelControllerCtx()
			<-shutdownHandler
			klog.Info("Received SIGTERM or SIGINT signal, shutting down controller.")
		}()
	}

	run := func(ctx context.Context) {
		informerFactory.Start(ctx.Done())
		if utilfeature.DefaultFeatureGate.Enabled(features.ReleaseLeaderElectionOnExit) {
			var wg sync.WaitGroup
			if rc != nil {
				go rc.Run(*workers, controllerCtx, &wg)
			}
			if mc != nil && utilfeature.DefaultFeatureGate.Enabled(features.VolumeAttributesClass) {
				go mc.Run(*workers, controllerCtx, &wg)
			}
			<-controllerCtx.Done()
			wg.Wait()
			terminate()
		} else {
			if rc != nil {
				go rc.Run(*workers, ctx, nil)
			}
			if mc != nil && utilfeature.DefaultFeatureGate.Enabled(features.VolumeAttributesClass) {
				go mc.Run(*workers, ctx, nil)
			}
			<-ctx.Done()
		}
	}

	leaderelection.RunWithLeaderElection(
		ctx,
		config,
		standardflags.Configuration,
		run,
		"external-resizer-"+util.SanitizeName(resizerName),
		mux,
		utilfeature.DefaultFeatureGate.Enabled(features.ReleaseLeaderElectionOnExit),
	)
}

func getDriverName(client csi.Client, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return client.GetDriverName(ctx)
}

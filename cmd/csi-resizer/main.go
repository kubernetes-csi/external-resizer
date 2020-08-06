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
	"flag"
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/client-go/util/workqueue"

	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	"github.com/kubernetes-csi/external-resizer/pkg/controller"
	"github.com/kubernetes-csi/external-resizer/pkg/resizer"
	"github.com/kubernetes-csi/external-resizer/pkg/util"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/klog"
)

var (
	master       = flag.String("master", "", "Master URL to build a client config from. Either this or kubeconfig needs to be set if the provisioner is being run out of cluster.")
	kubeConfig   = flag.String("kubeconfig", "", "Absolute path to the kubeconfig")
	resyncPeriod = flag.Duration("resync-period", time.Minute*10, "Resync period for cache")
	workers      = flag.Int("workers", 10, "Concurrency to process multiple resize requests")

	csiAddress = flag.String("csi-address", "/run/csi/socket", "Address of the CSI driver socket.")
	timeout    = flag.Duration("timeout", 10*time.Second, "Timeout for waiting for CSI driver socket.")

	showVersion = flag.Bool("version", false, "Show version")

	retryIntervalStart = flag.Duration("retry-interval-start", time.Second, "Initial retry interval of failed volume resize. It exponentially increases with each failure, up to retry-interval-max.")
	retryIntervalMax   = flag.Duration("retry-interval-max", 5*time.Minute, "Maximum retry interval of failed volume resize.")

	enableLeaderElection    = flag.Bool("leader-election", false, "Enable leader election.")
	leaderElectionNamespace = flag.String("leader-election-namespace", "", "Namespace where the leader election resource lives. Defaults to the pod namespace if not set.")

	metricsAddress = flag.String("metrics-address", "", "The TCP network address where the prometheus metrics endpoint will listen (example: `:8080`). The default is empty string, which means metrics endpoint is disabled.")
	metricsPath    = flag.String("metrics-path", "/metrics", "The HTTP path where prometheus metrics will be exposed. Default is `/metrics`.")

	kubeAPIQPS   = flag.Float64("kube-api-qps", 5, "QPS to use while communicating with the kubernetes apiserver. Defaults to 5.0.")
	kubeAPIBurst = flag.Int("kube-api-burst", 10, "Burst to use while communicating with the kubernetes apiserver. Defaults to 10.")

	handleVolumeInUseError = flag.Bool("handle-volume-inuse-error", true, "Flag to turn on/off capability to handle volume in use error in resizer controller. Defaults to true if not set.")

	version = "unknown"
)

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	if *showVersion {
		fmt.Println(os.Args[0], version)
		os.Exit(0)
	}
	klog.Infof("Version : %s", version)

	var config *rest.Config
	var err error
	if *master != "" || *kubeConfig != "" {
		config, err = clientcmd.BuildConfigFromFlags(*master, *kubeConfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		klog.Fatal(err.Error())
	}

	config.QPS = float32(*kubeAPIQPS)
	config.Burst = *kubeAPIBurst

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err.Error())
	}

	informerFactory := informers.NewSharedInformerFactory(kubeClient, *resyncPeriod)

	csiResizer, err := resizer.NewResizer(
		*csiAddress,
		*timeout,
		kubeClient,
		informerFactory,
		*metricsAddress,
		*metricsPath)
	if err != nil {
		klog.Fatal(err.Error())
	}

	resizerName := csiResizer.Name()
	rc := controller.NewResizeController(resizerName, csiResizer, kubeClient, *resyncPeriod, informerFactory,
		workqueue.NewItemExponentialFailureRateLimiter(*retryIntervalStart, *retryIntervalMax),
		*handleVolumeInUseError)
	run := func(ctx context.Context) {
		informerFactory.Start(wait.NeverStop)
		rc.Run(*workers, ctx)

	}

	if !*enableLeaderElection {
		run(context.TODO())
	} else {
		lockName := "external-resizer-" + util.SanitizeName(resizerName)
		leKubeClient, err := kubernetes.NewForConfig(config)
		if err != nil {
			klog.Fatal(err.Error())
		}
		le := leaderelection.NewLeaderElection(leKubeClient, lockName, run)

		if *leaderElectionNamespace != "" {
			le.WithNamespace(*leaderElectionNamespace)
		}

		if err := le.Run(); err != nil {
			klog.Fatalf("error initializing leader election: %v", err)
		}
	}
}

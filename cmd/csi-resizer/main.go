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
	"flag"
	"time"

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
	csiTimeout = flag.Duration("csiTimeout", 15*time.Second, "Timeout for waiting for CSI driver socket.")

	enableLeaderElection      = flag.Bool("leader-election", false, "Enable leader election.")
	leaderElectionIdentity    = flag.String("leader-election-identity", "", "Unique identity of this resizer. Typically name of the pod where the resizer runs.")
	leaderElectionNamespace   = flag.String("leader-election-namespace", "kube-system", "Namespace where this resizer runs.")
	leaderElectionRetryPeriod = flag.Duration("leader-election-retry-period", time.Second*5,
		"The duration the clients should wait between attempting acquisition and renewal "+
			"of a leadership. This is only applicable if leader election is enabled.")
	leaderElectionLeaseDuration = flag.Duration("leader-election-lease-duration", time.Second*15,
		"The duration that non-leader candidates will wait after observing a leadership "+
			"renewal until attempting to acquire leadership of a led but unrenewed leader "+
			"slot. This is effectively the maximum duration that a leader can be stopped "+
			"before it is replaced by another candidate. This is only applicable if leader "+
			"election is enabled.")
	leaderElectionRenewDeadLine = flag.Duration("leader-election-renew-deadline", time.Second*10,
		"The duration that non-leader candidates will wait after observing a leadership "+
			"renewal until attempting to acquire leadership of a led but unrenewed leader "+
			"slot. This is effectively the maximum duration that a leader can be stopped "+
			"before it is replaced by another candidate. This is only applicable if leader "+
			"election is enabled.")
)

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	kubeClient, err := util.NewK8sClient(*master, *kubeConfig)
	if err != nil {
		klog.Fatal(err.Error())
	}

	informerFactory := informers.NewSharedInformerFactory(kubeClient, *resyncPeriod)

	csiResizer, err := resizer.NewCSIResizer(*csiAddress, *csiTimeout, kubeClient, informerFactory)
	if err != nil {
		klog.Fatal(err.Error())
	}

	resizerName := csiResizer.Name()

	var leaderElectionConfig *util.LeaderElectionConfig
	if *enableLeaderElection {
		if leaderElectionIdentity == nil || *leaderElectionIdentity == "" {
			klog.Fatal("--leader-election-identity must not be empty")
		}
		leaderElectionConfig = &util.LeaderElectionConfig{
			Identity:      *leaderElectionIdentity,
			LockName:      util.SanitizeName(resizerName),
			Namespace:     *leaderElectionNamespace,
			RetryPeriod:   *leaderElectionRetryPeriod,
			LeaseDuration: *leaderElectionLeaseDuration,
			RenewDeadLine: *leaderElectionRenewDeadLine,
		}
	}

	rc := controller.NewResizeController(resizerName, csiResizer, kubeClient, *resyncPeriod, informerFactory)

	informerFactory.Start(wait.NeverStop)
	rc.Run(*workers, leaderElectionConfig)
}

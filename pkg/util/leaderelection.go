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

package util

import (
	"context"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

// LeaderElectionConfig is a combination of common configurations to create an LeaderElector.
type LeaderElectionConfig struct {
	Identity      string
	LockName      string
	Namespace     string
	RetryPeriod   time.Duration
	LeaseDuration time.Duration
	RenewDeadLine time.Duration
}

// NewLeaderLock creates a ResourceLock, which can be used to create leaderElector.
func NewLeaderLock(
	kubeClient kubernetes.Interface,
	eventRecorder record.EventRecorder,
	config *LeaderElectionConfig) (resourcelock.Interface, error) {
	return resourcelock.New(resourcelock.EndpointsResourceLock,
		config.Namespace,
		config.LockName,
		kubeClient.CoreV1(),
		resourcelock.ResourceLockConfig{
			Identity:      config.Identity,
			EventRecorder: eventRecorder,
		})
}

// RunAsLeader creates an leaderElector and starts the main function only after becoming a leader.
func RunAsLeader(lock resourcelock.Interface, config *LeaderElectionConfig, startFunc func(context.Context)) {
	leaderelection.RunOrDie(context.TODO(), leaderelection.LeaderElectionConfig{
		Lock:          lock,
		RetryPeriod:   config.RetryPeriod,
		LeaseDuration: config.LeaseDuration,
		RenewDeadline: config.RenewDeadLine,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				klog.V(3).Info("Became leader, starting")
				startFunc(ctx)
			},
			OnStoppedLeading: func() {
				klog.Fatal("Stopped leading")
			},
			OnNewLeader: func(identity string) {
				klog.V(3).Infof("Current leader: %s", identity)
			},
		},
	})
}

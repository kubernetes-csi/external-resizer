[![Build Status](https://travis-ci.org/kubernetes-csi/external-resizer.svg?branch=master)](https://travis-ci.org/kubernetes-csi/external-resizer)

# CSI Resizer

The CSI `external-resizer` is a sidecar container that watches the Kubernetes API server for `PersistentVolumeClaim` updates and
triggers `ControllerExpandVolume` operations against a CSI endpoint if user requested more storage on `PersistentVolumeClaim` object.

## Overview

A storage provider that allows volume expansion after creation, may choose to implement volume expansion either via a
control-plane CSI RPC call or via node CSI RPC call or both as a two step process. The external-resizer is an external-controller that watches Kubernetes API server for `PersistentVolumeClaim` modifications and triggers CSI calls for control-plane volume-expansion. More details can be found on - [CSI Volume expansion](https://kubernetes-csi.github.io/docs/volume-expansion.html)

## Compatibility

This information reflects the head of this branch.

| Compatible with CSI Version                                                                | Container Image                | Recommended K8s Version |
| ------------------------------------------------------------------------------------------ | -------------------------------| --------------- |
| [CSI Spec v1.2.0](https://github.com/container-storage-interface/spec/releases/tag/v1.2.0) | quay.io/k8scsi/csi-resizer | 1.16            |



## Feature status

Currently all CSI volume expansion features are supported as Beta features by external-resizer.

## Usage

It is necessary to create a new service account and give it enough privileges to run the external-resizer, see `deploy/kubernetes/rbac.yaml`. The resizer is then deployed as single Deployment as illustrated below:

```sh
kubectl create deploy/kubernetes/deployment.yaml
```

The external-resizer may run in the same pod with other external CSI controllers such as the external-attacher, external-snapshotter and/or external-provisioner.

Note that the external-resizer does not scale with more replicas. Only one external-resizer is elected as leader and running. The others are waiting for the leader to die. They re-elect a new active leader in ~15 seconds after death of the old leader.

### Command line options

#### Recommended optional arguments

* `--csi-address <path to CSI socket>`: This is the path to the CSI driver socket inside the pod that the external-resizer container will use to issue CSI operations (`/run/csi/socket` is used by default).

* `--leader-election`: Enables leader election. This is mandatory when there are multiple replicas of the same external-resizer running for one CSI driver. Only one of them may be active (=leader). A new leader will be re-elected when current leader dies or becomes unresponsive for ~15 seconds.

* `--leader-election-namespace`: Namespace where the leader election resource lives. Defaults to the pod namespace if not set.

* `--csiTimeout <duration>`: Timeout of all calls to CSI driver. It should be set to value that accommodates majority of `ControllerExpandVolume` calls. 15 seconds is used by default.

* `--workers <num>`: Number of simultaneously running `ControllerExpandVolume` operations. Default value is `10`.

* `--metrics-address`: The TCP network address where the prometheus metrics endpoint will run (example: `:8080` which corresponds to port 8080 on local host). The default is empty string, which means metrics endpoint is disabled.

* `--metrics-path`: The HTTP path where prometheus metrics will be exposed. Default is `/metrics`.

#### Other recognized arguments

* `--kubeconfig <path>`: Path to Kubernetes client configuration that the external-resizer uses to connect to Kubernetes API server. When omitted, default token provided by Kubernetes will be used. This option is useful only when the external-resizer does not run as a Kubernetes pod, e.g. for debugging. Either this or `--master` needs to be set if the external-resizer is being run out of cluster.

* `--master <url>`: Master URL to build a client config from. When omitted, default token provided by Kubernetes will be used. This option is useful only when the external-resizer does not run as a Kubernetes pod, e.g. for debugging. Either this or `--kubeconfig` needs to be set if the external-resizer is being run out of cluster.

* `--version`: Prints current external-resizer version and quits.

* All glog / klog arguments are supported, such as `-v <log level>` or `-alsologtostderr`.


## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack](http://slack.k8s.io/)
- [Mailing List](https://groups.google.com/forum/#!forum/kubernetes-dev)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).

[owners]: https://git.k8s.io/community/contributors/guide/owners.md
[Creative Commons 4.0]: https://git.k8s.io/website/LICENSE

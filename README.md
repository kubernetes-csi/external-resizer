# CSI Resizer

The CSI `external-resizer` is a sidecar container that watches the Kubernetes API server for `PersistentVolumeClaim` updates and
triggers `ControllerExpandVolume` operations against a CSI endpoint if user requested more storage on `PersistentVolumeClaim` object.

## Overview

A storage provider that allows volume expansion after creation, may choose to implement volume expansion either via a
control-plane CSI RPC call or via node CSI RPC call or both as a two step process. The external-resizer is an external-controller that watches Kubernetes API server for `PersistentVolumeClaim` modifications and triggers CSI calls for control-plane volume-expansion. More details can be found on - [CSI Volume expansion](https://kubernetes-csi.github.io/docs/volume-expansion.html)

## Compatibility

This information reflects the head of this branch.

| Compatible with CSI Version | Container Image | [Min K8s Version](https://kubernetes-csi.github.io/docs/kubernetes-compatibility.html#minimum-version) | [Recommended K8s Version](https://kubernetes-csi.github.io/docs/kubernetes-compatibility.html#recommended-version) |
| ------------------------------------------------------------------------------------------ | -------------------------------| --------------- | ------------- |
| [CSI Spec v1.5.0](https://github.com/container-storage-interface/spec/releases/tag/v1.5.0) | k8s.gcr.io/sig-storage/csi-resizer | 1.16 | 1.28 |

## Feature status

Various external-resizer releases come with different alpha / beta features.

The following table reflects the head of this branch.

| Feature                | Status  | Default | Description                                                                                                                   |
| ---------------------- | ------- | ------- | ----------------------------------------------------------------------------------------------------------------------------- |
| VolumeExpansion        | Beta    | On      | [Support for expanding CSI volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#csi-volume-expansion).    |
| ReadWriteOncePod       | Alpha   | Off     | [Single pod access mode for PersistentVolumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#access-modes). |
| VolumeAttributesClass  | Alpha   | Off     | [Volume Attributes Classes](https://kubernetes.io/docs/concepts/storage/volume-attributes-classes).                           |

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

* `--leader-election-lease-duration <duration>`: Duration, in seconds, that non-leader candidates will wait to force acquire leadership. Defaults to 15 seconds.

* `--leader-election-renew-deadline <duration>`: Duration, in seconds, that the acting leader will retry refreshing leadership before giving up. Defaults to 10 seconds.

* `--leader-election-retry-period <duration>`: Duration, in seconds, the LeaderElector clients should wait between tries of actions. Defaults to 5 seconds.

* `--timeout <duration>`: Timeout of all calls to CSI driver. It should be set to value that accommodates majority of `ControllerExpandVolume` calls. 10 seconds is used by default.

* `-kube-api-burst <int>` : Burst to use while communicating with the kubernetes apiserver. Defaults to 10. (default 10).

* `-kube-api-qps <float>` : QPS to use while communicating with the kubernetes apiserver. Defaults to 5.0. (default 5).

* `--retry-interval-start`: The starting value of the exponential backoff for failures. 1 second is used by default.

* `--retry-interval-max`: The exponential backoff maximum value. 5 minutes is used by default.

* `--workers <num>`: Number of simultaneously running `ControllerExpandVolume` operations. Default value is `10`.

* `--http-endpoint`: The TCP network address where the HTTP server for diagnostics, including metrics and leader election health check, will listen (example: `:8080` which corresponds to port 8080 on local host). The default is empty string, which means the server is disabled.

* `--metrics-path`: The HTTP path where prometheus metrics will be exposed. Default is `/metrics`.

* `--handle-volume-inuse-error <true/false>`: Enable or disable volume-in-use error handling in external-resizer. Defaults to `true` and resize-controller will watch for all pods in all namespaces to check if PVC being expanded is in-use by a pod or not before retrying volume expansion if CSI driver throws volume-in-use error. Setting this to `false` will cause external-resizer to ignore volume-in-use error and resize-controller will retry volume expansion even if volume is already in use by a pod and CSI driver does not support expansion of in-use volumes. If CSI driver being used supports online expansion, it might be desirable to set `handle-volume-inuse-error` to `false` - to save costs associated with watching all pods in the cluster.

* `-feature-gates**: A set of key/value pairs that describe alpha/experimental features of external-resizer.
  * `AnnotateFsResize=true|false` (ALPHA - default=false): Store current size of pvc in pv's annotation, so as if pvc is deleted while expansion was pending on the node, the size of pvc can be restored to old value. This permits
    expansion on the node in case pvc was deleted while expansion was pending on the node (but completed in the controller). Use of this feature depends on Kubernetes version 1.21.

  * `RecoverVolumeExpansionFailure=true|false` (ALPHA - default=false): Allow users to reduce size of PVC if expansion to current size is failing. If the feature gate `RecoverVolumeExpansionFailure` is enabled
    and expansion has failed for a PVC, you can retry expansion with a smaller size than the previously requested value. To request a new expansion attempt with a
    smaller proposed size, edit `.spec.resources` for that PVC and choose a value that is less than the value you previously tried.
    This is useful if expansion to a higher value did not succeed because of capacity constraint.
    If that has happened, or you suspect that it might have, you can retry expansion by specifying a
    size that is within the capacity limits of underlying storage provider. You can monitor status of resize operation by watching `.status.resizeStatus` and events on the PVC. Use of this feature-gate requires Kubernetes 1.28.


#### Other recognized arguments

* `--kubeconfig <path>`: Path to Kubernetes client configuration that the external-resizer uses to connect to Kubernetes API server. When omitted, default token provided by Kubernetes will be used. This option is useful only when the external-resizer does not run as a Kubernetes pod, e.g. for debugging. Either this or `--master` needs to be set if the external-resizer is being run out of cluster.

* `--master <url>`: Master URL to build a client config from. When omitted, default token provided by Kubernetes will be used. This option is useful only when the external-resizer does not run as a Kubernetes pod, e.g. for debugging. Either this or `--kubeconfig` needs to be set if the external-resizer is being run out of cluster.

* `--metrics-address`: (deprecated) The TCP network address where the prometheus metrics endpoint will run (example: `:8080` which corresponds to port 8080 on local host). The default is empty string, which means metrics endpoint is disabled.

* `--version`: Prints current external-resizer version and quits.

* All glog / klog arguments are supported, such as `-v <log level>` or `-alsologtostderr`.

### HTTP endpoint

The external-resizer optionally exposes an HTTP endpoint at address:port specified by `--http-endpoint` argument. When set, these two paths are exposed:

* Metrics path, as set by `--metrics-path` argument (default is `/metrics`).
* Leader election health check at `/healthz/leader-election`. It is recommended to run a liveness probe against this endpoint when leader election is used to kill external-resizer leader that fails to connect to the API server to renew its leadership. See https://github.com/kubernetes-csi/csi-lib-utils/issues/66 for details.


## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

* [Slack](http://slack.k8s.io/)
* [Mailing List](https://groups.google.com/forum/#!forum/kubernetes-dev)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).

[owners]: https://git.k8s.io/community/contributors/guide/owners.md
[Creative Commons 4.0]: https://git.k8s.io/website/LICENSE

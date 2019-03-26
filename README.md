# external-resizer

This is an external Kubernetes controller which can expand volumes using CSI volume Drivers. It's under heavy development
and of alpha quality.

# Build

```bash
make csi-resizer
```

# Running external resizer

## With mock driver

```bash
./bin/csi-resizer --kubeconfig /var/run/kubernetes/admin.kubeconfig --csi-address /var/lib/kubelet/plugins/csi-mock/csi.sock
```

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack](http://slack.k8s.io/)
- [Mailing List](https://groups.google.com/forum/#!forum/kubernetes-dev)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).

[owners]: https://git.k8s.io/community/contributors/guide/owners.md
[Creative Commons 4.0]: https://git.k8s.io/website/LICENSE

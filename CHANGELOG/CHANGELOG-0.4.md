# Changelog since v0.3.0

## New Features

- Add prometheus metrics to CSI external-resizer under the /metrics endpoint. This can be enabled via the "--metrics-address" and "--metrics-path" options. ([#67](https://github.com/kubernetes-csi/external-resizer/pull/67), [@saad-ali](https://github.com/saad-ali))

## Other Notable Changes

- Migrated to Go modules, so the source builds also outside of GOPATH. ([#60](https://github.com/kubernetes-csi/external-resizer/pull/60), [@pohly](https://github.com/pohly))

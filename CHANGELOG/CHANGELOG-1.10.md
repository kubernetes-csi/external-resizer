# Release notes for 1.10.0

[Documentation](https://kubernetes-csi.github.io)

# Changelog since 1.9.0

## Changes by Kind

### Feature

- Add Modify Volume Support for VolumeAttributesClass, this requires:
  1. The feature requies Kubernetes 1.29 cluster
  2. The feature gate VolumeAttributesClass and API to be enabled in Kubernetes cluster
  3. The feature gate VolumeAttributesClass feature gate enabled in external-resizer ([#351](https://github.com/kubernetes-csi/external-resizer/pull/351), [@sunnylovestiramisu](https://github.com/sunnylovestiramisu))

### Bug or Regression

- CVE fixes: CVE-2023-44487 ([#340](https://github.com/kubernetes-csi/external-resizer/pull/340), [@dannawang0221](https://github.com/dannawang0221))

### Uncategorized

- Added support structured logging ([#330](https://github.com/kubernetes-csi/external-resizer/pull/330), [@saku3](https://github.com/saku3))
- Fix panic in recovery path if marking pvc as resize in progress fails ([#246](https://github.com/kubernetes-csi/external-resizer/pull/246), [@gnufied](https://github.com/gnufied))
- Update kubernetes dependencies to v1.29.0 ([#364](https://github.com/kubernetes-csi/external-resizer/pull/364), [@sunnylovestiramisu](https://github.com/sunnylovestiramisu))

## Dependencies

### Added
- github.com/benbjohnson/clock: [v1.1.0](https://github.com/benbjohnson/clock/tree/v1.1.0)
- golang.org/x/lint: 1621716

### Changed
- cloud.google.com/go/compute: v1.15.1 → v1.23.0
- github.com/cncf/xds/go: [06c439d → e9ce688](https://github.com/cncf/xds/go/compare/06c439d...e9ce688)
- github.com/container-storage-interface/spec: [v1.8.0 → v1.9.0](https://github.com/container-storage-interface/spec/compare/v1.8.0...v1.9.0)
- github.com/emicklei/go-restful/v3: [v3.9.0 → v3.11.0](https://github.com/emicklei/go-restful/v3/compare/v3.9.0...v3.11.0)
- github.com/envoyproxy/go-control-plane: [v0.10.3 → v0.11.1](https://github.com/envoyproxy/go-control-plane/compare/v0.10.3...v0.11.1)
- github.com/envoyproxy/protoc-gen-validate: [v0.9.1 → v1.0.2](https://github.com/envoyproxy/protoc-gen-validate/compare/v0.9.1...v1.0.2)
- github.com/fsnotify/fsnotify: [v1.6.0 → v1.7.0](https://github.com/fsnotify/fsnotify/compare/v1.6.0...v1.7.0)
- github.com/go-logr/logr: [v1.2.4 → v1.3.0](https://github.com/go-logr/logr/compare/v1.2.4...v1.3.0)
- github.com/golang/glog: [v1.0.0 → v1.1.2](https://github.com/golang/glog/compare/v1.0.0...v1.1.2)
- github.com/google/cel-go: [v0.16.0 → v0.17.7](https://github.com/google/cel-go/compare/v0.16.0...v0.17.7)
- github.com/google/go-cmp: [v0.5.9 → v0.6.0](https://github.com/google/go-cmp/compare/v0.5.9...v0.6.0)
- github.com/google/uuid: [v1.3.0 → v1.3.1](https://github.com/google/uuid/compare/v1.3.0...v1.3.1)
- github.com/gorilla/websocket: [v1.4.2 → v1.5.0](https://github.com/gorilla/websocket/compare/v1.4.2...v1.5.0)
- github.com/grpc-ecosystem/grpc-gateway/v2: [v2.7.0 → v2.16.0](https://github.com/grpc-ecosystem/grpc-gateway/v2/compare/v2.7.0...v2.16.0)
- github.com/kubernetes-csi/csi-lib-utils: [v0.14.0 → v0.17.0](https://github.com/kubernetes-csi/csi-lib-utils/compare/v0.14.0...v0.17.0)
- github.com/onsi/ginkgo/v2: [v2.9.4 → v2.13.0](https://github.com/onsi/ginkgo/v2/compare/v2.9.4...v2.13.0)
- github.com/onsi/gomega: [v1.27.6 → v1.29.0](https://github.com/onsi/gomega/compare/v1.27.6...v1.29.0)
- github.com/stretchr/testify: [v1.8.2 → v1.8.4](https://github.com/stretchr/testify/compare/v1.8.2...v1.8.4)
- github.com/yuin/goldmark: [v1.2.1 → v1.4.13](https://github.com/yuin/goldmark/compare/v1.2.1...v1.4.13)
- go.etcd.io/bbolt: v1.3.7 → v1.3.8
- go.etcd.io/etcd/api/v3: v3.5.9 → v3.5.10
- go.etcd.io/etcd/client/pkg/v3: v3.5.9 → v3.5.10
- go.etcd.io/etcd/client/v2: v2.305.9 → v2.305.10
- go.etcd.io/etcd/client/v3: v3.5.9 → v3.5.10
- go.etcd.io/etcd/pkg/v3: v3.5.9 → v3.5.10
- go.etcd.io/etcd/raft/v3: v3.5.9 → v3.5.10
- go.etcd.io/etcd/server/v3: v3.5.9 → v3.5.10
- go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc: v0.35.0 → v0.44.0
- go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp: v0.35.1 → v0.44.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc: v1.10.0 → v1.19.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace: v1.10.0 → v1.19.0
- go.opentelemetry.io/otel/metric: v0.31.0 → v1.19.0
- go.opentelemetry.io/otel/sdk: v1.10.0 → v1.19.0
- go.opentelemetry.io/otel/trace: v1.10.0 → v1.19.0
- go.opentelemetry.io/otel: v1.10.0 → v1.19.0
- go.opentelemetry.io/proto/otlp: v0.19.0 → v1.0.0
- go.uber.org/goleak: v1.2.1 → v1.1.10
- golang.org/x/crypto: v0.11.0 → v0.15.0
- golang.org/x/net: v0.13.0 → v0.18.0
- golang.org/x/oauth2: v0.8.0 → v0.13.0
- golang.org/x/sync: v0.2.0 → v0.4.0
- golang.org/x/sys: v0.10.0 → v0.14.0
- golang.org/x/term: v0.10.0 → v0.14.0
- golang.org/x/text: v0.11.0 → v0.14.0
- golang.org/x/tools: v0.8.0 → v0.12.0
- google.golang.org/appengine: v1.6.7 → v1.6.8
- google.golang.org/genproto/googleapis/api: dd9d682 → d307bd8
- google.golang.org/genproto/googleapis/rpc: 28d5490 → bbf56f3
- google.golang.org/genproto: 0005af6 → d783a09
- google.golang.org/grpc: v1.54.0 → v1.60.1
- google.golang.org/protobuf: v1.30.0 → v1.31.0
- k8s.io/api: v0.28.0 → v0.29.0
- k8s.io/apimachinery: v0.28.0 → v0.29.0
- k8s.io/apiserver: v0.28.0 → v0.29.0
- k8s.io/client-go: v0.28.0 → v0.29.0
- k8s.io/component-base: v0.28.0 → v0.29.0
- k8s.io/csi-translation-lib: v0.28.0 → v0.29.0
- k8s.io/gengo: 485abfe → 9cce18d
- k8s.io/klog/v2: v2.100.1 → v2.110.1
- k8s.io/kms: v0.28.0 → v0.29.0
- k8s.io/kube-openapi: 2695361 → 2dd684a
- k8s.io/utils: d93618c → 3b25d92
- sigs.k8s.io/apiserver-network-proxy/konnectivity-client: v0.1.2 → v0.28.0
- sigs.k8s.io/structured-merge-diff/v4: v4.2.3 → v4.4.1

### Removed
- github.com/google/gnostic: [v0.5.7-v3refs](https://github.com/google/gnostic/tree/v0.5.7-v3refs)
- go.opentelemetry.io/otel/exporters/otlp/internal/retry: v1.10.0

# Release notes for v1.11.2

[Documentation](https://kubernetes-csi.github.io)

## Changes by Kind

### Bug or Regression

- Clear modify status of a pvc when the modification operation is completed ([#423](https://github.com/kubernetes-csi/external-resizer/pull/423), [@k8s-infra-cherrypick-robot](https://github.com/k8s-infra-cherrypick-robot))
- Ensure external-resizer does not modify volumes owned by other CSI drivers ([#422](https://github.com/kubernetes-csi/external-resizer/pull/422), [@k8s-infra-cherrypick-robot](https://github.com/k8s-infra-cherrypick-robot))

## Dependencies

### Added
_Nothing has changed._

### Changed
_Nothing has changed._

### Removed
_Nothing has changed._

# Release notes for v1.11.1

[Documentation](https://kubernetes-csi.github.io)

## Changes by Kind

### Bug or Regression

- Update csi-lib-utils to v0.18.1 ([#405](https://github.com/kubernetes-csi/external-resizer/pull/405), [@solumath](https://github.com/solumath))

## Dependencies

### Added
_Nothing has changed._

### Changed
- github.com/kubernetes-csi/csi-lib-utils: [v0.18.0 → v0.18.1](https://github.com/kubernetes-csi/csi-lib-utils/compare/v0.18.0...v0.18.1)

### Removed
_Nothing has changed._

# Release notes for 1.11.0

[Documentation](https://kubernetes-csi.github.io)

# Changelog since 1.10.0

## Changes by Kind

### Feature

- Updated Kubernetes deps to v1.30 ([#397](https://github.com/kubernetes-csi/external-resizer/pull/397), [@jsafrane](https://github.com/jsafrane))

### Bug or Regression

- Fixed external-resizer crash when modifying volume with no previous volume attributes class ([#375](https://github.com/kubernetes-csi/external-resizer/pull/375), [@AndrewSirenko](https://github.com/AndrewSirenko))
- When encountering a resource conflict while attempting set `FileSystemResizeRequired` or `VolumeResizeSuccess`, there is no longer a `VolumeResizeFailed` event emitted. ([#392](https://github.com/kubernetes-csi/external-resizer/pull/392), [@jakobmoellerdev](https://github.com/jakobmoellerdev))

### Uncategorized

- Update google.golang.org/protobuf to v1.33.0 to resolve CVE-2024-24786 ([#377](https://github.com/kubernetes-csi/external-resizer/pull/377), [@dobsonj](https://github.com/dobsonj))

## Dependencies

### Added
- github.com/fxamacker/cbor/v2: [v2.6.0](https://github.com/fxamacker/cbor/v2/tree/v2.6.0)
- github.com/x448/float16: [v0.8.4](https://github.com/x448/float16/tree/v0.8.4)
- k8s.io/gengo/v2: 51d4e06

### Changed
- cloud.google.com/go/compute/metadata: v0.2.3 → v0.3.0
- cloud.google.com/go/compute: v1.23.0 → v1.24.0
- github.com/alecthomas/kingpin/v2: [v2.3.2 → v2.4.0](https://github.com/alecthomas/kingpin/v2/compare/v2.3.2...v2.4.0)
- github.com/cespare/xxhash/v2: [v2.2.0 → v2.3.0](https://github.com/cespare/xxhash/v2/compare/v2.2.0...v2.3.0)
- github.com/cncf/xds/go: [e9ce688 → 0fa0005](https://github.com/cncf/xds/go/compare/e9ce688...0fa0005)
- github.com/cpuguy83/go-md2man/v2: [v2.0.2 → v2.0.3](https://github.com/cpuguy83/go-md2man/v2/compare/v2.0.2...v2.0.3)
- github.com/emicklei/go-restful/v3: [v3.11.0 → v3.12.0](https://github.com/emicklei/go-restful/v3/compare/v3.11.0...v3.12.0)
- github.com/envoyproxy/go-control-plane: [v0.11.1 → v0.12.0](https://github.com/envoyproxy/go-control-plane/compare/v0.11.1...v0.12.0)
- github.com/envoyproxy/protoc-gen-validate: [v1.0.2 → v1.0.4](https://github.com/envoyproxy/protoc-gen-validate/compare/v1.0.2...v1.0.4)
- github.com/evanphx/json-patch: [v5.6.0+incompatible → v5.9.0+incompatible](https://github.com/evanphx/json-patch/compare/v5.6.0...v5.9.0)
- github.com/go-logr/logr: [v1.3.0 → v1.4.1](https://github.com/go-logr/logr/compare/v1.3.0...v1.4.1)
- github.com/go-logr/zapr: [v1.2.3 → v1.3.0](https://github.com/go-logr/zapr/compare/v1.2.3...v1.3.0)
- github.com/go-openapi/jsonpointer: [v0.19.6 → v0.21.0](https://github.com/go-openapi/jsonpointer/compare/v0.19.6...v0.21.0)
- github.com/go-openapi/jsonreference: [v0.20.2 → v0.21.0](https://github.com/go-openapi/jsonreference/compare/v0.20.2...v0.21.0)
- github.com/go-openapi/swag: [v0.22.3 → v0.23.0](https://github.com/go-openapi/swag/compare/v0.22.3...v0.23.0)
- github.com/golang/glog: [v1.1.2 → v1.2.0](https://github.com/golang/glog/compare/v1.1.2...v1.2.0)
- github.com/golang/protobuf: [v1.5.3 → v1.5.4](https://github.com/golang/protobuf/compare/v1.5.3...v1.5.4)
- github.com/google/cel-go: [v0.17.7 → v0.17.8](https://github.com/google/cel-go/compare/v0.17.7...v0.17.8)
- github.com/google/uuid: [v1.3.1 → v1.6.0](https://github.com/google/uuid/compare/v1.3.1...v1.6.0)
- github.com/kubernetes-csi/csi-lib-utils: [v0.17.0 → v0.18.0](https://github.com/kubernetes-csi/csi-lib-utils/compare/v0.17.0...v0.18.0)
- github.com/onsi/ginkgo/v2: [v2.13.0 → v2.15.0](https://github.com/onsi/ginkgo/v2/compare/v2.13.0...v2.15.0)
- github.com/onsi/gomega: [v1.29.0 → v1.31.0](https://github.com/onsi/gomega/compare/v1.29.0...v1.31.0)
- github.com/prometheus/client_golang: [v1.16.0 → v1.19.1](https://github.com/prometheus/client_golang/compare/v1.16.0...v1.19.1)
- github.com/prometheus/client_model: [v0.4.0 → v0.6.1](https://github.com/prometheus/client_model/compare/v0.4.0...v0.6.1)
- github.com/prometheus/common: [v0.44.0 → v0.53.0](https://github.com/prometheus/common/compare/v0.44.0...v0.53.0)
- github.com/prometheus/procfs: [v0.10.1 → v0.14.0](https://github.com/prometheus/procfs/compare/v0.10.1...v0.14.0)
- github.com/rogpeppe/go-internal: [v1.10.0 → v1.11.0](https://github.com/rogpeppe/go-internal/compare/v1.10.0...v1.11.0)
- github.com/spf13/cobra: [v1.7.0 → v1.8.0](https://github.com/spf13/cobra/compare/v1.7.0...v1.8.0)
- github.com/stretchr/objx: [v0.5.0 → v0.1.0](https://github.com/stretchr/objx/compare/v0.5.0...v0.1.0)
- github.com/stretchr/testify: [v1.8.4 → v1.9.0](https://github.com/stretchr/testify/compare/v1.8.4...v1.9.0)
- github.com/yuin/goldmark: [v1.4.13 → v1.2.1](https://github.com/yuin/goldmark/compare/v1.4.13...v1.2.1)
- go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc: v0.44.0 → v0.51.0
- go.opentelemetry.io/otel/metric: v1.19.0 → v1.26.0
- go.opentelemetry.io/otel/trace: v1.19.0 → v1.26.0
- go.opentelemetry.io/otel: v1.19.0 → v1.26.0
- go.uber.org/goleak: v1.1.10 → v1.3.0
- go.uber.org/zap: v1.19.0 → v1.27.0
- golang.org/x/crypto: v0.15.0 → v0.22.0
- golang.org/x/mod: v0.8.0 → v0.15.0
- golang.org/x/net: v0.18.0 → v0.24.0
- golang.org/x/oauth2: v0.13.0 → v0.20.0
- golang.org/x/sync: v0.4.0 → v0.7.0
- golang.org/x/sys: v0.14.0 → v0.20.0
- golang.org/x/term: v0.14.0 → v0.20.0
- golang.org/x/text: v0.14.0 → v0.15.0
- golang.org/x/time: v0.3.0 → v0.5.0
- golang.org/x/tools: v0.12.0 → v0.18.0
- google.golang.org/genproto/googleapis/api: d307bd8 → 6ceb2ff
- google.golang.org/genproto/googleapis/rpc: bbf56f3 → 6ceb2ff
- google.golang.org/genproto: d783a09 → 6ceb2ff
- google.golang.org/grpc: v1.60.1 → v1.63.2
- google.golang.org/protobuf: v1.31.0 → v1.34.1
- k8s.io/api: v0.29.0 → v0.30.0
- k8s.io/apimachinery: v0.29.0 → v0.30.0
- k8s.io/apiserver: v0.29.0 → v0.30.0
- k8s.io/client-go: v0.29.0 → v0.30.0
- k8s.io/component-base: v0.29.0 → v0.30.0
- k8s.io/csi-translation-lib: v0.29.0 → v0.30.0
- k8s.io/klog/v2: v2.110.1 → v2.120.1
- k8s.io/kms: v0.29.0 → v0.30.0
- k8s.io/kube-openapi: 2dd684a → 70dd376
- sigs.k8s.io/apiserver-network-proxy/konnectivity-client: v0.28.0 → v0.29.0
- sigs.k8s.io/yaml: v1.3.0 → v1.4.0

### Removed
- github.com/benbjohnson/clock: [v1.1.0](https://github.com/benbjohnson/clock/tree/v1.1.0)
- github.com/cncf/udpa/go: [c52dc94](https://github.com/cncf/udpa/go/tree/c52dc94)
- github.com/creack/pty: [v1.1.9](https://github.com/creack/pty/tree/v1.1.9)
- github.com/kr/pty: [v1.1.1](https://github.com/kr/pty/tree/v1.1.1)
- go.uber.org/atomic: v1.10.0
- golang.org/x/lint: 1621716
- k8s.io/gengo: 9cce18d

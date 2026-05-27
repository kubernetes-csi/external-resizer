## Release notes for 2.2.0

[Documentation](https://kubernetes-csi.github.io)


# Changelog since 2.1.0

## Changes by Kind

### Other (Cleanup or Flake)

- Bump k8s dependencies to v1.36.1 (#578, @dfajmon)

### Feature

- Allow resuming of resize operation if PVC is deleted while resize is pending (#581, @gnufied)
- Move RecoverVolumeExpansionFailure to GA (#577, @gnufied)

## Dependencies

### Added
- buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go: 8976f5b
- buf.build/go/protovalidate: v0.12.0
- k8s.io/streaming: v0.36.1

### Changed
- github.com/cncf/xds/go: [0feb691 → ee656c7](https://github.com/cncf/xds/compare/0feb691...ee656c7)
- github.com/coreos/go-oidc: [v2.3.0+incompatible → v2.5.0+incompatible](https://github.com/coreos/go-oidc/compare/v2.3.0...v2.5.0)
- github.com/coreos/go-systemd/v22: [v22.6.0 → v22.7.0](https://github.com/coreos/go-systemd/compare/v22.6.0...v22.7.0)
- github.com/envoyproxy/go-control-plane/envoy: [v1.35.0 → v1.36.0](https://github.com/envoyproxy/go-control-plane/compare/envoy/v1.35.0...envoy/v1.36.0)
- github.com/envoyproxy/go-control-plane: [75eaa19 → v0.14.0](https://github.com/envoyproxy/go-control-plane/compare/75eaa19...v0.14.0)
- github.com/envoyproxy/protoc-gen-validate: [v1.2.1 → v1.3.0](https://github.com/envoyproxy/protoc-gen-validate/compare/v1.2.1...v1.3.0)
- github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus: [v1.0.1 → v1.1.0](https://github.com/grpc-ecosystem/go-grpc-middleware/compare/providers/prometheus/v1.0.1...providers/prometheus/v1.1.0)
- github.com/grpc-ecosystem/go-grpc-middleware/v2: [v2.3.0 → v2.3.3](https://github.com/grpc-ecosystem/go-grpc-middleware/compare/v2.3.0...v2.3.3)
- github.com/grpc-ecosystem/grpc-gateway/v2: [v2.27.4 → v2.27.7](https://github.com/grpc-ecosystem/grpc-gateway/compare/v2.27.4...v2.27.7)
- github.com/kubernetes-csi/csi-lib-utils: [v0.23.1 → v0.24.0](https://github.com/kubernetes-csi/csi-lib-utils/compare/v0.23.1...v0.24.0)
- github.com/moby/spdystream: [v0.5.0 → v0.5.1](https://github.com/moby/spdystream/compare/v0.5.0...v0.5.1)
- go.etcd.io/etcd/api/v3: v3.6.7 → v3.6.8
- go.etcd.io/etcd/client/pkg/v3: v3.6.7 → v3.6.8
- go.etcd.io/etcd/client/v3: v3.6.7 → v3.6.8
- go.etcd.io/etcd/pkg/v3: v3.6.5 → v3.6.8
- go.etcd.io/etcd/server/v3: v3.6.5 → v3.6.8
- go.opentelemetry.io/contrib/detectors/gcp: v1.38.0 → v1.39.0
- go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc: v0.64.0 → v0.65.0
- go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp: v0.64.0 → v0.65.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc: v1.39.0 → v1.40.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace: v1.39.0 → v1.40.0
- go.opentelemetry.io/otel/metric: v1.39.0 → v1.41.0
- go.opentelemetry.io/otel/sdk/metric: v1.39.0 → v1.40.0
- go.opentelemetry.io/otel/sdk: v1.39.0 → v1.40.0
- go.opentelemetry.io/otel/trace: v1.39.0 → v1.41.0
- go.opentelemetry.io/otel: v1.39.0 → v1.41.0
- golang.org/x/crypto: v0.46.0 → v0.47.0
- golang.org/x/net: v0.48.0 → v0.49.0
- google.golang.org/genproto/googleapis/api: 0a764e5 → 8636f87
- google.golang.org/genproto/googleapis/rpc: 0a764e5 → 8636f87
- google.golang.org/grpc: v1.78.0 → v1.79.3
- google.golang.org/protobuf: v1.36.11 → f2248ac
- k8s.io/api: v0.35.0 → v0.36.1
- k8s.io/apimachinery: v0.35.0 → v0.36.1
- k8s.io/apiserver: v0.35.0 → v0.36.1
- k8s.io/client-go: v0.35.0 → v0.36.1
- k8s.io/component-base: v0.35.0 → v0.36.1
- k8s.io/csi-translation-lib: v0.35.0 → v0.36.1
- k8s.io/klog/v2: v2.130.1 → v2.140.0
- k8s.io/kms: v0.35.0 → v0.36.1
- k8s.io/kube-openapi: 4e65d59 → 43fb72c
- k8s.io/utils: 914a6e7 → b8788ab
- sigs.k8s.io/structured-merge-diff/v6: v6.3.1 → v6.3.2

### Removed
- github.com/Masterminds/semver/v3: [v3.4.0](https://github.com/Masterminds/semver/tree/v3.4.0)
- github.com/cenkalti/backoff/v4: [v4.3.0](https://github.com/cenkalti/backoff/tree/v4.3.0)
- github.com/go-task/slim-sprig/v3: [v3.0.0](https://github.com/go-task/slim-sprig/tree/v3.0.0)
- github.com/google/pprof: [27863c8](https://github.com/google/pprof/tree/27863c8)
- github.com/gregjones/httpcache: [901d907](https://github.com/gregjones/httpcache/tree/901d907)
- github.com/grpc-ecosystem/go-grpc-prometheus: [v1.2.0](https://github.com/grpc-ecosystem/go-grpc-prometheus/tree/v1.2.0)
- github.com/onsi/ginkgo/v2: [v2.27.2](https://github.com/onsi/ginkgo/tree/v2.27.2)
- github.com/onsi/gomega: [v1.38.2](https://github.com/onsi/gomega/tree/v1.38.2)
- github.com/pkg/errors: [v0.9.1](https://github.com/pkg/errors/tree/v0.9.1)

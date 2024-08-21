# Release notes for 1.12.0

[Documentation](https://kubernetes-csi.github.io)

# Changelog since 1.11.0

## Feature

- Upgraded VolumeAttributesClass objects and listers to v1beta1. If VolumeAttributesClass feature gate gate is enabled, this sidecar may only be used with Kubernetes v1.31. (#432, @AndrewSirenko)
- Implement less noisy failure reporting when expansion fails. If RecoverVolumeExpansionFailure feature gate is enabled, this sidecar may only be used with kubernetes v1.31. (#418, @gnufied)
- Promote VolumeAttributesClass to Beta (#429, @AndrewSirenko)

## Changes by Kind

### Bug or Regression

- Clear modify status of a pvc when the modification operation is completed (#421, @carlory)
- Fixed a bug causing external-resizer to attempt to modify volumes owned by other CSI drivers (#419, @ConnorJC3)

## Dependencies

### Added
- cel.dev/expr: v0.15.0
- github.com/antlr4-go/antlr/v4: [v4.13.0](https://github.com/antlr4-go/antlr/v4/tree/v4.13.0)
- github.com/go-task/slim-sprig/v3: [v3.0.0](https://github.com/go-task/slim-sprig/v3/tree/v3.0.0)
- gopkg.in/evanphx/json-patch.v4: v4.12.0

### Changed
- github.com/asaskevich/govalidator: [f61b66f → a9d515a](https://github.com/asaskevich/govalidator/compare/f61b66f...a9d515a)
- github.com/cenkalti/backoff/v4: [v4.2.1 → v4.3.0](https://github.com/cenkalti/backoff/v4/compare/v4.2.1...v4.3.0)
- github.com/cncf/xds/go: [0fa0005 → 555b57e](https://github.com/cncf/xds/go/compare/0fa0005...555b57e)
- github.com/container-storage-interface/spec: [v1.9.0 → v1.10.0](https://github.com/container-storage-interface/spec/compare/v1.9.0...v1.10.0)
- github.com/cpuguy83/go-md2man/v2: [v2.0.3 → v2.0.4](https://github.com/cpuguy83/go-md2man/v2/compare/v2.0.3...v2.0.4)
- github.com/davecgh/go-spew: [v1.1.1 → d8f796a](https://github.com/davecgh/go-spew/compare/v1.1.1...d8f796a)
- github.com/emicklei/go-restful/v3: [v3.12.0 → v3.12.1](https://github.com/emicklei/go-restful/v3/compare/v3.12.0...v3.12.1)
- github.com/felixge/httpsnoop: [v1.0.3 → v1.0.4](https://github.com/felixge/httpsnoop/compare/v1.0.3...v1.0.4)
- github.com/fxamacker/cbor/v2: [v2.6.0 → v2.7.0](https://github.com/fxamacker/cbor/v2/compare/v2.6.0...v2.7.0)
- github.com/go-logr/logr: [v1.4.1 → v1.4.2](https://github.com/go-logr/logr/compare/v1.4.1...v1.4.2)
- github.com/golang/glog: [v1.2.0 → v1.2.1](https://github.com/golang/glog/compare/v1.2.0...v1.2.1)
- github.com/google/cel-go: [v0.17.8 → v0.20.1](https://github.com/google/cel-go/compare/v0.17.8...v0.20.1)
- github.com/google/pprof: [4bb14d4 → 813a5fb](https://github.com/google/pprof/compare/4bb14d4...813a5fb)
- github.com/grpc-ecosystem/grpc-gateway/v2: [v2.16.0 → v2.20.0](https://github.com/grpc-ecosystem/grpc-gateway/v2/compare/v2.16.0...v2.20.0)
- github.com/imdario/mergo: [v0.3.12 → v0.3.16](https://github.com/imdario/mergo/compare/v0.3.12...v0.3.16)
- github.com/kubernetes-csi/csi-lib-utils: [v0.18.0 → v0.19.0](https://github.com/kubernetes-csi/csi-lib-utils/compare/v0.18.0...v0.19.0)
- github.com/moby/spdystream: [v0.2.0 → v0.4.0](https://github.com/moby/spdystream/compare/v0.2.0...v0.4.0)
- github.com/moby/term: [1aeaba8 → v0.5.0](https://github.com/moby/term/compare/1aeaba8...v0.5.0)
- github.com/onsi/ginkgo/v2: [v2.15.0 → v2.20.0](https://github.com/onsi/ginkgo/v2/compare/v2.15.0...v2.20.0)
- github.com/onsi/gomega: [v1.31.0 → v1.34.1](https://github.com/onsi/gomega/compare/v1.31.0...v1.34.1)
- github.com/pmezard/go-difflib: [v1.0.0 → 5d4384e](https://github.com/pmezard/go-difflib/compare/v1.0.0...5d4384e)
- github.com/prometheus/common: [v0.53.0 → v0.55.0](https://github.com/prometheus/common/compare/v0.53.0...v0.55.0)
- github.com/prometheus/procfs: [v0.14.0 → v0.15.1](https://github.com/prometheus/procfs/compare/v0.14.0...v0.15.1)
- github.com/rogpeppe/go-internal: [v1.11.0 → v1.12.0](https://github.com/rogpeppe/go-internal/compare/v1.11.0...v1.12.0)
- github.com/sirupsen/logrus: [v1.9.0 → v1.9.3](https://github.com/sirupsen/logrus/compare/v1.9.0...v1.9.3)
- github.com/spf13/cobra: [v1.8.0 → v1.8.1](https://github.com/spf13/cobra/compare/v1.8.0...v1.8.1)
- go.etcd.io/bbolt: v1.3.8 → v1.3.9
- go.etcd.io/etcd/api/v3: v3.5.10 → v3.5.14
- go.etcd.io/etcd/client/pkg/v3: v3.5.10 → v3.5.14
- go.etcd.io/etcd/client/v2: v2.305.10 → v2.305.13
- go.etcd.io/etcd/client/v3: v3.5.10 → v3.5.14
- go.etcd.io/etcd/pkg/v3: v3.5.10 → v3.5.13
- go.etcd.io/etcd/raft/v3: v3.5.10 → v3.5.13
- go.etcd.io/etcd/server/v3: v3.5.10 → v3.5.13
- go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc: v0.51.0 → v0.53.0
- go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp: v0.44.0 → v0.53.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc: v1.19.0 → v1.27.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace: v1.19.0 → v1.28.0
- go.opentelemetry.io/otel/metric: v1.26.0 → v1.28.0
- go.opentelemetry.io/otel/sdk: v1.19.0 → v1.28.0
- go.opentelemetry.io/otel/trace: v1.26.0 → v1.28.0
- go.opentelemetry.io/otel: v1.26.0 → v1.28.0
- go.opentelemetry.io/proto/otlp: v1.0.0 → v1.3.1
- golang.org/x/crypto: v0.22.0 → v0.26.0
- golang.org/x/exp: a9213ee → 8a7402a
- golang.org/x/mod: v0.15.0 → v0.20.0
- golang.org/x/net: v0.24.0 → v0.28.0
- golang.org/x/oauth2: v0.20.0 → v0.22.0
- golang.org/x/sync: v0.7.0 → v0.8.0
- golang.org/x/sys: v0.20.0 → v0.24.0
- golang.org/x/term: v0.20.0 → v0.23.0
- golang.org/x/text: v0.15.0 → v0.17.0
- golang.org/x/time: v0.5.0 → v0.6.0
- golang.org/x/tools: v0.18.0 → v0.24.0
- golang.org/x/xerrors: 04be3eb → 5ec99f8
- google.golang.org/appengine: v1.6.8 → v1.6.7
- google.golang.org/genproto/googleapis/api: 6ceb2ff → 5315273
- google.golang.org/genproto/googleapis/rpc: 6ceb2ff → 8ffd90a
- google.golang.org/genproto: 6ceb2ff → b8732ec
- google.golang.org/grpc: v1.63.2 → v1.65.0
- google.golang.org/protobuf: v1.34.1 → v1.34.2
- k8s.io/api: v0.30.0 → v0.31.0
- k8s.io/apimachinery: v0.30.0 → v0.31.0
- k8s.io/apiserver: v0.30.0 → v0.31.0
- k8s.io/client-go: v0.30.0 → v0.31.0
- k8s.io/component-base: v0.30.0 → v0.31.0
- k8s.io/csi-translation-lib: v0.30.0 → v0.31.0
- k8s.io/gengo/v2: 51d4e06 → 3b05ca7
- k8s.io/klog/v2: v2.120.1 → v2.130.1
- k8s.io/kms: v0.30.0 → v0.31.0
- k8s.io/kube-openapi: 70dd376 → 91dab69
- k8s.io/utils: 3b25d92 → 18e509b
- sigs.k8s.io/apiserver-network-proxy/konnectivity-client: v0.29.0 → v0.30.3

### Removed
- cloud.google.com/go/compute: v1.24.0
- github.com/antlr/antlr4/runtime/Go/antlr/v4: [8188dc5](https://github.com/antlr/antlr4/runtime/Go/antlr/v4/tree/8188dc5)
- github.com/evanphx/json-patch: [v5.9.0+incompatible](https://github.com/evanphx/json-patch/tree/v5.9.0)
- github.com/go-task/slim-sprig: [52ccab3](https://github.com/go-task/slim-sprig/tree/52ccab3)
- github.com/matttproud/golang_protobuf_extensions: [v1.0.4](https://github.com/matttproud/golang_protobuf_extensions/tree/v1.0.4)

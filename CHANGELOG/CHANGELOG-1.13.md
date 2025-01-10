# Release notes for 1.13.0

[Documentation](https://kubernetes-csi.github.io)

# Changelog since 1.12.0

### Feature

- Adds the flag --extra-modify-metadata, which when set to true, will inject extra PVC and PV metadata as parameters when calling ModifyVolume on CSI Drivers ([#420](https://github.com/kubernetes-csi/external-resizer/pull/420), [@mdzraf](https://github.com/mdzraf))
- Infeasible PVC modifications will be retried at a slower pace than normal failures. ([#453](https://github.com/kubernetes-csi/external-resizer/pull/453), [@AndrewSirenko](https://github.com/AndrewSirenko))
- Move RecoverVolumeExpansionFailure to beta and enabled by default. ([#459](https://github.com/kubernetes-csi/external-resizer/pull/459), [@gnufied](https://github.com/gnufied))

## Dependencies

### Added
- github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp: [v1.24.2](https://github.com/GoogleCloudPlatform/opentelemetry-operations-go/tree/detectors/gcp/v1.24.2)
- github.com/klauspost/compress: [v1.17.11](https://github.com/klauspost/compress/tree/v1.17.11)
- github.com/kylelemons/godebug: [v1.1.0](https://github.com/kylelemons/godebug/tree/v1.1.0)
- github.com/planetscale/vtprotobuf: [0393e58](https://github.com/planetscale/vtprotobuf/tree/0393e58)
- go.opentelemetry.io/auto/sdk: v1.1.0
- go.opentelemetry.io/contrib/detectors/gcp: v1.31.0
- go.opentelemetry.io/otel/sdk/metric: v1.31.0

### Changed
- cel.dev/expr: v0.15.0 → v0.18.0
- cloud.google.com/go/compute/metadata: v0.3.0 → v0.5.2
- github.com/Azure/go-ansiterm: [d185dfc → 306776e](https://github.com/Azure/go-ansiterm/compare/d185dfc...306776e)
- github.com/asaskevich/govalidator: [a9d515a → f61b66f](https://github.com/asaskevich/govalidator/compare/a9d515a...f61b66f)
- github.com/cncf/xds/go: [555b57e → b4127c9](https://github.com/cncf/xds/compare/555b57e...b4127c9)
- github.com/container-storage-interface/spec: [v1.10.0 → v1.11.0](https://github.com/container-storage-interface/spec/compare/v1.10.0...v1.11.0)
- github.com/envoyproxy/go-control-plane: [v0.12.0 → v0.13.1](https://github.com/envoyproxy/go-control-plane/compare/v0.12.0...v0.13.1)
- github.com/envoyproxy/protoc-gen-validate: [v1.0.4 → v1.1.0](https://github.com/envoyproxy/protoc-gen-validate/compare/v1.0.4...v1.1.0)
- github.com/golang/glog: [v1.2.1 → v1.2.2](https://github.com/golang/glog/compare/v1.2.1...v1.2.2)
- github.com/google/cel-go: [v0.20.1 → v0.22.0](https://github.com/google/cel-go/compare/v0.20.1...v0.22.0)
- github.com/google/gnostic-models: [v0.6.8 → v0.6.9](https://github.com/google/gnostic-models/compare/v0.6.8...v0.6.9)
- github.com/google/pprof: [813a5fb → d1b30fe](https://github.com/google/pprof/compare/813a5fb...d1b30fe)
- github.com/gregjones/httpcache: [9cad4c3 → 901d907](https://github.com/gregjones/httpcache/compare/9cad4c3...901d907)
- github.com/jonboulle/clockwork: [v0.2.2 → v0.4.0](https://github.com/jonboulle/clockwork/compare/v0.2.2...v0.4.0)
- github.com/kubernetes-csi/csi-lib-utils: [v0.19.0 → v0.20.0](https://github.com/kubernetes-csi/csi-lib-utils/compare/v0.19.0...v0.20.0)
- github.com/mailru/easyjson: [v0.7.7 → v0.9.0](https://github.com/mailru/easyjson/compare/v0.7.7...v0.9.0)
- github.com/moby/spdystream: [v0.4.0 → v0.5.0](https://github.com/moby/spdystream/compare/v0.4.0...v0.5.0)
- github.com/onsi/ginkgo/v2: [v2.20.0 → v2.21.0](https://github.com/onsi/ginkgo/compare/v2.20.0...v2.21.0)
- github.com/onsi/gomega: [v1.34.1 → v1.35.1](https://github.com/onsi/gomega/compare/v1.34.1...v1.35.1)
- github.com/prometheus/client_golang: [v1.19.1 → v1.20.5](https://github.com/prometheus/client_golang/compare/v1.19.1...v1.20.5)
- github.com/prometheus/common: [v0.55.0 → v0.61.0](https://github.com/prometheus/common/compare/v0.55.0...v0.61.0)
- github.com/rogpeppe/go-internal: [v1.12.0 → v1.13.1](https://github.com/rogpeppe/go-internal/compare/v1.12.0...v1.13.1)
- github.com/stoewer/go-strcase: [v1.2.0 → v1.3.0](https://github.com/stoewer/go-strcase/compare/v1.2.0...v1.3.0)
- github.com/stretchr/testify: [v1.9.0 → v1.10.0](https://github.com/stretchr/testify/compare/v1.9.0...v1.10.0)
- github.com/xiang90/probing: [43a291a → a49e3df](https://github.com/xiang90/probing/compare/43a291a...a49e3df)
- go.etcd.io/bbolt: v1.3.9 → v1.3.11
- go.etcd.io/etcd/api/v3: v3.5.14 → v3.5.16
- go.etcd.io/etcd/client/pkg/v3: v3.5.14 → v3.5.16
- go.etcd.io/etcd/client/v2: v2.305.13 → v2.305.16
- go.etcd.io/etcd/client/v3: v3.5.14 → v3.5.16
- go.etcd.io/etcd/pkg/v3: v3.5.13 → v3.5.16
- go.etcd.io/etcd/raft/v3: v3.5.13 → v3.5.16
- go.etcd.io/etcd/server/v3: v3.5.13 → v3.5.16
- go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc: v0.53.0 → v0.58.0
- go.opentelemetry.io/otel/metric: v1.28.0 → v1.33.0
- go.opentelemetry.io/otel/sdk: v1.28.0 → v1.31.0
- go.opentelemetry.io/otel/trace: v1.28.0 → v1.33.0
- go.opentelemetry.io/otel: v1.28.0 → v1.33.0
- golang.org/x/crypto: v0.26.0 → v0.32.0
- golang.org/x/net: v0.28.0 → v0.34.0
- golang.org/x/oauth2: v0.22.0 → v0.25.0
- golang.org/x/sync: v0.8.0 → v0.10.0
- golang.org/x/sys: v0.24.0 → v0.29.0
- golang.org/x/term: v0.23.0 → v0.28.0
- golang.org/x/text: v0.17.0 → v0.21.0
- golang.org/x/time: v0.6.0 → v0.9.0
- golang.org/x/tools: v0.24.0 → v0.26.0
- google.golang.org/genproto/googleapis/api: 5315273 → 796eee8
- google.golang.org/genproto/googleapis/rpc: 8ffd90a → 5f5ef82
- google.golang.org/genproto: b8732ec → ef43131
- google.golang.org/grpc: v1.65.0 → v1.69.2
- google.golang.org/protobuf: v1.34.2 → v1.36.2
- k8s.io/api: v0.31.0 → v0.32.0
- k8s.io/apimachinery: v0.31.0 → v0.32.0
- k8s.io/apiserver: v0.31.0 → v0.32.0
- k8s.io/client-go: v0.31.0 → v0.32.0
- k8s.io/component-base: v0.31.0 → v0.32.0
- k8s.io/csi-translation-lib: v0.31.0 → v0.32.0
- k8s.io/gengo/v2: 3b05ca7 → a7b603a
- k8s.io/kms: v0.31.0 → v0.32.0
- k8s.io/kube-openapi: 91dab69 → 2c72e55
- k8s.io/utils: 18e509b → 24370be
- sigs.k8s.io/apiserver-network-proxy/konnectivity-client: v0.30.3 → v0.31.0
- sigs.k8s.io/json: bc3834c → cfa47c3
- sigs.k8s.io/structured-merge-diff/v4: v4.4.1 → v4.5.0

### Removed
- github.com/golang/groupcache: [41bb18b](https://github.com/golang/groupcache/tree/41bb18b)
- github.com/imdario/mergo: [v0.3.16](https://github.com/imdario/mergo/tree/v0.3.16)
- google.golang.org/appengine: v1.6.7

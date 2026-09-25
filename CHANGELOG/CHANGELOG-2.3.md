## Release notes for 2.3.0

[Documentation](https://kubernetes-csi.github.io)


# Changelog since 2.2.0

## Changes by Kind

### Feature

- Add separate --resize-timeout and --modify-timeout flags (#596, @xing-yang)
- Serve client-go REST client metrics (rest_client_*) (#614, @AndrewSirenko)
- Move feature that allows resuming of resize operation if PVC is deleted while resize is pending to beta (#581, @gnufied)

### Other (Cleanup or Flake)

- Bump k8s dependencies to v1.36.1 (#578, @dfajmon)
- Bump kubernetes dependencies to v1.37.1 (#619, @gnufied)
- Bump up golang.org/x/crypto to v0.53.0 to address the vulnerability - CVE-2026-42508 (#587, @chethanv28)
- Remove RecoverVolumeExpansionFailure feature gate, the feature has been GA for awhile now (#613, @gnufied)

## Dependencies

### Added
- cloud.google.com/go/auth: v0.18.2
- github.com/apapsch/go-jsonmerge/v2: [v2.0.0](https://github.com/apapsch/go-jsonmerge/tree/v2.0.0)
- github.com/go-openapi/analysis: [v0.25.5](https://github.com/go-openapi/analysis/tree/v0.25.5)
- github.com/go-openapi/errors: [v0.22.8](https://github.com/go-openapi/errors/tree/v0.22.8)
- github.com/go-openapi/loads: [v0.25.0](https://github.com/go-openapi/loads/tree/v0.25.0)
- github.com/go-openapi/runtime/server-middleware: [v0.30.0](https://github.com/go-openapi/runtime/tree/server-middleware/v0.30.0)
- github.com/go-openapi/runtime: [v0.33.0](https://github.com/go-openapi/runtime/tree/v0.33.0)
- github.com/go-openapi/spec: [v0.22.9](https://github.com/go-openapi/spec/tree/v0.22.9)
- github.com/go-openapi/strfmt: [v0.27.0](https://github.com/go-openapi/strfmt/tree/v0.27.0)
- github.com/go-openapi/swag/pools: [v0.29.1](https://github.com/go-openapi/swag/tree/pools/v0.29.1)
- github.com/go-openapi/validate: [v0.26.1](https://github.com/go-openapi/validate/tree/v0.26.1)
- github.com/go-viper/mapstructure/v2: [v2.5.0](https://github.com/go-viper/mapstructure/tree/v2.5.0)
- github.com/google/s2a-go: [v0.1.9](https://github.com/google/s2a-go/tree/v0.1.9)
- github.com/googleapis/enterprise-certificate-proxy: [v0.3.11](https://github.com/googleapis/enterprise-certificate-proxy/tree/v0.3.11)
- github.com/googleapis/gax-go/v2: [v2.17.0](https://github.com/googleapis/gax-go/tree/v2.17.0)
- github.com/oapi-codegen/runtime: [v1.6.0](https://github.com/oapi-codegen/runtime/tree/v1.6.0)
- github.com/oklog/ulid/v2: [v2.1.1](https://github.com/oklog/ulid/tree/v2.1.1)
- go.opentelemetry.io/otel/exporters/stdout/stdouttrace: v1.45.0

### Changed
- buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go: 8976f5b → 52f3232
- buf.build/go/protovalidate: v0.12.0 → v1.0.0
- cel.dev/expr: v0.25.1 → v0.25.3
- github.com/Azure/go-ansiterm: [306776e → faa5f7b](https://github.com/Azure/go-ansiterm/compare/306776e...faa5f7b)
- github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp: [v1.30.0 → v1.33.0](https://github.com/GoogleCloudPlatform/opentelemetry-operations-go/compare/detectors/gcp/v1.30.0...detectors/gcp/v1.33.0)
- github.com/cncf/xds/go: [ee656c7 → dba9d58](https://github.com/cncf/xds/compare/ee656c7...dba9d58)
- github.com/container-storage-interface/spec: [v1.12.0 → v1.13.0](https://github.com/container-storage-interface/spec/compare/v1.12.0...v1.13.0)
- github.com/envoyproxy/go-control-plane/envoy: [v1.36.0 → v1.37.0](https://github.com/envoyproxy/go-control-plane/compare/envoy/v1.36.0...envoy/v1.37.0)
- github.com/envoyproxy/protoc-gen-validate: [v1.3.0 → v1.3.3](https://github.com/envoyproxy/protoc-gen-validate/compare/v1.3.0...v1.3.3)
- github.com/felixge/httpsnoop: [v1.0.4 → v1.1.0](https://github.com/felixge/httpsnoop/compare/v1.0.4...v1.1.0)
- github.com/fsnotify/fsnotify: [v1.9.0 → v1.10.1](https://github.com/fsnotify/fsnotify/compare/v1.9.0...v1.10.1)
- github.com/fxamacker/cbor/v2: [v2.9.0 → v2.9.3](https://github.com/fxamacker/cbor/compare/v2.9.0...v2.9.3)
- github.com/go-jose/go-jose/v4: [v4.1.3 → v4.1.4](https://github.com/go-jose/go-jose/compare/v4.1.3...v4.1.4)
- github.com/go-logr/logr: [v1.4.3 → v1.4.4](https://github.com/go-logr/logr/compare/v1.4.3...v1.4.4)
- github.com/go-openapi/jsonpointer: [v0.22.4 → v1.0.0](https://github.com/go-openapi/jsonpointer/compare/v0.22.4...v1.0.0)
- github.com/go-openapi/jsonreference: [v0.21.4 → v1.0.0](https://github.com/go-openapi/jsonreference/compare/v0.21.4...v1.0.0)
- github.com/go-openapi/swag/cmdutils: [v0.25.4 → v0.29.1](https://github.com/go-openapi/swag/compare/cmdutils/v0.25.4...cmdutils/v0.29.1)
- github.com/go-openapi/swag/conv: [v0.25.4 → v0.29.1](https://github.com/go-openapi/swag/compare/conv/v0.25.4...conv/v0.29.1)
- github.com/go-openapi/swag/fileutils: [v0.25.4 → v0.29.1](https://github.com/go-openapi/swag/compare/fileutils/v0.25.4...fileutils/v0.29.1)
- github.com/go-openapi/swag/jsonutils/fixtures_test: [v0.25.4 → v0.29.1](https://github.com/go-openapi/swag/compare/jsonutils/fixtures_test/v0.25.4...jsonutils/fixtures_test/v0.29.1)
- github.com/go-openapi/swag/jsonutils: [v0.25.4 → v0.29.1](https://github.com/go-openapi/swag/compare/jsonutils/v0.25.4...jsonutils/v0.29.1)
- github.com/go-openapi/swag/loading: [v0.25.4 → v0.29.1](https://github.com/go-openapi/swag/compare/loading/v0.25.4...loading/v0.29.1)
- github.com/go-openapi/swag/mangling: [v0.25.4 → v0.29.1](https://github.com/go-openapi/swag/compare/mangling/v0.25.4...mangling/v0.29.1)
- github.com/go-openapi/swag/netutils: [v0.25.4 → v0.29.1](https://github.com/go-openapi/swag/compare/netutils/v0.25.4...netutils/v0.29.1)
- github.com/go-openapi/swag/stringutils: [v0.25.4 → v0.29.1](https://github.com/go-openapi/swag/compare/stringutils/v0.25.4...stringutils/v0.29.1)
- github.com/go-openapi/swag/typeutils: [v0.25.4 → v0.29.1](https://github.com/go-openapi/swag/compare/typeutils/v0.25.4...typeutils/v0.29.1)
- github.com/go-openapi/swag/yamlutils: [v0.25.4 → v0.29.1](https://github.com/go-openapi/swag/compare/yamlutils/v0.25.4...yamlutils/v0.29.1)
- github.com/go-openapi/swag: [v0.25.4 → v0.29.1](https://github.com/go-openapi/swag/compare/v0.25.4...v0.29.1)
- github.com/go-openapi/testify/enable/yaml/v2: [v2.0.2 → v2.6.1](https://github.com/go-openapi/testify/compare/enable/yaml/v2/v2.0.2...enable/yaml/v2/v2.6.1)
- github.com/go-openapi/testify/v2: [v2.0.2 → v2.6.1](https://github.com/go-openapi/testify/compare/v2.0.2...v2.6.1)
- github.com/golang-jwt/jwt/v5: [v5.3.0 → v5.3.1](https://github.com/golang-jwt/jwt/compare/v5.3.0...v5.3.1)
- github.com/google/cel-go: [v0.26.1 → v0.31.0](https://github.com/google/cel-go/compare/v0.26.1...v0.31.0)
- github.com/grpc-ecosystem/go-grpc-middleware/v2: [v2.3.3 → v2.3.4](https://github.com/grpc-ecosystem/go-grpc-middleware/compare/v2.3.3...v2.3.4)
- github.com/grpc-ecosystem/grpc-gateway/v2: [v2.27.7 → v2.30.0](https://github.com/grpc-ecosystem/grpc-gateway/compare/v2.27.7...v2.30.0)
- github.com/klauspost/compress: [v1.18.0 → v1.19.1](https://github.com/klauspost/compress/compare/v1.18.0...v1.19.1)
- github.com/moby/term: [v0.5.0 → v0.5.2](https://github.com/moby/term/compare/v0.5.0...v0.5.2)
- github.com/prometheus/client_golang: [v1.23.2 → v1.24.1](https://github.com/prometheus/client_golang/compare/v1.23.2...v1.24.1)
- github.com/prometheus/client_model: [v0.6.2 → v0.6.3](https://github.com/prometheus/client_model/compare/v0.6.2...v0.6.3)
- github.com/prometheus/common: [v0.67.5 → v0.70.1](https://github.com/prometheus/common/compare/v0.67.5...v0.70.1)
- github.com/prometheus/procfs: [v0.19.2 → v0.21.1](https://github.com/prometheus/procfs/compare/v0.19.2...v0.21.1)
- github.com/sirupsen/logrus: [v1.9.3 → v1.9.4](https://github.com/sirupsen/logrus/compare/v1.9.3...v1.9.4)
- github.com/spiffe/go-spiffe/v2: [v2.6.0 → v2.7.0](https://github.com/spiffe/go-spiffe/compare/v2.6.0...v2.7.0)
- github.com/stretchr/objx: [v0.5.2 → v0.5.3](https://github.com/stretchr/objx/compare/v0.5.2...v0.5.3)
- github.com/stretchr/testify: [v1.11.1 → v1.12.1](https://github.com/stretchr/testify/compare/v1.11.1...v1.12.1)
- go.etcd.io/bbolt: v1.4.3 → v1.5.0
- go.etcd.io/etcd/api/v3: v3.6.8 → v3.7.1
- go.etcd.io/etcd/client/pkg/v3: v3.6.8 → v3.7.1
- go.etcd.io/etcd/client/v3: v3.6.8 → v3.7.1
- go.etcd.io/etcd/pkg/v3: v3.6.8 → v3.7.0
- go.etcd.io/etcd/server/v3: v3.6.8 → v3.7.0
- go.etcd.io/raft/v3: v3.6.0 → v3.7.0
- go.opentelemetry.io/contrib/detectors/gcp: v1.39.0 → v1.44.0
- go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc: v0.65.0 → v0.70.0
- go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp: v0.65.0 → v0.70.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc: v1.40.0 → v1.46.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace: v1.40.0 → v1.46.0
- go.opentelemetry.io/otel/metric: v1.41.0 → v1.46.0
- go.opentelemetry.io/otel/sdk/metric: v1.40.0 → v1.46.0
- go.opentelemetry.io/otel/sdk: v1.40.0 → v1.46.0
- go.opentelemetry.io/otel/trace: v1.41.0 → v1.46.0
- go.opentelemetry.io/otel: v1.41.0 → v1.46.0
- go.opentelemetry.io/proto/otlp: v1.9.0 → v1.11.0
- go.uber.org/zap: v1.27.1 → v1.28.0
- go.yaml.in/yaml/v2: v2.4.3 → v2.4.4
- go.yaml.in/yaml/v3: v3.0.4 → v3.0.5
- golang.org/x/crypto: v0.47.0 → v0.56.0
- golang.org/x/exp: 944ab1f → 746e56f
- golang.org/x/mod: v0.31.0 → v0.38.0
- golang.org/x/net: v0.49.0 → v0.58.0
- golang.org/x/oauth2: v0.34.0 → v0.36.0
- golang.org/x/sync: v0.19.0 → v0.22.0
- golang.org/x/sys: v0.40.0 → v0.47.0
- golang.org/x/term: v0.39.0 → v0.45.0
- golang.org/x/text: v0.33.0 → v0.41.0
- golang.org/x/time: v0.14.0 → v0.15.0
- golang.org/x/tools: v0.40.0 → v0.48.0
- gonum.org/v1/gonum: v0.16.0 → v0.17.0
- google.golang.org/genproto/googleapis/api: 8636f87 → 08b0e42
- google.golang.org/genproto/googleapis/rpc: 8636f87 → 08b0e42
- google.golang.org/grpc: v1.79.3 → v1.83.2
- google.golang.org/protobuf: f2248ac → v1.36.12
- k8s.io/api: v0.36.1 → v0.37.1
- k8s.io/apimachinery: v0.36.1 → v0.37.1
- k8s.io/apiserver: v0.36.1 → v0.37.1
- k8s.io/client-go: v0.36.1 → v0.37.1
- k8s.io/component-base: v0.36.1 → v0.37.1
- k8s.io/csi-translation-lib: v0.36.1 → v0.37.1
- k8s.io/gengo/v2: 85fd79d → ec3ebc5
- k8s.io/kms: v0.36.1 → v0.37.1
- k8s.io/kube-openapi: 43fb72c → d427ff9
- k8s.io/streaming: v0.36.1 → v0.37.1
- k8s.io/utils: b8788ab → be93311
- sigs.k8s.io/apiserver-network-proxy/konnectivity-client: v0.34.0 → v0.36.0
- sigs.k8s.io/structured-merge-diff/v6: v6.3.2 → v6.4.2

### Removed
- github.com/antihax/optional: [v1.0.0](https://github.com/antihax/optional/tree/v1.0.0)
- github.com/go-openapi/swag/jsonname: [v0.25.4](https://github.com/go-openapi/swag/tree/jsonname/v0.25.4)
- github.com/gogo/protobuf: [v1.3.2](https://github.com/gogo/protobuf/tree/v1.3.2)
- github.com/kisielk/errcheck: [v1.5.0](https://github.com/kisielk/errcheck/tree/v1.5.0)
- github.com/kisielk/gotool: [v1.0.0](https://github.com/kisielk/gotool/tree/v1.0.0)
- github.com/yuin/goldmark: [v1.2.1](https://github.com/yuin/goldmark/tree/v1.2.1)
- golang.org/x/xerrors: 5ec99f8

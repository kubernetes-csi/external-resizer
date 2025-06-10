## Release notes for 1.14.0

[Documentation](https://kubernetes-csi.github.io)

# Changelog since 1.13.0

### Feature

- Add a new `--automaxprocs` flag to set the `GOMAXPROCS` environment variable to match the configured Linux container CPU quota. (#499, @nixpanic)
- Serve additional leader election, work queue, process, and Go runtime metrics (#475, @AndrewSirenko)

### Bug or Regression

- Add VolumeModifyFailed PVC Event For Nonexistent VAC (#500, @sunnylovestiramisu)
- Fix: CVE-2025-22872 (#495, @andyzhangx)
- Fix panic when attempting to modify non-CSI volumes (#471, @AndrewSirenko)
- Upgrade csi-translation-lib to fix in-tree to CSI migration for Portworx volumes, in clusters where Portworx security feature is enabled (it's a Portworx feature, not Kubernetes feature). It required secret data from the secret mentioned in-tree SC, to be passed in CSI ControllerExpandVolume request which was not happening before this fix. (#480, @gohilankit)

## Dependencies
- Update kubernetes dependencies to v1.33.0 (#501, @Aishwarya-Hebbar)

### Added
- github.com/prashantv/gostub: [v1.1.0](https://github.com/prashantv/gostub/tree/v1.1.0)
- go.uber.org/automaxprocs: v1.6.0
- gopkg.in/go-jose/go-jose.v2: v2.6.3
- sigs.k8s.io/randfill: v1.0.0

### Changed
- cel.dev/expr: v0.18.0 → v0.19.1
- github.com/coreos/go-oidc: [v2.2.1+incompatible → v2.3.0+incompatible](https://github.com/coreos/go-oidc/compare/v2.2.1...v2.3.0)
- github.com/golang-jwt/jwt/v4: [v4.5.0 → v4.5.2](https://github.com/golang-jwt/jwt/compare/v4.5.0...v4.5.2)
- github.com/google/btree: [v1.0.1 → v1.1.3](https://github.com/google/btree/compare/v1.0.1...v1.1.3)
- github.com/google/cel-go: [v0.22.0 → v0.23.2](https://github.com/google/cel-go/compare/v0.22.0...v0.23.2)
- github.com/google/go-cmp: [v0.6.0 → v0.7.0](https://github.com/google/go-cmp/compare/v0.6.0...v0.7.0)
- github.com/google/gofuzz: [v1.2.0 → v1.0.0](https://github.com/google/gofuzz/compare/v1.2.0...v1.0.0)
- github.com/gorilla/websocket: [v1.5.0 → e064f32](https://github.com/gorilla/websocket/compare/v1.5.0...e064f32)
- github.com/grpc-ecosystem/grpc-gateway/v2: [v2.20.0 → v2.24.0](https://github.com/grpc-ecosystem/grpc-gateway/compare/v2.20.0...v2.24.0)
- github.com/klauspost/compress: [v1.17.11 → v1.18.0](https://github.com/klauspost/compress/compare/v1.17.11...v1.18.0)
- github.com/kubernetes-csi/csi-lib-utils: [v0.20.0 → v0.22.0](https://github.com/kubernetes-csi/csi-lib-utils/compare/v0.20.0...v0.22.0)
- github.com/prometheus/client_golang: [v1.20.5 → v1.22.0](https://github.com/prometheus/client_golang/compare/v1.20.5...v1.22.0)
- github.com/prometheus/common: [v0.61.0 → v0.62.0](https://github.com/prometheus/common/compare/v0.61.0...v0.62.0)
- github.com/stretchr/objx: [v0.1.0 → v0.5.2](https://github.com/stretchr/objx/compare/v0.1.0...v0.5.2)
- go.etcd.io/etcd/api/v3: v3.5.16 → v3.5.21
- go.etcd.io/etcd/client/pkg/v3: v3.5.16 → v3.5.21
- go.etcd.io/etcd/client/v2: v2.305.16 → v2.305.21
- go.etcd.io/etcd/client/v3: v3.5.16 → v3.5.21
- go.etcd.io/etcd/pkg/v3: v3.5.16 → v3.5.21
- go.etcd.io/etcd/raft/v3: v3.5.16 → v3.5.21
- go.etcd.io/etcd/server/v3: v3.5.16 → v3.5.21
- go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp: v0.53.0 → v0.58.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc: v1.27.0 → v1.33.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace: v1.28.0 → v1.33.0
- go.opentelemetry.io/otel/sdk: v1.31.0 → v1.33.0
- go.opentelemetry.io/proto/otlp: v1.3.1 → v1.4.0
- golang.org/x/crypto: v0.32.0 → v0.37.0
- golang.org/x/net: v0.34.0 → v0.39.0
- golang.org/x/oauth2: v0.25.0 → v0.27.0
- golang.org/x/sync: v0.10.0 → v0.13.0
- golang.org/x/sys: v0.29.0 → v0.32.0
- golang.org/x/term: v0.28.0 → v0.31.0
- golang.org/x/text: v0.21.0 → v0.24.0
- google.golang.org/genproto/googleapis/api: 796eee8 → e6fa225
- google.golang.org/protobuf: v1.36.2 → v1.36.5
- k8s.io/api: v0.32.0 → v0.33.0
- k8s.io/apimachinery: v0.32.0 → v0.33.0
- k8s.io/apiserver: v0.32.0 → v0.33.0
- k8s.io/client-go: v0.32.0 → v0.33.0
- k8s.io/component-base: v0.32.0 → v0.33.0
- k8s.io/csi-translation-lib: v0.32.0 → v0.33.0
- k8s.io/kms: v0.32.0 → v0.33.0
- k8s.io/kube-openapi: 2c72e55 → c8a335a
- sigs.k8s.io/apiserver-network-proxy/konnectivity-client: v0.31.0 → v0.31.2
- sigs.k8s.io/structured-merge-diff/v4: v4.5.0 → v4.6.0

### Removed
- github.com/asaskevich/govalidator: [f61b66f](https://github.com/asaskevich/govalidator/tree/f61b66f)
- github.com/go-kit/log: [v0.2.1](https://github.com/go-kit/log/tree/v0.2.1)
- github.com/go-logfmt/logfmt: [v0.5.1](https://github.com/go-logfmt/logfmt/tree/v0.5.1)
- gopkg.in/square/go-jose.v2: v2.6.0

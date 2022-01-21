# Release notes for 1.4.0

[Documentation](https://kubernetes-csi.github.io)

# Changelog since 1.3.0

## Changes by Kind

### Bug or Regression

- Fix unnecessary warning when PV bound to PVC is not created yet. ([#172](https://github.com/kubernetes-csi/external-resizer/pull/172), [@takuhiro](https://github.com/takuhiro))

### Feature
- Add support to allow users to recover from expansion failures ([#187](https://github.com/kubernetes-csi/external-resizer/pull/187), [@gnufied](https://github.com/gnufied))


### Other (Cleanup or Flake)

- Kubernetes client-go dependencies are updated to v1.23.1 ([#186](https://github.com/kubernetes-csi/external-resizer/pull/186), [@humblec](https://github.com/humblec))
- Updates Kubernetes dependencies to v1.23.0 ([#184](https://github.com/kubernetes-csi/external-resizer/pull/184), [@humblec](https://github.com/humblec))

## Dependencies

### Added
- github.com/cncf/xds/go: [fbca930](https://github.com/cncf/xds/go/tree/fbca930)
- github.com/getkin/kin-openapi: [v0.76.0](https://github.com/getkin/kin-openapi/tree/v0.76.0)
- github.com/go-logr/zapr: [v1.2.0](https://github.com/go-logr/zapr/tree/v1.2.0)
- github.com/gorilla/mux: [v1.8.0](https://github.com/gorilla/mux/tree/v1.8.0)
- github.com/kr/fs: [v0.1.0](https://github.com/kr/fs/tree/v0.1.0)
- github.com/pkg/sftp: [v1.10.1](https://github.com/pkg/sftp/tree/v1.10.1)
- sigs.k8s.io/json: c049b76

### Changed
- cloud.google.com/go: v0.65.0 → v0.81.0
- github.com/benbjohnson/clock: [v1.0.3 → v1.1.0](https://github.com/benbjohnson/clock/compare/v1.0.3...v1.1.0)
- github.com/bketelsen/crypt: [5cbc8cc → v0.0.4](https://github.com/bketelsen/crypt/compare/5cbc8cc...v0.0.4)
- github.com/envoyproxy/go-control-plane: [668b12f → 63b5d3c](https://github.com/envoyproxy/go-control-plane/compare/668b12f...63b5d3c)
- github.com/evanphx/json-patch: [v4.11.0+incompatible → v4.12.0+incompatible](https://github.com/evanphx/json-patch/compare/v4.11.0...v4.12.0)
- github.com/go-logr/logr: [v0.4.0 → v1.2.0](https://github.com/go-logr/logr/compare/v0.4.0...v1.2.0)
- github.com/golang/mock: [v1.4.4 → v1.5.0](https://github.com/golang/mock/compare/v1.4.4...v1.5.0)
- github.com/google/martian/v3: [v3.0.0 → v3.1.0](https://github.com/google/martian/v3/compare/v3.0.0...v3.1.0)
- github.com/google/pprof: [1a94d86 → cbba55b](https://github.com/google/pprof/compare/1a94d86...cbba55b)
- github.com/ianlancetaylor/demangle: [5e5cf60 → 28f6c0f](https://github.com/ianlancetaylor/demangle/compare/5e5cf60...28f6c0f)
- github.com/json-iterator/go: [v1.1.11 → v1.1.12](https://github.com/json-iterator/go/compare/v1.1.11...v1.1.12)
- github.com/magiconair/properties: [v1.8.1 → v1.8.5](https://github.com/magiconair/properties/compare/v1.8.1...v1.8.5)
- github.com/mitchellh/mapstructure: [v1.1.2 → v1.4.1](https://github.com/mitchellh/mapstructure/compare/v1.1.2...v1.4.1)
- github.com/modern-go/reflect2: [v1.0.1 → v1.0.2](https://github.com/modern-go/reflect2/compare/v1.0.1...v1.0.2)
- github.com/pelletier/go-toml: [v1.2.0 → v1.9.3](https://github.com/pelletier/go-toml/compare/v1.2.0...v1.9.3)
- github.com/prometheus/common: [v0.26.0 → v0.28.0](https://github.com/prometheus/common/compare/v0.26.0...v0.28.0)
- github.com/spf13/afero: [v1.2.2 → v1.6.0](https://github.com/spf13/afero/compare/v1.2.2...v1.6.0)
- github.com/spf13/cast: [v1.3.0 → v1.3.1](https://github.com/spf13/cast/compare/v1.3.0...v1.3.1)
- github.com/spf13/cobra: [v1.1.3 → v1.2.1](https://github.com/spf13/cobra/compare/v1.1.3...v1.2.1)
- github.com/spf13/jwalterweatherman: [v1.0.0 → v1.1.0](https://github.com/spf13/jwalterweatherman/compare/v1.0.0...v1.1.0)
- github.com/spf13/viper: [v1.7.0 → v1.8.1](https://github.com/spf13/viper/compare/v1.7.0...v1.8.1)
- github.com/yuin/goldmark: [v1.3.5 → v1.4.0](https://github.com/yuin/goldmark/compare/v1.3.5...v1.4.0)
- go.opencensus.io: v0.22.4 → v0.23.0
- go.uber.org/zap: v1.17.0 → v1.19.0
- golang.org/x/crypto: 5ea612d → 32db794
- golang.org/x/net: 37e1c6a → 491a49a
- golang.org/x/oauth2: cd4f82c → 2bc19b1
- golang.org/x/sys: 59db8d7 → f4d4317
- golang.org/x/term: de623e6 → 6886f2d
- golang.org/x/text: v0.3.6 → v0.3.7
- golang.org/x/tools: v0.1.2 → d4cc65f
- google.golang.org/api: v0.30.0 → v0.44.0
- google.golang.org/genproto: f16073e → fe13028
- google.golang.org/grpc: v1.38.0 → v1.40.0
- google.golang.org/protobuf: v1.26.0 → v1.27.1
- gopkg.in/ini.v1: v1.51.0 → v1.62.0
- k8s.io/api: v0.22.0 → v0.23.1
- k8s.io/apimachinery: v0.22.0 → v0.23.1
- k8s.io/apiserver: v0.22.0 → v0.23.1
- k8s.io/client-go: v0.22.0 → v0.23.1
- k8s.io/component-base: v0.22.0 → v0.23.1
- k8s.io/csi-translation-lib: v0.22.0 → v0.23.1
- k8s.io/gengo: 3a45101 → 485abfe
- k8s.io/klog/v2: v2.9.0 → v2.30.0
- k8s.io/kube-openapi: 9528897 → e816edb
- k8s.io/utils: 4b05e18 → cb0fa31
- sigs.k8s.io/apiserver-network-proxy/konnectivity-client: v0.0.22 → v0.0.25

### Removed
_Nothing has changed._

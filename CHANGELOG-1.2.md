# Release notes for 1.2.0

[Documentation](https://kubernetes-csi.github.io)

# Changelog since 1.1.0

## Changes by Kind

### Other (Cleanup or Flake)

- Fix a bug that when CSI migration is enabled and PV is using GA topology label, resizing is not work. ([#139](https://github.com/kubernetes-csi/external-resizer/pull/139), [@Jiawei0227](https://github.com/Jiawei0227))

### Feature

- Add a feature-gate to allow users to restore size of deleted PVCs ([#140](https://github.com/kubernetes-csi/external-resizer/pull/140), [@sunpa93](https://github.com/sunpa93))

### Bug fixes

- Set the value of "migrated" field in the metrics to true or false to indicate if the call is a migration enabled feature or not ([#138](https://github.com/kubernetes-csi/external-resizer/pull/138), [@nearora-msft](https://github.com/nearora-msft))
- Updated runtime (Go 1.16) and dependencies ([#141](https://github.com/kubernetes-csi/external-resizer/pull/141), [@pohly](https://github.com/pohly))

## Dependencies

### Added
- github.com/coreos/go-oidc: [v2.1.0+incompatible](https://github.com/coreos/go-oidc/tree/v2.1.0)
- github.com/moby/spdystream: [v0.2.0](https://github.com/moby/spdystream/tree/v0.2.0)
- github.com/niemeyer/pretty: [a10e7ca](https://github.com/niemeyer/pretty/tree/a10e7ca)
- github.com/pquerna/cachecontrol: [0dec1b3](https://github.com/pquerna/cachecontrol/tree/0dec1b3)
- gopkg.in/natefinch/lumberjack.v2: v2.0.0
- gopkg.in/square/go-jose.v2: v2.2.2
- k8s.io/apiserver: v0.20.0
- sigs.k8s.io/apiserver-network-proxy/konnectivity-client: v0.0.14

### Changed
- github.com/Azure/go-autorest/autorest: [v0.11.1 → v0.11.12](https://github.com/Azure/go-autorest/autorest/compare/v0.11.1...v0.11.12)
- github.com/cncf/udpa/go: [efcf912 → 5459f2c](https://github.com/cncf/udpa/go/compare/efcf912...5459f2c)
- github.com/container-storage-interface/spec: [v1.3.0 → v1.4.0](https://github.com/container-storage-interface/spec/compare/v1.3.0...v1.4.0)
- github.com/coreos/go-semver: [v0.2.0 → v0.3.0](https://github.com/coreos/go-semver/compare/v0.2.0...v0.3.0)
- github.com/coreos/go-systemd: [39ca1b0 → 95778df](https://github.com/coreos/go-systemd/compare/39ca1b0...95778df)
- github.com/coreos/pkg: [3ac0863 → 399ea9e](https://github.com/coreos/pkg/compare/3ac0863...399ea9e)
- github.com/creack/pty: [v1.1.7 → v1.1.11](https://github.com/creack/pty/compare/v1.1.7...v1.1.11)
- github.com/dustin/go-humanize: [bb3d318 → v1.0.0](https://github.com/dustin/go-humanize/compare/bb3d318...v1.0.0)
- github.com/emicklei/go-restful: [ff4f55a → v2.9.5+incompatible](https://github.com/emicklei/go-restful/compare/ff4f55a...v2.9.5)
- github.com/envoyproxy/go-control-plane: [v0.9.7 → fd9021f](https://github.com/envoyproxy/go-control-plane/compare/v0.9.7...fd9021f)
- github.com/fsnotify/fsnotify: [v1.4.9 → v1.4.7](https://github.com/fsnotify/fsnotify/compare/v1.4.9...v1.4.7)
- github.com/go-logr/logr: [v0.3.0 → v0.4.0](https://github.com/go-logr/logr/compare/v0.3.0...v0.4.0)
- github.com/gogo/protobuf: [v1.3.1 → v1.3.2](https://github.com/gogo/protobuf/compare/v1.3.1...v1.3.2)
- github.com/golang/protobuf: [v1.4.3 → v1.5.1](https://github.com/golang/protobuf/compare/v1.4.3...v1.5.1)
- github.com/google/go-cmp: [v0.5.4 → v0.5.5](https://github.com/google/go-cmp/compare/v0.5.4...v0.5.5)
- github.com/googleapis/gnostic: [v0.5.3 → v0.5.4](https://github.com/googleapis/gnostic/compare/v0.5.3...v0.5.4)
- github.com/gorilla/websocket: [4201258 → v1.4.2](https://github.com/gorilla/websocket/compare/4201258...v1.4.2)
- github.com/imdario/mergo: [v0.3.11 → v0.3.12](https://github.com/imdario/mergo/compare/v0.3.11...v0.3.12)
- github.com/kisielk/errcheck: [v1.2.0 → v1.5.0](https://github.com/kisielk/errcheck/compare/v1.2.0...v1.5.0)
- github.com/kr/text: [v0.1.0 → v0.2.0](https://github.com/kr/text/compare/v0.1.0...v0.2.0)
- github.com/kubernetes-csi/csi-lib-utils: [v0.9.0 → v0.9.1](https://github.com/kubernetes-csi/csi-lib-utils/compare/v0.9.0...v0.9.1)
- github.com/mailru/easyjson: [b2ccc51 → v0.7.0](https://github.com/mailru/easyjson/compare/b2ccc51...v0.7.0)
- github.com/moby/term: [672ec06 → df9cb8a](https://github.com/moby/term/compare/672ec06...df9cb8a)
- github.com/munnerz/goautoneg: [a547fc6 → a7dc8b6](https://github.com/munnerz/goautoneg/compare/a547fc6...a7dc8b6)
- github.com/prometheus/common: [v0.15.0 → v0.19.0](https://github.com/prometheus/common/compare/v0.15.0...v0.19.0)
- github.com/prometheus/procfs: [v0.2.0 → v0.6.0](https://github.com/prometheus/procfs/compare/v0.2.0...v0.6.0)
- github.com/tmc/grpc-websocket-proxy: [89b8d40 → 0ad062e](https://github.com/tmc/grpc-websocket-proxy/compare/89b8d40...0ad062e)
- github.com/yuin/goldmark: [v1.1.32 → v1.2.1](https://github.com/yuin/goldmark/compare/v1.1.32...v1.2.1)
- go.etcd.io/bbolt: v1.3.3 → v1.3.5
- go.etcd.io/etcd: 3cf2f69 → dd1b699
- golang.org/x/crypto: 9d13527 → 5ea612d
- golang.org/x/net: 986b41b → d523dce
- golang.org/x/oauth2: 08078c5 → cd4f82c
- golang.org/x/sync: 6e8e738 → 09787c9
- golang.org/x/sys: f9fddec → c4fcb01
- golang.org/x/term: 2321bbc → de623e6
- golang.org/x/text: v0.3.4 → v0.3.5
- golang.org/x/time: 7e3f01d → f8bda1e
- golang.org/x/tools: b303f43 → 113979e
- google.golang.org/genproto: 8c77b98 → 75c7a85
- google.golang.org/grpc: v1.34.0 → v1.36.0
- google.golang.org/protobuf: v1.25.0 → v1.26.0
- gopkg.in/check.v1: 41f04d3 → 8fa4692
- gopkg.in/yaml.v3: eeeca48 → 496545a
- gotest.tools/v3: v3.0.2 → v3.0.3
- k8s.io/api: v0.20.0 → v0.21.0
- k8s.io/apimachinery: v0.21.0-alpha.0 → v0.21.0
- k8s.io/client-go: v0.20.0 → v0.21.0
- k8s.io/component-base: v0.20.0 → v0.21.0
- k8s.io/csi-translation-lib: v0.20.0 → v0.21.0
- k8s.io/klog/v2: v2.4.0 → v2.8.0
- k8s.io/kube-openapi: d219536 → f622666
- k8s.io/utils: 67b214c → 2afb431
- sigs.k8s.io/structured-merge-diff/v4: v4.0.2 → v4.1.0

### Removed
- github.com/docker/spdystream: [449fdfc](https://github.com/docker/spdystream/tree/449fdfc)
- gotest.tools: v2.2.0+incompatible

# Release notes for 1.5.0

[Documentation](https://kubernetes-csi.github.io)

# Changelog since 1.4.0

## Changes by Kind

### ### Other (Cleanup or Flake)

- Kube client dependencies are updated to v1.24.0 ([#204](https://github.com/kubernetes-csi/external-resizer/pull/204), [@humblec](https://github.com/humblec))

## Dependencies

### Added
- github.com/armon/go-socks5: [e753329](https://github.com/armon/go-socks5/tree/e753329)
- github.com/blang/semver/v4: [v4.0.0](https://github.com/blang/semver/v4/tree/v4.0.0)
- github.com/google/gnostic: [v0.5.7-v3refs](https://github.com/google/gnostic/tree/v0.5.7-v3refs)

### Changed
- github.com/bketelsen/crypt: [v0.0.4 → 5cbc8cc](https://github.com/bketelsen/crypt/compare/v0.0.4...5cbc8cc)
- github.com/cespare/xxhash/v2: [v2.1.1 → v2.1.2](https://github.com/cespare/xxhash/v2/compare/v2.1.1...v2.1.2)
- github.com/cpuguy83/go-md2man/v2: [v2.0.0 → v2.0.1](https://github.com/cpuguy83/go-md2man/v2/compare/v2.0.0...v2.0.1)
- github.com/kubernetes-csi/csi-lib-utils: [v0.10.0 → v0.11.0](https://github.com/kubernetes-csi/csi-lib-utils/compare/v0.10.0...v0.11.0)
- github.com/magiconair/properties: [v1.8.5 → v1.8.1](https://github.com/magiconair/properties/compare/v1.8.5...v1.8.1)
- github.com/mitchellh/mapstructure: [v1.4.1 → v1.1.2](https://github.com/mitchellh/mapstructure/compare/v1.4.1...v1.1.2)
- github.com/moby/term: [9d4ed18 → 3f7ff69](https://github.com/moby/term/compare/9d4ed18...3f7ff69)
- github.com/pelletier/go-toml: [v1.9.3 → v1.2.0](https://github.com/pelletier/go-toml/compare/v1.9.3...v1.2.0)
- github.com/prometheus/client_golang: [v1.11.0 → v1.12.1](https://github.com/prometheus/client_golang/compare/v1.11.0...v1.12.1)
- github.com/prometheus/common: [v0.28.0 → v0.32.1](https://github.com/prometheus/common/compare/v0.28.0...v0.32.1)
- github.com/prometheus/procfs: [v0.6.0 → v0.7.3](https://github.com/prometheus/procfs/compare/v0.6.0...v0.7.3)
- github.com/russross/blackfriday/v2: [v2.0.1 → v2.1.0](https://github.com/russross/blackfriday/v2/compare/v2.0.1...v2.1.0)
- github.com/spf13/cast: [v1.3.1 → v1.3.0](https://github.com/spf13/cast/compare/v1.3.1...v1.3.0)
- github.com/spf13/cobra: [v1.2.1 → v1.4.0](https://github.com/spf13/cobra/compare/v1.2.1...v1.4.0)
- github.com/spf13/jwalterweatherman: [v1.1.0 → v1.0.0](https://github.com/spf13/jwalterweatherman/compare/v1.1.0...v1.0.0)
- github.com/spf13/viper: [v1.8.1 → v1.7.0](https://github.com/spf13/viper/compare/v1.8.1...v1.7.0)
- github.com/yuin/goldmark: [v1.4.0 → v1.4.1](https://github.com/yuin/goldmark/compare/v1.4.0...v1.4.1)
- go.etcd.io/etcd/api/v3: v3.5.0 → v3.5.1
- go.etcd.io/etcd/client/pkg/v3: v3.5.0 → v3.5.1
- go.etcd.io/etcd/client/v3: v3.5.0 → v3.5.1
- golang.org/x/crypto: 32db794 → 8634188
- golang.org/x/mod: v0.4.2 → 9b9b3d8
- golang.org/x/net: 491a49a → cd36cc0
- golang.org/x/oauth2: 2bc19b1 → d3ed0bb
- golang.org/x/sys: f4d4317 → 3681064
- golang.org/x/term: 6886f2d → 03fcf44
- golang.org/x/time: 1f47c86 → 90d013b
- golang.org/x/tools: d4cc65f → 897bd77
- google.golang.org/api: v0.44.0 → v0.43.0
- google.golang.org/genproto: fe13028 → 42d7afd
- gopkg.in/ini.v1: v1.62.0 → v1.51.0
- k8s.io/api: v0.23.1 → v0.24.0
- k8s.io/apimachinery: v0.23.1 → v0.24.0
- k8s.io/apiserver: v0.23.1 → v0.24.0
- k8s.io/client-go: v0.23.1 → v0.24.0
- k8s.io/component-base: v0.23.1 → v0.24.0
- k8s.io/csi-translation-lib: v0.23.1 → v0.24.0
- k8s.io/klog/v2: v2.30.0 → v2.60.1
- k8s.io/kube-openapi: e816edb → 3ee0da9
- k8s.io/utils: cb0fa31 → 3a6ce19
- sigs.k8s.io/apiserver-network-proxy/konnectivity-client: v0.0.25 → v0.0.30
- sigs.k8s.io/json: c049b76 → 9f7c6b3
- sigs.k8s.io/structured-merge-diff/v4: v4.1.2 → v4.2.1

### Removed
- github.com/blang/semver: [v3.5.1+incompatible](https://github.com/blang/semver/tree/v3.5.1)
- github.com/googleapis/gnostic: [v0.5.5](https://github.com/googleapis/gnostic/tree/v0.5.5)

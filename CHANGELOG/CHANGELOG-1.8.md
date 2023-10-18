# Release notes for 1.8.1

[Documentation](https://kubernetes-csi.github.io)

## Changes by Kind

### Uncategorized

- CVE fixes: CVE-2023-44487 ([#343](https://github.com/kubernetes-csi/external-resizer/pull/343), [@dannawang0221](https://github.com/dannawang0221)

## Dependencies

### Added
_Nothing has changed._

### Changed
- golang.org/x/crypto: v0.1.0 → v0.14.0
- golang.org/x/net: v0.8.0 → v0.17.0
- golang.org/x/sys: v0.6.0 → v0.13.0
- golang.org/x/term: v0.6.0 → v0.13.0
- golang.org/x/text: v0.8.0 → v0.13.0

### Removed
_Nothing has changed._

# Release notes for 1.8.0

# Changelog since 1.7.0

## Changes by Kind

## Dependencies

### Added
- github.com/go-task/slim-sprig: [348f09d](https://github.com/go-task/slim-sprig/tree/348f09d)
- github.com/golang-jwt/jwt/v4: [v4.4.2](https://github.com/golang-jwt/jwt/v4/tree/v4.4.2)

### Changed
- github.com/coreos/go-systemd/v22: [v22.3.2 → v22.4.0](https://github.com/coreos/go-systemd/v22/compare/v22.3.2...v22.4.0)
- github.com/go-openapi/jsonpointer: [v0.19.5 → v0.19.6](https://github.com/go-openapi/jsonpointer/compare/v0.19.5...v0.19.6)
- github.com/go-openapi/jsonreference: [v0.20.0 → v0.20.1](https://github.com/go-openapi/jsonreference/compare/v0.20.0...v0.20.1)
- github.com/go-openapi/swag: [v0.19.14 → v0.22.3](https://github.com/go-openapi/swag/compare/v0.19.14...v0.22.3)
- github.com/golang/protobuf: [v1.5.2 → v1.5.3](https://github.com/golang/protobuf/compare/v1.5.2...v1.5.3)
- github.com/google/cel-go: [v0.12.5 → v0.12.6](https://github.com/google/cel-go/compare/v0.12.5...v0.12.6)
- github.com/google/pprof: [1a94d86 → 4bb14d4](https://github.com/google/pprof/compare/1a94d86...4bb14d4)
- github.com/google/uuid: [v1.1.2 → v1.3.0](https://github.com/google/uuid/compare/v1.1.2...v1.3.0)
- github.com/kr/pretty: [v0.2.0 → v0.3.0](https://github.com/kr/pretty/compare/v0.2.0...v0.3.0)
- github.com/kubernetes-csi/csi-lib-utils: [v0.12.0 → v0.13.0](https://github.com/kubernetes-csi/csi-lib-utils/compare/v0.12.0...v0.13.0)
- github.com/mailru/easyjson: [v0.7.6 → v0.7.7](https://github.com/mailru/easyjson/compare/v0.7.6...v0.7.7)
- github.com/mitchellh/mapstructure: [v1.1.2 → v1.4.1](https://github.com/mitchellh/mapstructure/compare/v1.1.2...v1.4.1)
- github.com/moby/term: [39b0c02 → 1aeaba8](https://github.com/moby/term/compare/39b0c02...1aeaba8)
- github.com/onsi/ginkgo/v2: [v2.4.0 → v2.9.1](https://github.com/onsi/ginkgo/v2/compare/v2.4.0...v2.9.1)
- github.com/onsi/gomega: [v1.23.0 → v1.27.4](https://github.com/onsi/gomega/compare/v1.23.0...v1.27.4)
- github.com/rogpeppe/go-internal: [v1.3.0 → v1.10.0](https://github.com/rogpeppe/go-internal/compare/v1.3.0...v1.10.0)
- github.com/sirupsen/logrus: [v1.8.1 → v1.9.0](https://github.com/sirupsen/logrus/compare/v1.8.1...v1.9.0)
- github.com/stretchr/objx: [v0.1.1 → v0.5.0](https://github.com/stretchr/objx/compare/v0.1.1...v0.5.0)
- github.com/stretchr/testify: [v1.8.0 → v1.8.1](https://github.com/stretchr/testify/compare/v1.8.0...v1.8.1)
- github.com/tmc/grpc-websocket-proxy: [e5319fd → 673ab2c](https://github.com/tmc/grpc-websocket-proxy/compare/e5319fd...673ab2c)
- go.etcd.io/etcd/api/v3: v3.5.5 → v3.5.7
- go.etcd.io/etcd/client/pkg/v3: v3.5.5 → v3.5.7
- go.etcd.io/etcd/client/v2: v2.305.5 → v2.305.7
- go.etcd.io/etcd/client/v3: v3.5.5 → v3.5.7
- go.etcd.io/etcd/pkg/v3: v3.5.5 → v3.5.7
- go.etcd.io/etcd/raft/v3: v3.5.5 → v3.5.7
- go.etcd.io/etcd/server/v3: v3.5.5 → v3.5.7
- go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp: v0.35.0 → v0.35.1
- go.uber.org/goleak: v1.2.0 → v1.2.1
- golang.org/x/mod: 86c51ed → v0.8.0
- golang.org/x/net: v0.4.0 → v0.8.0
- golang.org/x/sync: 886fb93 → v0.1.0
- golang.org/x/sys: v0.3.0 → v0.6.0
- golang.org/x/term: v0.3.0 → v0.6.0
- golang.org/x/text: v0.5.0 → v0.8.0
- golang.org/x/tools: v0.1.12 → v0.7.0
- golang.org/x/xerrors: 5ec99f8 → 04be3eb
- gopkg.in/check.v1: 8fa4692 → 10cb982
- gopkg.in/square/go-jose.v2: v2.2.2 → v2.6.0
- k8s.io/api: v0.26.0 → v0.27.0
- k8s.io/apimachinery: v0.26.0 → v0.27.0
- k8s.io/apiserver: v0.26.0 → v0.27.0
- k8s.io/client-go: v0.26.0 → v0.27.0
- k8s.io/component-base: v0.26.0 → v0.27.0
- k8s.io/csi-translation-lib: v0.26.0 → v0.27.0
- k8s.io/klog/v2: v2.80.1 → v2.90.1
- k8s.io/kms: v0.26.0 → v0.27.0
- k8s.io/kube-openapi: 172d655 → 15aac26
- k8s.io/utils: 1a15be2 → a36077c
- sigs.k8s.io/apiserver-network-proxy/konnectivity-client: v0.0.33 → v0.1.1
- sigs.k8s.io/json: f223a00 → bc3834c

### Removed
- github.com/PuerkitoBio/purell: [v1.1.1](https://github.com/PuerkitoBio/purell/tree/v1.1.1)
- github.com/PuerkitoBio/urlesc: [de5bf2a](https://github.com/PuerkitoBio/urlesc/tree/de5bf2a)
- github.com/elazarl/goproxy: [947c36d](https://github.com/elazarl/goproxy/tree/947c36d)
- github.com/form3tech-oss/jwt-go: [v3.2.3+incompatible](https://github.com/form3tech-oss/jwt-go/tree/v3.2.3)
- github.com/niemeyer/pretty: [a10e7ca](https://github.com/niemeyer/pretty/tree/a10e7ca)
- gotest.tools/v3: v3.0.3

# Release notes for 1.0.0

[Documentation](https://kubernetes-csi.github.io)

# Changelog since v0.5.0


## Urgent Upgrade Notes 

### (No, really, you MUST read this before you upgrade)

- external-resizer will not call controller-expand if volume is in-use and can not be expanded while in-use. New RBAC rules for listing pods are required. ([#86](https://github.com/kubernetes-csi/external-resizer/pull/86), [@gnufied](https://github.com/gnufied)) . This behaviour is enabled by default but can be disabled by setting `handle-volume-inuse-error` to `false`, which would allow external-resizer to perform volume expansion without checking if volume being expanded is in-use or not([#89](https://github.com/kubernetes-csi/external-resizer/pull/89), [@saikat-royc](https://github.com/saikat-royc)). If CSI driver being used supports online expansion, it might be desirable to set `handle-volume-inuse-error` to `false` - to save costs associated with watching all pods in the cluster.
- The `--csiTimeout` flag of external-resizer is obsoleted with this release. Instead `--timeout` flag has been introduced for the same purpose. This timeout is used by external-resizer to decide how long it should wait for a response from CSI driver. `timeout` value defaults to 10 seconds. From now on, the deployment  templates or  artifacts has to make use of `--timeout` flag instead of obsoleted `--csiTimeout`  flag. ([#82](https://github.com/kubernetes-csi/external-resizer/pull/82), [@humblec](https://github.com/humblec))


## Changes by Kind

### Feature

- Adding configurable QPS and Burst to k8s client and use a separate client for leader election. ([#91](https://github.com/kubernetes-csi/external-resizer/pull/91), [@RaunakShah](https://github.com/RaunakShah))

### Bug or Regression

- Ensure that we do not resize recently resized volumes ([#96](https://github.com/kubernetes-csi/external-resizer/pull/96), [@gnufied](https://github.com/gnufied))

### Uncategorized

- Build with Go 1.15 ([#98](https://github.com/kubernetes-csi/external-resizer/pull/98), [@pohly](https://github.com/pohly))
- Publishing of images on k8s.gcr.io ([#84](https://github.com/kubernetes-csi/external-resizer/pull/84), [@pohly](https://github.com/pohly))

## Dependencies

### Added
- cloud.google.com/go/bigquery: v1.0.1
- cloud.google.com/go/datastore: v1.0.0
- cloud.google.com/go/pubsub: v1.0.1
- cloud.google.com/go/storage: v1.0.0
- dmitri.shuralyov.com/gpu/mtl: 666a987
- github.com/Azure/go-ansiterm: [d6e3b33](https://github.com/Azure/go-ansiterm/tree/d6e3b33)
- github.com/BurntSushi/xgb: [27f1227](https://github.com/BurntSushi/xgb/tree/27f1227)
- github.com/cespare/xxhash/v2: [v2.1.1](https://github.com/cespare/xxhash/v2/tree/v2.1.1)
- github.com/chzyer/logex: [v1.1.10](https://github.com/chzyer/logex/tree/v1.1.10)
- github.com/chzyer/readline: [2972be2](https://github.com/chzyer/readline/tree/2972be2)
- github.com/chzyer/test: [a1ea475](https://github.com/chzyer/test/tree/a1ea475)
- github.com/cncf/udpa/go: [269d4d4](https://github.com/cncf/udpa/go/tree/269d4d4)
- github.com/docopt/docopt-go: [ee0de3b](https://github.com/docopt/docopt-go/tree/ee0de3b)
- github.com/go-gl/glfw/v3.3/glfw: [12ad95a](https://github.com/go-gl/glfw/v3.3/glfw/tree/12ad95a)
- github.com/google/renameio: [v0.1.0](https://github.com/google/renameio/tree/v0.1.0)
- github.com/ianlancetaylor/demangle: [5e5cf60](https://github.com/ianlancetaylor/demangle/tree/5e5cf60)
- github.com/moby/term: [672ec06](https://github.com/moby/term/tree/672ec06)
- github.com/rogpeppe/go-internal: [v1.3.0](https://github.com/rogpeppe/go-internal/tree/v1.3.0)
- go.uber.org/atomic: v1.4.0
- go.uber.org/multierr: v1.1.0
- go.uber.org/zap: v1.10.0
- golang.org/x/image: cff245a
- golang.org/x/mobile: d2bd2a2
- golang.org/x/mod: c90efee
- golang.org/x/xerrors: 9bdfabe
- google.golang.org/protobuf: v1.24.0
- gopkg.in/errgo.v2: v2.1.0
- gotest.tools/v3: v3.0.2
- gotest.tools: v2.2.0+incompatible
- k8s.io/klog/v2: v2.2.0
- rsc.io/binaryregexp: v0.2.0
- sigs.k8s.io/structured-merge-diff/v4: v4.0.1

### Changed
- cloud.google.com/go: v0.38.0 → v0.51.0
- github.com/Azure/go-autorest/autorest/adal: [v0.5.0 → v0.8.2](https://github.com/Azure/go-autorest/autorest/adal/compare/v0.5.0...v0.8.2)
- github.com/Azure/go-autorest/autorest/date: [v0.1.0 → v0.2.0](https://github.com/Azure/go-autorest/autorest/date/compare/v0.1.0...v0.2.0)
- github.com/Azure/go-autorest/autorest/mocks: [v0.2.0 → v0.3.0](https://github.com/Azure/go-autorest/autorest/mocks/compare/v0.2.0...v0.3.0)
- github.com/Azure/go-autorest/autorest: [v0.9.0 → v0.9.6](https://github.com/Azure/go-autorest/autorest/compare/v0.9.0...v0.9.6)
- github.com/alecthomas/template: [a0175ee → fb15b89](https://github.com/alecthomas/template/compare/a0175ee...fb15b89)
- github.com/alecthomas/units: [2efee85 → c3de453](https://github.com/alecthomas/units/compare/2efee85...c3de453)
- github.com/beorn7/perks: [v1.0.0 → v1.0.1](https://github.com/beorn7/perks/compare/v1.0.0...v1.0.1)
- github.com/elazarl/goproxy: [c4fc265 → 947c36d](https://github.com/elazarl/goproxy/compare/c4fc265...947c36d)
- github.com/envoyproxy/go-control-plane: [5f8ba28 → v0.9.4](https://github.com/envoyproxy/go-control-plane/compare/5f8ba28...v0.9.4)
- github.com/evanphx/json-patch: [v4.5.0+incompatible → v4.9.0+incompatible](https://github.com/evanphx/json-patch/compare/v4.5.0...v4.9.0)
- github.com/fsnotify/fsnotify: [v1.4.7 → v1.4.9](https://github.com/fsnotify/fsnotify/compare/v1.4.7...v1.4.9)
- github.com/go-kit/kit: [v0.8.0 → v0.9.0](https://github.com/go-kit/kit/compare/v0.8.0...v0.9.0)
- github.com/go-logfmt/logfmt: [v0.3.0 → v0.4.0](https://github.com/go-logfmt/logfmt/compare/v0.3.0...v0.4.0)
- github.com/go-logr/logr: [v0.1.0 → v0.2.0](https://github.com/go-logr/logr/compare/v0.1.0...v0.2.0)
- github.com/gogo/protobuf: [65acae2 → v1.3.1](https://github.com/gogo/protobuf/compare/65acae2...v1.3.1)
- github.com/golang/groupcache: [5b532d6 → 215e871](https://github.com/golang/groupcache/compare/5b532d6...215e871)
- github.com/golang/mock: [v1.2.0 → v1.3.1](https://github.com/golang/mock/compare/v1.2.0...v1.3.1)
- github.com/golang/protobuf: [v1.3.2 → v1.4.2](https://github.com/golang/protobuf/compare/v1.3.2...v1.4.2)
- github.com/google/go-cmp: [v0.3.0 → v0.4.0](https://github.com/google/go-cmp/compare/v0.3.0...v0.4.0)
- github.com/google/gofuzz: [v1.0.0 → v1.1.0](https://github.com/google/gofuzz/compare/v1.0.0...v1.1.0)
- github.com/google/pprof: [3ea8567 → d4f498a](https://github.com/google/pprof/compare/3ea8567...d4f498a)
- github.com/googleapis/gax-go/v2: [v2.0.4 → v2.0.5](https://github.com/googleapis/gax-go/v2/compare/v2.0.4...v2.0.5)
- github.com/googleapis/gnostic: [v0.2.0 → v0.4.1](https://github.com/googleapis/gnostic/compare/v0.2.0...v0.4.1)
- github.com/imdario/mergo: [v0.3.7 → v0.3.9](https://github.com/imdario/mergo/compare/v0.3.7...v0.3.9)
- github.com/json-iterator/go: [v1.1.8 → v1.1.10](https://github.com/json-iterator/go/compare/v1.1.8...v1.1.10)
- github.com/jstemmer/go-junit-report: [af01ea7 → v0.9.1](https://github.com/jstemmer/go-junit-report/compare/af01ea7...v0.9.1)
- github.com/konsorten/go-windows-terminal-sequences: [v1.0.1 → v1.0.3](https://github.com/konsorten/go-windows-terminal-sequences/compare/v1.0.1...v1.0.3)
- github.com/kr/pretty: [v0.1.0 → v0.2.0](https://github.com/kr/pretty/compare/v0.1.0...v0.2.0)
- github.com/matttproud/golang_protobuf_extensions: [v1.0.1 → c182aff](https://github.com/matttproud/golang_protobuf_extensions/compare/v1.0.1...c182aff)
- github.com/onsi/ginkgo: [v1.10.2 → v1.11.0](https://github.com/onsi/ginkgo/compare/v1.10.2...v1.11.0)
- github.com/pkg/errors: [v0.8.1 → v0.9.1](https://github.com/pkg/errors/compare/v0.8.1...v0.9.1)
- github.com/prometheus/client_golang: [v1.0.0 → v1.7.1](https://github.com/prometheus/client_golang/compare/v1.0.0...v1.7.1)
- github.com/prometheus/client_model: [14fe0d1 → v0.2.0](https://github.com/prometheus/client_model/compare/14fe0d1...v0.2.0)
- github.com/prometheus/common: [v0.4.1 → v0.10.0](https://github.com/prometheus/common/compare/v0.4.1...v0.10.0)
- github.com/prometheus/procfs: [v0.0.2 → v0.1.3](https://github.com/prometheus/procfs/compare/v0.0.2...v0.1.3)
- github.com/sirupsen/logrus: [v1.2.0 → v1.6.0](https://github.com/sirupsen/logrus/compare/v1.2.0...v1.6.0)
- go.opencensus.io: v0.21.0 → v0.22.2
- golang.org/x/crypto: 60c769a → 75b2880
- golang.org/x/exp: 509febe → da58074
- golang.org/x/lint: d0100b6 → fdd1cda
- golang.org/x/net: c0dbc17 → ab34263
- golang.org/x/oauth2: 0f29369 → 858c2ad
- golang.org/x/sync: 1122301 → cd5d95a
- golang.org/x/sys: 0732a99 → ed371f2
- golang.org/x/text: v0.3.2 → v0.3.3
- golang.org/x/time: 9d24e82 → 555d28b
- golang.org/x/tools: 2c0ae70 → 7b8e75d
- google.golang.org/api: v0.4.0 → v0.15.0
- google.golang.org/appengine: v1.5.0 → v1.6.5
- google.golang.org/genproto: 5c49e3e → cb27e3a
- google.golang.org/grpc: v1.26.0 → v1.28.0
- gopkg.in/check.v1: 788fd78 → 41f04d3
- gopkg.in/yaml.v2: v2.2.4 → v2.2.8
- honnef.co/go/tools: ea95bdf → v0.0.1-2019.2.3
- k8s.io/api: v0.17.0 → v0.19.0
- k8s.io/apimachinery: v0.17.1-beta.0 → v0.19.0
- k8s.io/client-go: v0.17.0 → v0.19.0
- k8s.io/cloud-provider: v0.17.0 → v0.19.0
- k8s.io/component-base: v0.17.0 → v0.19.0
- k8s.io/csi-translation-lib: 97c07dc → v0.19.0
- k8s.io/gengo: 0689ccc → 3a45101
- k8s.io/kube-openapi: 30be4d1 → 6aeccd4
- k8s.io/utils: e782cd3 → d5654de
- sigs.k8s.io/yaml: v1.1.0 → v1.2.0

### Removed
_Nothing has changed._

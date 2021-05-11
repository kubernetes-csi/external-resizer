module github.com/kubernetes-csi/external-resizer

go 1.16

require (
	github.com/container-storage-interface/spec v1.5.0
	github.com/golang/protobuf v1.5.1 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.4 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/kubernetes-csi/csi-lib-utils v0.9.1
	github.com/prometheus/client_golang v1.9.0 // indirect
	github.com/prometheus/common v0.19.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	golang.org/x/net v0.0.0-20210316092652-d523dce5a7f4 // indirect
	golang.org/x/oauth2 v0.0.0-20210313182246-cd4f82c27b84 // indirect
	golang.org/x/sys v0.0.0-20210317225723-c4fcb01b228e // indirect
	golang.org/x/term v0.0.0-20210317153231-de623e64d2a6 // indirect
	golang.org/x/text v0.3.5 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210317182105-75c7a8546eb9 // indirect
	google.golang.org/grpc v1.37.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.0
	k8s.io/apiserver v0.0.0-00010101000000-000000000000
	k8s.io/client-go v0.21.1
	k8s.io/component-base v0.21.1
	k8s.io/csi-translation-lib v0.21.0
	k8s.io/klog/v2 v2.8.0
	k8s.io/utils v0.0.0-20210305010621-2afb4311ab10 // indirect
)

replace (
	// go get -d github.com/chrishenzie/csi-lib-utils@single-node-access-modes
	github.com/kubernetes-csi/csi-lib-utils => github.com/chrishenzie/csi-lib-utils v0.9.2-0.20210614221230-48c8713d1279
	// go get -d github.com/chrishenzie/kubernetes/staging/src/k8s.io/api@read-write-once-pod-access-mode
	k8s.io/api => github.com/chrishenzie/kubernetes/staging/src/k8s.io/api v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/apiextensions-apiserver => github.com/chrishenzie/kubernetes/staging/src/k8s.io/apiextensions-apiserver v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/apimachinery => github.com/chrishenzie/kubernetes/staging/src/k8s.io/apimachinery v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/apiserver => github.com/chrishenzie/kubernetes/staging/src/k8s.io/apiserver v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/cli-runtime => github.com/chrishenzie/kubernetes/staging/src/k8s.io/cli-runtime v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/client-go => github.com/chrishenzie/kubernetes/staging/src/k8s.io/client-go v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/cloud-provider => github.com/chrishenzie/kubernetes/staging/src/k8s.io/cloud-provider v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/cluster-bootstrap => github.com/chrishenzie/kubernetes/staging/src/k8s.io/cluster-bootstrap v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/code-generator => github.com/chrishenzie/kubernetes/staging/src/k8s.io/code-generator v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/component-base => github.com/chrishenzie/kubernetes/staging/src/k8s.io/component-base v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/component-helpers => github.com/chrishenzie/kubernetes/staging/src/k8s.io/component-helpers v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/controller-manager => github.com/chrishenzie/kubernetes/staging/src/k8s.io/controller-manager v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/cri-api => github.com/chrishenzie/kubernetes/staging/src/k8s.io/cri-api v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/csi-translation-lib => github.com/chrishenzie/kubernetes/staging/src/k8s.io/csi-translation-lib v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/kube-aggregator => github.com/chrishenzie/kubernetes/staging/src/k8s.io/kube-aggregator v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/kube-controller-manager => github.com/chrishenzie/kubernetes/staging/src/k8s.io/kube-controller-manager v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/kube-proxy => github.com/chrishenzie/kubernetes/staging/src/k8s.io/kube-proxy v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/kube-scheduler => github.com/chrishenzie/kubernetes/staging/src/k8s.io/kube-scheduler v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/kubectl => github.com/chrishenzie/kubernetes/staging/src/k8s.io/kubectl v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/kubelet => github.com/chrishenzie/kubernetes/staging/src/k8s.io/kubelet v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/legacy-cloud-providers => github.com/chrishenzie/kubernetes/staging/src/k8s.io/legacy-cloud-providers v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/metrics => github.com/chrishenzie/kubernetes/staging/src/k8s.io/metrics v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/mount-utils => github.com/chrishenzie/kubernetes/staging/src/k8s.io/mount-utils v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/node-api => github.com/chrishenzie/kubernetes/staging/src/k8s.io/node-api v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/sample-apiserver => github.com/chrishenzie/kubernetes/staging/src/k8s.io/sample-apiserver v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/sample-cli-plugin => github.com/chrishenzie/kubernetes/staging/src/k8s.io/sample-cli-plugin v0.0.0-20210507180302-a29b4b67ec78
	k8s.io/sample-controller => github.com/chrishenzie/kubernetes/staging/src/k8s.io/sample-controller v0.0.0-20210507180302-a29b4b67ec78
)

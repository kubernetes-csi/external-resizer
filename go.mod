module github.com/kubernetes-csi/external-resizer

go 1.16

require (
	github.com/container-storage-interface/spec v1.4.0
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
	google.golang.org/grpc v1.36.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.21.0
	k8s.io/component-base v0.21.0 // indirect
	k8s.io/csi-translation-lib v0.21.0
	k8s.io/klog/v2 v2.8.0
	k8s.io/kube-openapi v0.0.0-20210305164622-f622666832c1 // indirect
	k8s.io/utils v0.0.0-20210305010621-2afb4311ab10 // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.21.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.0
	k8s.io/client-go => k8s.io/client-go v0.21.0
	k8s.io/component-base => k8s.io/component-base v0.21.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.0
	k8s.io/node-api => k8s.io/node-api v0.21.0
)

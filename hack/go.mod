module github.com/opstrace/cortex-operator

go 1.15

require (
	github.com/alicebob/gopher-json v0.0.0-20200520072559-a9ecdc9d1d3a // indirect
	github.com/baiyubin/aliyun-sts-go-sdk v0.0.0-20180326062324-cfa1a18b161f // indirect
	github.com/cortexproject/cortex v1.8.1
	github.com/gomodule/redigo v1.8.4 // indirect
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/yuin/gopher-lua v0.0.0-20200816102855-ee81675732da // indirect
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200819165624-17cef6e3e9d5 // indirect
	k8s.io/api v0.21.0 // indirect
	k8s.io/apimachinery v0.21.0 // indirect
	k8s.io/client-go v0.21.0 // indirect
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800 // indirect
)

// Override to pin versions.

replace github.com/cortexproject/cortex => github.com/cortexproject/cortex v1.8.1

// replace github.com/cortexproject/cortex => /home/sreis/work/cortex

replace github.com/go-logr/logr => github.com/go-logr/logr v0.3.0

replace k8s.io/api => k8s.io/api v0.19.2

replace k8s.io/apimachinery => k8s.io/apimachinery v0.19.2

replace k8s.io/utils => k8s.io/utils v0.0.0-20200912215256-4140de9c8800

replace sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.7.2

// Overrides taken from https://github.com/cortexproject/cortex/blob/4afaa357469fb37fd92cb1e2a0c1d10cdddc5ecd/go.mod

// Override since git.apache.org is down.  The docs say to fetch from github.
replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

replace k8s.io/client-go => k8s.io/client-go v0.19.2

// replace k8s.io/api => k8s.io/api v0.19.4

// >v1.2.0 has some conflict with prometheus/alertmanager. Hence prevent the upgrade till it's fixed.
replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.0

// Use fork of gocql that has gokit logs and Prometheus metrics.
replace github.com/gocql/gocql => github.com/grafana/gocql v0.0.0-20200605141915-ba5dc39ece85

// Using a 3rd-party branch for custom dialer - see https://github.com/bradfitz/gomemcache/pull/86
replace github.com/bradfitz/gomemcache => github.com/themihai/gomemcache v0.0.0-20180902122335-24332e2d58ab

// Pin github.com/go-openapi versions to match Prometheus alertmanager to avoid
// breaking changing affecting the alertmanager.
replace github.com/go-openapi/errors => github.com/go-openapi/errors v0.19.4

replace github.com/go-openapi/loads => github.com/go-openapi/loads v0.19.5

replace github.com/go-openapi/runtime => github.com/go-openapi/runtime v0.19.15

replace github.com/go-openapi/spec => github.com/go-openapi/spec v0.19.8

replace github.com/go-openapi/strfmt => github.com/go-openapi/strfmt v0.19.5

replace github.com/go-openapi/swag => github.com/go-openapi/swag v0.19.9

replace github.com/go-openapi/validate => github.com/go-openapi/validate v0.19.8

replace github.com/weaveworks/common => github.com/weaveworks/common v0.0.0-20210112142934-23c8d7fa6120

replace github.com/sercand/kuberesolver => github.com/sercand/kuberesolver v2.4.0+incompatible

module github.com/vmware-tanzu/astrolabe

go 1.13

require (
	github.com/aws/aws-sdk-go v1.28.3
	github.com/go-openapi/errors v0.19.3
	github.com/go-openapi/loads v0.19.4
	github.com/go-openapi/runtime v0.19.11
	github.com/go-openapi/spec v0.19.6
	github.com/go-openapi/strfmt v0.19.4
	github.com/go-openapi/swag v0.19.7
	github.com/go-openapi/validate v0.19.6
	github.com/go-swagger/go-swagger v0.22.0 // indirect
	github.com/google/uuid v1.1.1
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jessevdk/go-flags v1.4.0
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.3.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/vmware/govmomi v0.22.1
	github.com/vmware/gvddk v0.8.1
	golang.org/x/net v0.0.0-20200114155413-6afb5195e5aa
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.0
	k8s.io/utils v0.0.0-20200109141947-94aeca20bf09 // indirect
)

replace github.com/vmware/gvddk => ./vendor/github.com/vmware/gvddk

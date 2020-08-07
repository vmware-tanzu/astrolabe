module github.com/vmware-tanzu/astrolabe

go 1.13

require (
	github.com/aws/aws-sdk-go v1.29.19
	github.com/go-openapi/errors v0.19.4
	github.com/go-openapi/loads v0.19.5
	github.com/go-openapi/runtime v0.19.12
	github.com/go-openapi/spec v0.19.7
	github.com/go-openapi/strfmt v0.19.5
	github.com/go-openapi/swag v0.19.8
	github.com/go-openapi/validate v0.19.7
	github.com/google/uuid v1.1.1
	github.com/jessevdk/go-flags v1.4.0
	github.com/labstack/echo v3.3.10+incompatible
	github.com/magiconair/properties v1.8.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/urfave/cli/v2 v2.2.0
	github.com/vmware-tanzu/velero-plugin-for-vsphere v1.0.1
	github.com/vmware/govmomi v0.22.2-0.20200329013745-f2eef8fc745f
	github.com/vmware/gvddk v0.8.1
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
)

replace github.com/vmware/gvddk => ./vendor/github.com/vmware/gvddk

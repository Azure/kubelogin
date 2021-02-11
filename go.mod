module github.com/Azure/kubelogin

go 1.13

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v0.14.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v0.8.0
	github.com/Azure/go-autorest/autorest v0.11.17
	github.com/Azure/go-autorest/autorest/adal v0.9.12
	github.com/golang/mock v1.2.0
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897
	k8s.io/apimachinery v0.17.4
	k8s.io/cli-runtime v0.17.4
	k8s.io/client-go v0.17.4
	k8s.io/klog v1.0.0
)

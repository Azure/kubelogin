package converter

import (
	"fmt"

	"github.com/Azure/kubelogin/pkg/internal/token"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type Options struct {
	Flags        *pflag.FlagSet
	configFlags  genericclioptions.RESTClientGetter
	TokenOptions token.Options
	// context is the kubeconfig context name
	context        string
	azureConfigDir string
}

func stringptr(str string) *string { return &str }

func New() Options {
	configFlags := &genericclioptions.ConfigFlags{
		KubeConfig: stringptr(""),
	}
	return Options{configFlags: configFlags}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.TokenOptions = token.NewOptions()
	if cf, ok := o.configFlags.(*genericclioptions.ConfigFlags); ok {
		cf.AddFlags(fs)
	}
	fs.StringVar(&o.context, flagContext, "", "The name of the kubeconfig context to use")
	fs.StringVar(&o.azureConfigDir, flagAzureConfigDir, "", "Azure CLI config path")
	o.TokenOptions.AddFlags(fs)
}

func (o *Options) Validate() error {
	return o.TokenOptions.Validate()
}

func (o *Options) UpdateFromEnv() {
	o.TokenOptions.UpdateFromEnv()
}

func (o *Options) ToString() string {
	return fmt.Sprintf("Context: %s, %s", o.context, o.TokenOptions.ToString())
}

func (o *Options) isSet(name string) bool {
	found := false
	o.Flags.Visit(func(f *pflag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

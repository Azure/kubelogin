package converter

import (
	"github.com/Azure/kubelogin/pkg/token"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type Options struct {
	Flags        *pflag.FlagSet
	configFlags  *genericclioptions.ConfigFlags
	TokenOptions token.Options
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
	o.configFlags.AddFlags(fs)
	o.TokenOptions.AddFlags(fs)
}

func (o *Options) Validate() error {
	return o.TokenOptions.Validate()
}

func (o *Options) UpdateFromEnv() {
	o.TokenOptions.UpdateFromEnv()
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

package converter

import (
	"fmt"

	"github.com/Azure/kubelogin/pkg/internal/token"
	"github.com/spf13/cobra"
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
	o.TokenOptions = token.NewOptions(true)
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

func (o *Options) AddCompletions(cmd *cobra.Command) {
	_ = cmd.RegisterFlagCompletionFunc(flagContext, completeContexts(o))
	_ = cmd.MarkFlagDirname(flagAzureConfigDir)
	_ = cmd.MarkFlagFilename("kubeconfig", "")

	o.TokenOptions.AddCompletions(cmd)

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		// Set a default completion function if none was set. We don't look
		// up if it does already have one set, because Cobra does this for
		// us, and returns an error (which we ignore for this reason).
		_ = cmd.RegisterFlagCompletionFunc(flag.Name, cobra.NoFileCompletions)
	})
}

func completeContexts(o *Options) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		clientConfig := o.configFlags.ToRawKubeConfigLoader()
		config, err := clientConfig.RawConfig()
		if err != nil {
			cobra.CompDebugln(fmt.Sprintf("unable to load kubeconfig: %s", err), false)
		}

		contexts := make([]string, 0, len(config.Contexts))
		for name := range config.Contexts {
			contexts = append(contexts, name)
		}

		return contexts, cobra.ShellCompDirectiveNoFileComp
	}
}

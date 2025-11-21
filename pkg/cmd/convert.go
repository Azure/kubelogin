package cmd

import (
	"github.com/Azure/kubelogin/pkg/internal/converter"
	"github.com/Azure/kubelogin/pkg/internal/options"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

// newConvertCmd provides a cobra command for convert sub command
func newConvertCmd() *cobra.Command {
	// Use unified options if feature flag is enabled
	if useUnifiedOptions() {
		return options.NewUnifiedCommand(options.ConvertCommand)
	}

	// Fallback to legacy implementation
	return newConvertCmdLegacy()
}

// Feature flag for gradual rollout
func useUnifiedOptions() bool {
	return options.UseUnifiedOptions()
}

// newConvertCmdLegacy provides the legacy converter command implementation
func newConvertCmdLegacy() *cobra.Command {
	o := converter.New()

	cmd := &cobra.Command{
		Use:          "convert-kubeconfig",
		Short:        "convert kubeconfig to use exec auth module",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			o.Flags = c.Flags()
			o.UpdateFromEnv()

			if err := o.Validate(); err != nil {
				return err
			}

			pathOptions := clientcmd.NewDefaultPathOptions()
			pathOptions.LoadingRules.ExplicitPath, _ = o.Flags.GetString("kubeconfig")

			if err := converter.Convert(o, pathOptions); err != nil {
				return err
			}
			return nil
		},
		ValidArgsFunction: cobra.NoFileCompletions,
	}

	o.AddFlags(cmd.Flags())
	o.AddCompletions(cmd)

	return cmd
}

package cmd

import (
	"github.com/Azure/kubelogin/pkg/token"
	"github.com/spf13/cobra"
)

// NewTokenCmd provides a cobra command for convert sub command
func NewTokenCmd() *cobra.Command {
	o := token.NewOptions()

	cmd := &cobra.Command{
		Use:          "get-token",
		Short:        "get AAD token",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			o.UpdateFromEnv()

			if err := o.Validate(); err != nil {
				return err
			}
			plugin, err := token.New(&o)
			if err != nil {
				return err
			}
			if err := plugin.Do(); err != nil {
				return err
			}
			return nil
		},
	}

	o.AddFlags(cmd.Flags())
	return cmd
}

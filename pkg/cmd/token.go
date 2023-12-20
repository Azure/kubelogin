package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/Azure/kubelogin/pkg/internal/token"
	"github.com/spf13/cobra"
)

// newTokenCmd provides a cobra command for convert sub command
func newTokenCmd() *cobra.Command {
	o := token.NewOptions()

	cmd := &cobra.Command{
		Use:          "get-token",
		Short:        "get AAD token",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			o.UpdateFromEnv()

			ctx := context.Background()
			ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
			defer cancel()

			if err := o.Validate(); err != nil {
				return err
			}

			plugin, err := token.New(&o)
			if err != nil {
				return err
			}
			if err := plugin.Do(ctx); err != nil {
				return err
			}
			return nil
		},
	}

	o.AddFlags(cmd.Flags())
	return cmd
}

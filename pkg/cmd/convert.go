package cmd

import (
	"github.com/Azure/kubelogin/pkg/converter"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

// NewConvertCmd provides a cobra command for convert sub command
func NewConvertCmd() *cobra.Command {
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
			if err := converter.Convert(o, pathOptions); err != nil {
				return err
			}
			return nil
		},
	}

	o.AddFlags(cmd.Flags())

	return cmd
}

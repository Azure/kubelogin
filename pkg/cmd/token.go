package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Azure/kubelogin/pkg/token"
	"github.com/spf13/cobra"
	"k8s.io/client-go/pkg/apis/clientauthentication"
)

const execInfoEnv = "KUBERNETES_EXEC_INFO"

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
			env := os.Getenv(execInfoEnv)
			fmt.Fprintln(os.Stderr, os.Getenv(execInfoEnv))
			var execCredential clientauthentication.ExecCredential
			error := json.Unmarshal([]byte(env), &execCredential)
			if error != nil {
				return fmt.Errorf("cannot convert to ExecCredential: %w", error)
			}
			if !execCredential.Spec.Interactive && o.LoginMethod == "DeviceCodeLogin" {
				return fmt.Errorf("devicelogin is not supported if interactiveMode is 'never'")
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

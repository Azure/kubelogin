package cmd

import (
	"os"

	"github.com/Azure/kubelogin/pkg/token"
	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"
)

// NewRemoveTokenCacheCmd provides a cobra command for removing token cache sub command
func NewRemoveTokenCacheCmd() *cobra.Command {
	var tokenCacheDir string

	cmd := &cobra.Command{
		Use:          "remove-tokens",
		Short:        "Remove all cached tokens from filesystem",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := os.RemoveAll(tokenCacheDir); err != nil {
				klog.V(5).Infof("unable to delete tokens cache in '%s': %s", tokenCacheDir, err)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&tokenCacheDir, "token-cache-dir", token.DefaultTokenCacheDir, "directory to cache token")
	return cmd
}

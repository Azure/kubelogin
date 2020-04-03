package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/klog"
)

// NewRemoveTokenCacheCmd provides a cobra command for removing token cache sub command
func NewRemoveTokenCacheCmd() *cobra.Command {
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

	addTokenCacheDirFlags(cmd.Flags())
	return cmd
}

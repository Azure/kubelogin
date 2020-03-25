package cmd

import (
	"os"
	"path"

	"github.com/spf13/cobra"
	"k8s.io/klog"
)

// NewRemoveTokenCacheCmd provides a cobra command for removing token cache sub command
func NewRemoveTokenCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "remove-token",
		Short:        "Remove cached token from filesystem",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			cache := path.Join(tokenCacheDir, cacheFile)
			if err := os.Remove(cache); err != nil {
				klog.V(5).Infof("unable to delete token cache '%s': %s", cache, err)
			}
			return nil
		},
	}

	addTokenCacheDirFlags(cmd.Flags())
	return cmd
}

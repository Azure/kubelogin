package cmd

import (
	"os"

	"github.com/Azure/kubelogin/pkg/internal/token"
	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"
)

// newRemoveAuthRecordCacheCmd provides a cobra command for removing token cache sub command
func newRemoveAuthRecordCacheCmdDeprecated() *cobra.Command {
	var authRecordCacheDir string

	cmd := &cobra.Command{
		Use:          "remove-tokens",
		Short:        "Remove all cached authentication record from filesystem",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := os.RemoveAll(authRecordCacheDir); err != nil {
				klog.V(5).Infof("unable to delete authentication record cache in '%s': %s", authRecordCacheDir, err)
			}
			return nil
		},
		ValidArgsFunction: cobra.NoFileCompletions,
		Deprecated:        "remove-tokens is deprecated, use remove-cache-dir instead",
	}

	cmd.Flags().StringVar(&authRecordCacheDir, "token-cache-dir", token.DefaultAuthRecordCacheDir, "directory to cache authentication record")
	return cmd
}

package cmd

import (
	"fmt"
	"os"

	"github.com/Azure/kubelogin/pkg/internal/token"
	"github.com/spf13/cobra"
)

// newRemoveAuthRecordCacheCmd provides a cobra command for removing token cache sub command
func newRemoveAuthRecordCacheCmd() *cobra.Command {
	var authRecordCacheDir string

	cmd := &cobra.Command{
		Use:          "remove-cache-dir",
		Short:        "Remove all cached authentication record from filesystem",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := os.RemoveAll(authRecordCacheDir); err != nil {
				return fmt.Errorf("unable to delete authentication record cache in %q: %w", authRecordCacheDir, err)
			}
			return nil
		},
		ValidArgsFunction: cobra.NoFileCompletions,
	}

	cmd.Flags().StringVar(&authRecordCacheDir, "cache-dir", token.DefaultAuthRecordCacheDir, "directory to cache authentication record")
	return cmd
}

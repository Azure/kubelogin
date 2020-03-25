package cmd

import (
	"path"

	"github.com/Azure/kubelogin/pkg/token"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/util/homedir"
)

const cacheFile = "azure.json"

var tokenCacheDir = homedir.HomeDir() + "/.kube/cache/kubelogin"

// NewTokenCmd provides a cobra command for convert sub command
func NewTokenCmd() *cobra.Command {
	o := token.NewOptions()

	cmd := &cobra.Command{
		Use:          "get-token",
		Short:        "get AAD token",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			o.TokenCacheFile = path.Join(tokenCacheDir, cacheFile)

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

	addTokenCacheDirFlags(cmd.Flags())
	o.AddFlags(cmd.Flags())
	return cmd
}

func addTokenCacheDirFlags(fs *pflag.FlagSet) {
	fs.StringVar(&tokenCacheDir, "token-cache-dir", tokenCacheDir, "directory to cache token")
}

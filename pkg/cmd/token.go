package cmd

import (
	"fmt"
	"path"

	"github.com/Azure/kubelogin/pkg/token"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/util/homedir"
)

var tokenCacheDir = homedir.HomeDir() + "/.kube/cache/kubelogin"

// NewTokenCmd provides a cobra command for convert sub command
func NewTokenCmd() *cobra.Command {
	o := token.NewOptions()

	cmd := &cobra.Command{
		Use:          "get-token",
		Short:        "get AAD token",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			o.UpdateFromEnv()

			cacheFile := getCacheFileName(o.Environment, o.ServerID, o.ClientID, o.TenantID, o.IsLegacy)
			o.TokenCacheFile = path.Join(tokenCacheDir, cacheFile)

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

func getCacheFileName(environment, serverID, clientID, tenantID string, legacy bool) string {
	// format: ${environment}-${server-id}-${client-id}-${tenant-id}[_legacy].json
	cacheFileNameFormat := "%s-%s-%s-%s.json"
	if legacy {
		cacheFileNameFormat = "%s-%s-%s-%s_legacy.json"
	}
	return fmt.Sprintf(cacheFileNameFormat, environment, serverID, clientID, tenantID)
}

package token

import (
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
)

func TestOptions(t *testing.T) {
	t.Run("Default option should produce token cache file under default token cache directory", func(t *testing.T) {
		o := NewOptions()
		o.AddFlags(&pflag.FlagSet{})
		o.UpdateFromEnv()
		if err := o.Validate(); err != nil {
			t.Fatalf("option validation failed: %s", err)
		}
		dir, _ := filepath.Split(o.tokenCacheFile)
		if dir != DefaultTokenCacheDir {
			t.Fatalf("token cache directory is expected to be %s, got %s", DefaultTokenCacheDir, dir)
		}
	})

	t.Run("option with customized token cache dir should produce token cache file under specified token cache directory", func(t *testing.T) {
		o := NewOptions()
		o.TokenCacheDir = "/tmp/foo/"
		o.AddFlags(&pflag.FlagSet{})
		o.UpdateFromEnv()
		if err := o.Validate(); err != nil {
			t.Fatalf("option validation failed: %s", err)
		}
		dir, _ := filepath.Split(o.tokenCacheFile)
		if dir != o.TokenCacheDir {
			t.Fatalf("token cache directory is expected to be %s, got %s", o.TokenCacheDir, dir)
		}
	})
}

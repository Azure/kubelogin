//go:build windows

package cache

import (
	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache/accessor"
)

// storage creates a platform-specific accessor for Windows
func storage(cachePath string) (accessor.Accessor, error) {
	return accessor.New(cachePath)
}

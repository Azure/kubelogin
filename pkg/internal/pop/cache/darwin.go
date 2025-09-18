//go:build darwin && cgo

package cache

import (
	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache/accessor"
)

// storage creates a platform-specific accessor for macOS
func storage(cachePath string) (accessor.Accessor, error) {
	// Use "kubelogin-pop" as the service name in macOS Keychain
	// "MSALCache" becomes the account identifier within that service
	return accessor.New("kubelogin-pop", accessor.WithAccount("MSALCache"))
}

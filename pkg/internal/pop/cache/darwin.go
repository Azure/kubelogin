//go:build darwin && cgo

package cache

import (
	"path/filepath"

	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache/accessor"
)

// storage creates a platform-specific accessor for macOS for MSAL cache
func storage(cachePath string) (accessor.Accessor, error) {
	// Use the filename from cachePath as the account identifier
	accountName := filepath.Base(cachePath)
	// Use "kubelogin-pop" as the service name in macOS Keychain
	return accessor.New("kubelogin-pop", accessor.WithAccount(accountName))
}

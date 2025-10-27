//go:build darwin && cgo

package cache

import (
	"path/filepath"

	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache/accessor"
)

// storage creates a platform-specific accessor for macOS
func storage(cachePath string) (accessor.Accessor, error) {
	// Use the cache filename as the account identifier to differentiate between
	// different caches (e.g., test cache vs actual cache) on macOS Keychain
	account := filepath.Base(cachePath)
	return accessor.New("kubelogin-pop", accessor.WithAccount(account))
}

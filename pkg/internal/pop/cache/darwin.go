//go:build darwin && cgo

package cache

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache/accessor"
)

// storage creates a platform-specific accessor for macOS
func storage(cachePath string) (accessor.Accessor, error) {
	// Use a hash of the full cache path as the account identifier to ensure uniqueness
	// This prevents cache conflicts when multiple cache directories are used
	// (e.g., different clusters with different --cache-dir settings)
	// We use a hash because macOS Keychain has limitations on account name length
	hash := sha256.Sum256([]byte(cachePath))
	account := hex.EncodeToString(hash[:16]) // Use first 16 bytes for reasonable uniqueness
	return accessor.New("kubelogin-pop", accessor.WithAccount(account))
}

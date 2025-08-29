package token

import (
	extcache "github.com/AzureAD/microsoft-authentication-extensions-for-go/cache"
	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache/accessor"
)

// MSALCacheProvider is a type alias for the official Microsoft authentication extensions cache.
// It provides secure, persistent token storage using system-appropriate storage mechanisms
// like libsecret (GNOME Keyring), KDE Wallet, or Windows Credential Manager.
type MSALCacheProvider = *extcache.Cache

// NewMSALCacheProvider creates a new MSAL cache provider using the official
// microsoft-authentication-extensions-for-go/cache implementation.
// It uses secure system storage (libsecret, KDE Wallet, etc.) for token persistence.
func NewMSALCacheProvider(file string) (MSALCacheProvider, error) {
	// Create a storage accessor for the cache file
	storage, err := accessor.New("kubelogin-cache")
	if err != nil {
		return nil, err
	}

	// Create and return the cache
	return extcache.New(storage, file)
}

package cache

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache/accessor"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
)

const popTokenCacheFileName = "pop_tokens.cache"

// getPoPCacheFilePath returns the file path for the PoP token cache.
// This is separate from the authentication record cache file.
func getPoPCacheFilePath(cacheDir string) string {
	return filepath.Join(cacheDir, popTokenCacheFileName)
}

// Cache implements the MSAL cache.ExportReplace interface using our platform-specific PoP cache.
// This provides secure, persistent PoP token storage without depending on libsecret on Linux.
// Cache provides a unified interface for PoP token caching following azidentity patterns.
type Cache struct {
	accessor accessor.Accessor
}

// NewCache creates a new MSAL cache provider using custom platform-specific PoP cache.
// This implementation provides secure storage on all platforms without external dependencies like libsecret on Linux.
// This is following the azidentity pattern.
func NewCache(cacheDir string) (*Cache, error) {
	cachePath := getPoPCacheFilePath(cacheDir)
	acc, err := storage(cachePath)
	if err != nil {
		return nil, err
	}

	return &Cache{
		accessor: acc,
	}, nil
}

// Export saves the current PoP token cache state to platform-specific secure storage.
// This method is called by MSAL to persist PoP tokens across application restarts.
func (c *Cache) Export(ctx context.Context, marshaler cache.Marshaler, hints cache.ExportHints) error {
	// Get the cache data from the marshaler
	data, err := marshaler.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal PoP cache data: %w", err)
	}

	return c.accessor.Write(ctx, data)
}

// Replace loads PoP token cache data from platform-specific secure storage and restores it into MSAL's in-memory cache.
// This method is called by MSAL during initialization to restore previously cached PoP tokens from persistent storage.
func (c *Cache) Replace(ctx context.Context, unmarshaler cache.Unmarshaler, hints cache.ReplaceHints) error {
	data, err := c.accessor.Read(ctx)
	if err != nil {
		// If cache doesn't exist, initialize with empty cache
		return unmarshaler.Unmarshal([]byte("{}"))
	}

	return unmarshaler.Unmarshal(data)
}

// Clear removes all PoP token data from the cache.
func (c *Cache) Clear(ctx context.Context) error {
	return c.accessor.Delete(ctx)
}

// NewSecureAccessor creates a new platform-specific secure storage accessor.
// This can be used for storing other sensitive data like RSA private keys
// using the same encrypted storage infrastructure as the PoP token cache.
func NewSecureAccessor(cachePath string) (accessor.Accessor, error) {
	return storage(cachePath)
}

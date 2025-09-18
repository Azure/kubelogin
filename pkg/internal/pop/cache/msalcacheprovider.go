package cache

import (
	"context"
	"fmt"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
)

// MSALCacheProvider implements the MSAL cache.ExportReplace interface using our platform-specific PoP cache.
// This provides secure, persistent PoP token storage without depending on libsecret on Linux.
type MSALCacheProvider struct {
	cache *Cache
}

// NewMSALCacheProvider creates a new MSAL cache provider using our custom platform-specific PoP cache.
// This implementation provides secure storage on all platforms without external dependencies like libsecret on Linux.
func NewMSALCacheProvider(name string) (*MSALCacheProvider, error) {
	popCache, err := New(name)
	if err != nil {
		return nil, fmt.Errorf("failed to create PoP cache: %w", err)
	}

	return &MSALCacheProvider{
		cache: popCache,
	}, nil
}

// Export writes the PoP token cache to external storage.
// This method is called by MSAL to persist the current cache state.
func (p *MSALCacheProvider) Export(ctx context.Context, marshaler cache.Marshaler, hints cache.ExportHints) error {
	// Get the cache data from the marshaler
	data, err := marshaler.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal PoP cache data: %w", err)
	}

	return p.cache.Write(ctx, data)
}

// Replace replaces the entire PoP token cache with data from external storage.
// This method is called by MSAL to restore the cache state.
func (p *MSALCacheProvider) Replace(ctx context.Context, unmarshaler cache.Unmarshaler, hints cache.ReplaceHints) error {
	data, err := p.cache.Read(ctx)
	if err != nil {
		// If cache doesn't exist, initialize with empty cache
		return unmarshaler.Unmarshal([]byte("{}"))
	}

	return unmarshaler.Unmarshal(data)
}

// Clear removes all PoP token data from the cache.
func (p *MSALCacheProvider) Clear(ctx context.Context) error {
	return p.cache.Delete(ctx)
}

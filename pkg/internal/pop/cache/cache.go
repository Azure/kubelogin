package cache

import (
	"context"

	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache/accessor"
)

// New creates a new cache instance following the azidentity pattern.
// It uses platform-specific storage implementations.
func New(name string) (*Cache, error) {
	dir, err := cacheDir()
	if err != nil {
		return nil, err
	}

	acc, err := storage(name)
	if err != nil {
		return nil, err
	}

	return &Cache{
		accessor: acc,
		dir:      dir,
		name:     name,
	}, nil
}

// Cache provides a unified interface for PoP token caching following azidentity patterns.
type Cache struct {
	accessor accessor.Accessor
	dir      string
	name     string
}

// Read retrieves cached PoP token data.
func (c *Cache) Read(ctx context.Context) ([]byte, error) {
	return c.accessor.Read(ctx)
}

// Write stores PoP token data in the cache.
func (c *Cache) Write(ctx context.Context, data []byte) error {
	return c.accessor.Write(ctx, data)
}

// Delete removes cached PoP token data.
func (c *Cache) Delete(ctx context.Context) error {
	return c.accessor.Delete(ctx)
}

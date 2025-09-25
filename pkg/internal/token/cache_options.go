package token

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"
	popCache "github.com/Azure/kubelogin/pkg/internal/pop/cache"
	msalCache "github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
)

// CacheOptions consolidates both azidentity and PoP token cache configurations
// into a single structure to simplify cache management across different credential types.
type CacheOptions struct {
	// AzIdentityCache is used by standard azidentity credentials (DeviceCode, Interactive, etc.)
	// when UsePersistentCache is enabled
	AzIdentityCache azidentity.Cache
	
	// PoPCache is used by PoP-enabled credentials (InteractiveWithPoP, ClientSecretWithPoP, etc.)
	// for MSAL token caching
	PoPCache msalCache.ExportReplace
	
	// CacheDir is the directory where cache files are stored
	CacheDir string
}

// NewCacheOptions creates a new CacheOptions instance with the specified cache directory.
// Both cache types are initialized if persistent caching is enabled.
func NewCacheOptions(cacheDir string, usePersistentCache bool) (*CacheOptions, error) {
	opts := &CacheOptions{
		CacheDir: cacheDir,
	}

	if usePersistentCache {
		// Initialize azidentity cache for standard credentials
		azCache, err := cache.New(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create azidentity cache: %w", err)
		}
		opts.AzIdentityCache = azCache

		// Initialize PoP cache for PoP-enabled credentials
		popCacheInstance, err := popCache.NewCache(cacheDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create PoP cache: %w", err)
		}
		opts.PoPCache = popCacheInstance
	}

	return opts, nil
}

// HasAzIdentityCache returns true if azidentity cache is available
func (c *CacheOptions) HasAzIdentityCache() bool {
	return c.AzIdentityCache != nil
}

// HasPoPCache returns true if PoP cache is available
func (c *CacheOptions) HasPoPCache() bool {
	return c.PoPCache != nil
}

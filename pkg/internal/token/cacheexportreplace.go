package token

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
)

// MSALCacheProvider exports and replaces in-memory cache data. It doesn't support nil Context or
// define the outcome of passing one. A Context without a timeout must receive a default timeout
// specified by the implementor. Retries must be implemented inside the implementation.
type MSALCacheProvider interface {
	// Replace replaces the cache with what is in external storage. Implementors should honor
	// Context cancellations and return context.Canceled or context.DeadlineExceeded in those cases.
	Replace(ctx context.Context, cache cache.Unmarshaler, hints cache.ReplaceHints) error
	// Export writes the binary representation of the cache (cache.Marshal()) to external storage.
	// This is considered opaque. Context cancellations should be honored as in Replace.
	Export(ctx context.Context, cache cache.Marshaler, hints cache.ExportHints) error
}

type defaultMSALCacheProvider struct {
	file string
}

func (c *defaultMSALCacheProvider) Replace(ctx context.Context, cache cache.Unmarshaler, hints cache.ReplaceHints) error {
	// Check for context cancellation or timeout
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Read the cache data from the file
	data, err := os.ReadFile(c.file)
	if err != nil {
		if os.IsNotExist(err) {
			// If the file doesn't exist, return nil (no cache to replace)
			return nil
		}
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	// Unmarshal the data into the provided cache
	if err := cache.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return nil

}

func (c *defaultMSALCacheProvider) Export(ctx context.Context, cache cache.Marshaler, hints cache.ExportHints) error {
	// Check for context cancellation or timeout
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Marshal the cache data into bytes
	data, err := cache.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	// Ensure the directory for the file exists
	dir := filepath.Dir(c.file)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Write the marshaled data to the file
	if err := os.WriteFile(c.file, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

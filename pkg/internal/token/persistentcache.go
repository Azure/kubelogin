package token

import (
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"
)

// cacheNewFunc is the function used to create a new persistent cache.
// It is a variable to allow overriding in tests.
var cacheNewFunc = func(opts *cache.Options) (azidentity.Cache, error) {
	return cache.New(opts)
}

// newPersistentCache creates a persistent token cache with cross-process
// synchronization to prevent a race condition when multiple kubelogin
// processes start concurrently. The upstream azidentity/cache package
// tests storage availability using a non-atomic check-then-add pattern
// on macOS keychain, which fails with "The specified item already exists
// in the keychain" (-25299) when two processes race.
//
// This function uses a file lock to serialize the storage test across
// processes. If the lock cannot be acquired, it proceeds without locking
// (best-effort) to avoid breaking existing behavior.
//
// See https://github.com/Azure/kubelogin/issues/740
func newPersistentCache() (azidentity.Cache, error) {
	lockPath := filepath.Join(os.TempDir(), "kubelogin-cache-test.lock")
	unlock := acquireProcessLock(lockPath)
	defer unlock()
	return cacheNewFunc(nil)
}

package token

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"
)

// cacheNewFunc is the function used to create a new persistent cache.
// It is a variable to allow overriding in tests.
var cacheNewFunc = cache.New

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
	lockDir := lockFileDir()
	lockPath := filepath.Join(lockDir, "cache-test.lock")
	unlock := acquireProcessLock(lockPath)
	defer unlock()
	return cacheNewFunc(nil)
}

// lockFileDir returns a user-scoped directory for the lock file.
// It prefers os.UserCacheDir()/kubelogin and falls back to
// os.TempDir()/kubelogin-<uid> so that the lock file is never
// placed directly in a shared, world-writable directory.
func lockFileDir() string {
	if cacheDir, err := os.UserCacheDir(); err == nil {
		dir := filepath.Join(cacheDir, "kubelogin")
		if err := os.MkdirAll(dir, 0700); err == nil {
			return dir
		}
	}
	// Fallback: use a UID-scoped subdirectory under the temp dir
	// so the lock file is not directly in a world-writable location.
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("kubelogin-%d", os.Getuid()))
	if err := os.MkdirAll(dir, 0700); err != nil {
		// Last resort: use temp dir directly. The lock is best-effort,
		// so this is acceptable.
		return os.TempDir()
	}
	return dir
}

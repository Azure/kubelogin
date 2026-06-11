package token

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPersistentCache_Success(t *testing.T) {
	original := cacheNewFunc
	defer func() { cacheNewFunc = original }()

	called := false
	cacheNewFunc = func(opts *cache.Options) (azidentity.Cache, error) {
		called = true
		assert.Nil(t, opts)
		return azidentity.Cache{}, nil
	}

	c, err := newPersistentCache()
	assert.NoError(t, err)
	assert.Equal(t, azidentity.Cache{}, c)
	assert.True(t, called)
}

func TestNewPersistentCache_Error(t *testing.T) {
	original := cacheNewFunc
	defer func() { cacheNewFunc = original }()

	expectedErr := fmt.Errorf("test error")
	cacheNewFunc = func(opts *cache.Options) (azidentity.Cache, error) {
		return azidentity.Cache{}, expectedErr
	}

	c, err := newPersistentCache()
	assert.ErrorIs(t, err, expectedErr)
	assert.Equal(t, azidentity.Cache{}, c)
}

func TestNewPersistentCache_ConcurrentAccess(t *testing.T) {
	original := cacheNewFunc
	defer func() { cacheNewFunc = original }()

	var mu sync.Mutex
	callCount := 0
	cacheNewFunc = func(opts *cache.Options) (azidentity.Cache, error) {
		mu.Lock()
		callCount++
		mu.Unlock()
		return azidentity.Cache{}, nil
	}

	var wg sync.WaitGroup
	const goroutines = 10
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := newPersistentCache()
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
	assert.Equal(t, goroutines, callCount)
}

func TestAcquireProcessLock_ReturnsUnlockFunc(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "test.lock")
	unlock := acquireProcessLock(lockPath)
	require.NotNil(t, unlock)

	// Release the lock — should not panic
	unlock()
}

func TestAcquireProcessLock_InvalidPath(t *testing.T) {
	// Use a path in a non-existent directory
	lockPath := filepath.Join(t.TempDir(), "nonexistent", "subdir", "test.lock")
	unlock := acquireProcessLock(lockPath)
	require.NotNil(t, unlock)

	// Should be a no-op function, calling it should not panic
	unlock()
}

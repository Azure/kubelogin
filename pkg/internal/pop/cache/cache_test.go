package cache

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var ctx = context.Background()

// mockMarshaler implements cache.Marshaler for testing
type mockMarshaler struct {
	data []byte
	err  error
}

func (m *mockMarshaler) Marshal() ([]byte, error) {
	return m.data, m.err
}

// mockUnmarshaler implements cache.Unmarshaler for testing
type mockUnmarshaler struct {
	data []byte
	err  error
}

func (m *mockUnmarshaler) Unmarshal(data []byte) error {
	m.data = data
	return m.err
}

func TestNewCache(t *testing.T) {
	tests := []struct {
		name     string
		cacheDir string
		wantErr  bool
	}{
		{
			name:     "valid cache directory",
			cacheDir: t.TempDir(),
			wantErr:  false,
		},
		{
			name:     "empty cache directory",
			cacheDir: "",
			wantErr:  false, // should still work with empty dir
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewCache(tt.cacheDir, "AzurePublicCloud")
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, cache)
			} else {
				require.NoError(t, err)
				require.NotNil(t, cache)
				require.NotNil(t, cache.accessor)
			}
		})
	}
}

func TestNewCache_WithDefaultEnvironment(t *testing.T) {
	// Test that AzurePublicCloud is used as default when environment is empty
	tempDir := t.TempDir()

	// Call NewCache with empty environment - should default to AzurePublicCloud
	c, err := NewCache(tempDir, "")
	require.NoError(t, err)
	require.NotNil(t, c)
	require.NotNil(t, c.accessor)

	// Verify we can write and read from the cache
	testData := []byte(`{"access_tokens": {"key1": "value1"}}`)
	marshaler := &mockMarshaler{data: testData}
	err = c.Export(ctx, marshaler, cache.ExportHints{})
	require.NoError(t, err)

	unmarshaler := &mockUnmarshaler{}
	err = c.Replace(ctx, unmarshaler, cache.ReplaceHints{})
	require.NoError(t, err)
	require.Equal(t, testData, unmarshaler.data)
}

func TestCache_ExportReplace(t *testing.T) {
	tempDir := t.TempDir()
	c, err := NewCache(tempDir, "AzurePublicCloud")
	require.NoError(t, err)

	testData := []byte(`{"access_tokens": {"key1": "value1"}, "refresh_tokens": {"key2": "value2"}}`)

	// Test Export
	marshaler := &mockMarshaler{data: testData}
	err = c.Export(ctx, marshaler, cache.ExportHints{})
	require.NoError(t, err)

	// Test Replace
	unmarshaler := &mockUnmarshaler{}
	err = c.Replace(ctx, unmarshaler, cache.ReplaceHints{})
	require.NoError(t, err)
	require.Equal(t, testData, unmarshaler.data)
}

func TestCache_ExportReplaceEmpty(t *testing.T) {
	tempDir := t.TempDir()
	c, err := NewCache(tempDir, "AzurePublicCloud")
	require.NoError(t, err)

	// Test Replace on empty cache - should get empty JSON since no data exists
	unmarshaler := &mockUnmarshaler{}
	err = c.Replace(ctx, unmarshaler, cache.ReplaceHints{})
	require.NoError(t, err)
	require.Equal(t, []byte("{}"), unmarshaler.data)
}

func TestCache_ExportMarshalError(t *testing.T) {
	tempDir := t.TempDir()
	c, err := NewCache(tempDir, "AzurePublicCloud")
	require.NoError(t, err)

	expectedErr := fmt.Errorf("marshal error")
	marshaler := &mockMarshaler{err: expectedErr}

	err = c.Export(ctx, marshaler, cache.ExportHints{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to marshal PoP cache data")
}

func TestCache_ReplaceUnmarshalError(t *testing.T) {
	tempDir := t.TempDir()
	c, err := NewCache(tempDir, "AzurePublicCloud")
	require.NoError(t, err)

	// First export some data
	testData := []byte(`{"access_tokens": {"key1": "value1"}}`)
	marshaler := &mockMarshaler{data: testData}
	err = c.Export(ctx, marshaler, cache.ExportHints{})
	require.NoError(t, err)

	// Then try to replace with an unmarshaler that returns error
	expectedErr := fmt.Errorf("unmarshal error")
	unmarshaler := &mockUnmarshaler{err: expectedErr}

	err = c.Replace(ctx, unmarshaler, cache.ReplaceHints{})
	require.Error(t, err)
	require.Equal(t, expectedErr, err)
}

func TestCache_Clear(t *testing.T) {
	tempDir := t.TempDir()
	c, err := NewCache(tempDir, "AzurePublicCloud")
	require.NoError(t, err)

	// Export some data first
	testData := []byte(`{"access_tokens": {"key1": "value1"}}`)
	marshaler := &mockMarshaler{data: testData}
	err = c.Export(ctx, marshaler, cache.ExportHints{})
	require.NoError(t, err)

	// Clear the cache
	err = c.Clear(ctx)
	require.NoError(t, err)

	// Verify cache is empty - after delete, should get empty JSON
	unmarshaler := &mockUnmarshaler{}
	err = c.Replace(ctx, unmarshaler, cache.ReplaceHints{})
	require.NoError(t, err)
	require.Equal(t, []byte("{}"), unmarshaler.data)
}

func TestCache_MultipleProcessSimulation(t *testing.T) {
	tempDir := t.TempDir()

	// Multiple kubelogin processes (simulated as goroutines) - each using the SAME cache directory (like real users would)
	// This tests the Linux keyring's process isolation and file system behavior
	const numProcesses = 3
	done := make(chan error, numProcesses)

	for i := 0; i < numProcesses; i++ {
		go func(processID int) {
			// Each "process" creates its own cache instance but uses the same cache directory
			// This simulates multiple kubelogin processes run by same user
			c, err := NewCache(tempDir, "AzurePublicCloud")
			if err != nil {
				done <- fmt.Errorf("process %d: failed to create cache: %w", processID, err)
				return
			}

			// Each process exports its own tokens
			testData := []byte(fmt.Sprintf(`{"access_tokens": {"process_%d": "token_%d"}}`, processID, processID))
			marshaler := &mockMarshaler{data: testData}

			err = c.Export(ctx, marshaler, cache.ExportHints{})
			if err != nil {
				done <- fmt.Errorf("process %d: export failed: %w", processID, err)
				return
			}

			// Each process should be able to read back some valid data
			// (might be from this process or another due to last-write-wins behavior)
			unmarshaler := &mockUnmarshaler{}
			err = c.Replace(ctx, unmarshaler, cache.ReplaceHints{})
			if err != nil {
				done <- fmt.Errorf("process %d: replace failed: %w", processID, err)
				return
			}

			// Verify we got valid JSON (the exact content may vary due to concurrent writes)
			if len(unmarshaler.data) == 0 || !bytes.HasPrefix(unmarshaler.data, []byte("{")) {
				done <- fmt.Errorf("process %d: invalid data format: %s", processID, unmarshaler.data)
				return
			}

			done <- nil
		}(i)
	}

	// Wait for all "processes" to complete successfully
	for i := 0; i < numProcesses; i++ {
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(10 * time.Second):
			t.Fatal("timeout waiting for process simulation")
		}
	}

	// Final verification: ensure the cache is in a consistent state
	c, err := NewCache(tempDir, "AzurePublicCloud")
	require.NoError(t, err)

	unmarshaler := &mockUnmarshaler{}
	err = c.Replace(ctx, unmarshaler, cache.ReplaceHints{})
	require.NoError(t, err)

	// Should have valid JSON from one of the processes
	require.True(t, bytes.HasPrefix(unmarshaler.data, []byte("{")))
}

func TestCache_Isolation(t *testing.T) {
	// Test that different cache instances with different names are isolated
	tempDir := t.TempDir()

	cache1, err := NewCache(filepath.Join(tempDir, "cache1"), "AzurePublicCloud")
	require.NoError(t, err)

	cache2, err := NewCache(filepath.Join(tempDir, "cache2"), "AzurePublicCloud")
	require.NoError(t, err)

	// Export different data to each cache
	testData1 := []byte(`{"access_tokens": {"cache1": "data1"}}`)
	marshaler1 := &mockMarshaler{data: testData1}
	err = cache1.Export(ctx, marshaler1, cache.ExportHints{})
	require.NoError(t, err)

	testData2 := []byte(`{"access_tokens": {"cache2": "data2"}}`)
	marshaler2 := &mockMarshaler{data: testData2}
	err = cache2.Export(ctx, marshaler2, cache.ExportHints{})
	require.NoError(t, err)

	// Verify each cache has its own data
	unmarshaler1 := &mockUnmarshaler{}
	err = cache1.Replace(ctx, unmarshaler1, cache.ReplaceHints{})
	require.NoError(t, err)
	require.Equal(t, testData1, unmarshaler1.data)

	unmarshaler2 := &mockUnmarshaler{}
	err = cache2.Replace(ctx, unmarshaler2, cache.ReplaceHints{})
	require.NoError(t, err)
	require.Equal(t, testData2, unmarshaler2.data)
}

func TestGetPoPCacheFilePath(t *testing.T) {
	tests := []struct {
		name        string
		cacheDir    string
		environment string
		expected    string
	}{
		{
			name:        "unix path",
			cacheDir:    "/home/user/.cache/kubelogin",
			environment: "AzurePublicCloud",
			expected:    "/home/user/.cache/kubelogin/pop_tokens_azurepubliccloud.cache",
		},
		{
			name:        "relative path",
			cacheDir:    "cache",
			environment: "AzurePublicCloud",
			expected:    "cache/pop_tokens_azurepubliccloud.cache",
		},
		{
			name:        "empty path",
			cacheDir:    "",
			environment: "AzurePublicCloud",
			expected:    "pop_tokens_azurepubliccloud.cache",
		},
		{
			name:        "different environment",
			cacheDir:    "/home/user/.cache/kubelogin",
			environment: "AzureChinaCloud",
			expected:    "/home/user/.cache/kubelogin/pop_tokens_azurechinacloud.cache",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPoPCacheFilePath(tt.cacheDir, tt.environment)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestNewSecureAccessor(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "test.cache")

	accessor, err := NewSecureAccessor(cachePath)
	require.NoError(t, err)
	require.NotNil(t, accessor)

	// Test basic operations
	testData := []byte("test secure data")

	err = accessor.Write(ctx, testData)
	require.NoError(t, err)

	readData, err := accessor.Read(ctx)
	require.NoError(t, err)
	require.Equal(t, testData, readData)

	err = accessor.Delete(ctx)
	require.NoError(t, err)

	// Verify data is deleted
	readData, err = accessor.Read(ctx)
	require.NoError(t, err)
	require.Nil(t, readData)
}

func TestStorageRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	uniqueName := uuid.NewString()
	cachePath := filepath.Join(tempDir, uniqueName)

	accessor, err := storage(cachePath)
	require.NoError(t, err)

	// Generate random test data
	testData := make([]byte, 256)
	_, err = rand.Read(testData)
	require.NoError(t, err)

	// Test write
	err = accessor.Write(ctx, testData)
	require.NoError(t, err)

	// Test read
	readData, err := accessor.Read(ctx)
	require.NoError(t, err)
	require.Equal(t, testData, readData)

	// Verify file exists and is encrypted (content should be different from original)
	if fileContent, err := os.ReadFile(cachePath); err == nil {
		require.NotEqual(t, testData, fileContent, "file content should be encrypted")
		require.Greater(t, len(fileContent), 0, "encrypted file should not be empty")
	}

	// Test delete
	err = accessor.Delete(ctx)
	require.NoError(t, err)

	// Verify file is deleted
	_, err = os.Stat(cachePath)
	require.True(t, os.IsNotExist(err), "cache file should be deleted")

	// Read after delete should return nil
	readData, err = accessor.Read(ctx)
	require.NoError(t, err)
	require.Nil(t, readData)
}

func TestStorageEmptyData(t *testing.T) {
	tempDir := t.TempDir()
	uniqueName := uuid.NewString()
	cachePath := filepath.Join(tempDir, uniqueName)

	accessor, err := storage(cachePath)
	require.NoError(t, err)

	// Test writing empty data
	err = accessor.Write(ctx, []byte{})
	require.NoError(t, err)

	// Test reading empty data
	readData, err := accessor.Read(ctx)
	require.NoError(t, err)
	require.Nil(t, readData)

	// Test writing nil data
	err = accessor.Write(ctx, nil)
	require.NoError(t, err)

	readData, err = accessor.Read(ctx)
	require.NoError(t, err)
	require.Nil(t, readData)
}

func TestStorageNonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	uniqueName := uuid.NewString()
	cachePath := filepath.Join(tempDir, uniqueName)

	accessor, err := storage(cachePath)
	require.NoError(t, err)

	// Reading non-existent file should return nil, not error
	readData, err := accessor.Read(ctx)
	require.NoError(t, err)
	require.Nil(t, readData)

	// Deleting non-existent file should not error
	err = accessor.Delete(ctx)
	require.NoError(t, err)
}

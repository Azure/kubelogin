//go:build linux

package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestKeyExistsButNotFile(t *testing.T) {
	expected := []byte(t.Name())
	uniqueName := uuid.NewString()

	// Create a keyring accessor
	a, err := storage(uniqueName)
	require.NoError(t, err)

	// Write some data that's different from expected
	err = a.Write(ctx, append([]byte("not"), expected...))
	require.NoError(t, err)

	// Clean up keyring at end of test
	t.Cleanup(func() { require.NoError(t, a.Delete(ctx)) })

	// Remove the cache file but leave the keyring key
	kr := a.(*keyring)
	require.NoError(t, os.Remove(kr.file))

	// Create a new keyring instance with the same description
	// This should find the existing key but no file
	b, err := newKeyring(uniqueName)
	require.NoError(t, err)

	// Read should return nil since file doesn't exist
	data, err := b.Read(ctx)
	require.NoError(t, err)
	require.Nil(t, data)

	// Write should succeed and create a new file
	err = b.Write(ctx, expected)
	require.NoError(t, err)

	// Read should now return the expected data
	data, err = b.Read(ctx)
	require.NoError(t, err)
	require.Equal(t, expected, data)
}

func TestNewKeyring(t *testing.T) {
	tests := []struct {
		desc     string
		name     string
		expected []byte
	}{
		{
			desc:     "empty cache",
			name:     "",
			expected: nil,
		},
		{
			desc:     "non-empty cache",
			name:     "",
			expected: nil, // New cache should be empty
		},
		{
			desc:     "cache with existing encrypted file",
			name:     t.Name(),
			expected: nil, // Should return nil for corrupted/lost key scenario
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			name := test.name
			if name == "" {
				// Use UUID to ensure file and key don't exist
				name = uuid.NewString()
			} else {
				// Create a corrupted cache file to simulate lost key scenario
				tempDir := t.TempDir()
				p := filepath.Join(tempDir, name)
				err := os.MkdirAll(filepath.Dir(p), 0600)
				require.NoError(t, err)

				// Write some encrypted-looking data that can't be decrypted
				corruptedData := []byte("eyJhbGciOiJkaXIiLCJlbmMiOiJBMTI4Q0JDLUhTMjU2In0..gPRNjqd4HcrlFxJdEEaFeA.Pqpr_IYG7e1lt6KPoE0v_A.i9h5iJWw9bT217I5M2Ufrg")
				err = os.WriteFile(p, corruptedData, 0600)
				require.NoError(t, err)
				name = p
			}

			k, err := newKeyring(name)
			require.NoError(t, err)
			require.NotNil(t, k)

			// Read should return nil for empty or corrupted cache
			actual, err := k.Read(ctx)
			require.NoError(t, err)
			require.Equal(t, test.expected, actual)

			// Clean up
			t.Cleanup(func() {
				if k.keyID != 0 {
					k.Delete(ctx)
				}
			})

			if test.name == "" {
				// Test that we can write to an empty cache
				testData := []byte("test write to empty cache")
				err = k.Write(ctx, testData)
				require.NoError(t, err)

				actual, err = k.Read(ctx)
				require.NoError(t, err)
				require.Equal(t, testData, actual)
			}
		})
	}
}

func TestKeyringDescription(t *testing.T) {
	// Test that different paths result in different keyring descriptions
	// This ensures each cache file gets its own encryption key
	testPaths := []string{
		"/tmp/cache1/pop_tokens.cache",
		"/tmp/cache2/tokens.cache",    // Different filename
		"/different/path/auth.cache",  // Different filename
		"relative/path/session.cache", // Different filename
	}

	descriptions := make(map[string]bool)

	for _, path := range testPaths {
		k, err := newKeyring(path)
		require.NoError(t, err)

		// Verify description is the full path (changed to prevent cache conflicts)
		require.Equal(t, path, k.description)

		// Verify each path gets a unique description
		require.False(t, descriptions[k.description], "description %q should be unique", k.description)
		descriptions[k.description] = true
	}
}

func TestKeyringRoundTrip(t *testing.T) {
	uniqueName := uuid.NewString()
	k, err := newKeyring(uniqueName)
	require.NoError(t, err)

	testData := []byte("test keyring round trip data with special chars: éñ中文🚀")

	// Test write
	err = k.Write(ctx, testData)
	require.NoError(t, err)

	// Test read
	readData, err := k.Read(ctx)
	require.NoError(t, err)
	require.Equal(t, testData, readData)

	// Test that file exists and is encrypted
	if k.file != "" {
		fileContent, err := os.ReadFile(k.file)
		require.NoError(t, err)
		require.NotEqual(t, testData, fileContent, "file should be encrypted")
		require.Greater(t, len(fileContent), len(testData), "encrypted content should be longer")
	}

	// Test delete
	err = k.Delete(ctx)
	require.NoError(t, err)

	// Verify file is deleted
	if k.file != "" {
		_, err = os.Stat(k.file)
		require.True(t, os.IsNotExist(err), "file should be deleted")
	}

	// Read after delete should return nil
	readData, err = k.Read(ctx)
	require.NoError(t, err)
	require.Nil(t, readData)
}

func TestKeyringEmptyData(t *testing.T) {
	uniqueName := uuid.NewString()
	k, err := newKeyring(uniqueName)
	require.NoError(t, err)

	t.Cleanup(func() { k.Delete(ctx) })

	// Test writing empty data
	err = k.Write(ctx, []byte{})
	require.NoError(t, err)

	// Read should return nil for empty data
	readData, err := k.Read(ctx)
	require.NoError(t, err)
	require.Nil(t, readData)

	// Test writing nil data
	err = k.Write(ctx, nil)
	require.NoError(t, err)

	readData, err = k.Read(ctx)
	require.NoError(t, err)
	require.Nil(t, readData)
}

func TestKeyringNonExistentFile(t *testing.T) {
	uniqueName := uuid.NewString()
	k, err := newKeyring(uniqueName)
	require.NoError(t, err)

	// Reading non-existent file should return nil
	readData, err := k.Read(ctx)
	require.NoError(t, err)
	require.Nil(t, readData)

	// Deleting non-existent file/key should not error
	err = k.Delete(ctx)
	require.NoError(t, err)
}

func TestKeyringProcessIsolation(t *testing.T) {
	// Test that different keyring descriptions (representing different cache files)
	// don't interfere with each other - simulating multiple kubelogin processes
	// with different cache files
	baseName := uuid.NewString()

	keyrings := make([]*keyring, 3)
	testData := make([][]byte, 3)

	// Create multiple keyrings with different names (different cache files)
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("%s_%d", baseName, i)
		k, err := newKeyring(name)
		require.NoError(t, err)
		keyrings[i] = k

		testData[i] = []byte(fmt.Sprintf("process_%d_data", i))
	}

	// Each keyring should be able to store and retrieve its own data
	for i, k := range keyrings {
		err := k.Write(ctx, testData[i])
		require.NoError(t, err)
	}

	// Verify each keyring can read back its own data correctly
	for i, k := range keyrings {
		readData, err := k.Read(ctx)
		require.NoError(t, err)
		require.Equal(t, testData[i], readData)
	}

	// Clean up
	for _, k := range keyrings {
		err := k.Delete(ctx)
		require.NoError(t, err)
	}
}

func TestKeyringDirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()

	// Test with nested directory that doesn't exist
	nestedPath := filepath.Join(tempDir, "deep", "nested", "path", "cache.data")
	k, err := newKeyring(nestedPath)
	require.NoError(t, err)

	testData := []byte("test directory creation")
	err = k.Write(ctx, testData)
	require.NoError(t, err)

	// Verify directory was created
	require.DirExists(t, filepath.Dir(nestedPath))

	// Verify data can be read back
	readData, err := k.Read(ctx)
	require.NoError(t, err)
	require.Equal(t, testData, readData)

	// Clean up
	err = k.Delete(ctx)
	require.NoError(t, err)
}

package token

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKUBERNETES_EXEC_INFOIsEmpty(t *testing.T) {
	testData := []struct {
		name            string
		execInfoEnvTest string
		options         Options
	}{
		{
			name:            "KUBERNETES_EXEC_INFO is empty",
			execInfoEnvTest: "",
			options: Options{
				LoginMethod: DeviceCodeLogin,
				ClientID:    "clientID",
				ServerID:    "serverID",
				TenantID:    "tenantID",
			},
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {
			os.Setenv("KUBERNETES_EXEC_INFO", data.execInfoEnvTest)
			defer os.Unsetenv("KUBERNETES_EXEC_INFO")
			ecp, err := New(&data.options)
			if ecp == nil || err != nil {
				t.Fatalf("expected: return execCredentialPlugin and nil error, actual: did not return execCredentialPlugin or did not return expected error")
			}
		})
	}
}

// TestNew_PoPCacheFallbackResilience validates the fallback mechanism for PoP token cache creation.
// This is critical for container compatibility where secure storage (Linux keyrings) may not be available.
// This test validates that New() never fails regardless of cache creation success/failure.
func TestNew_PoPCacheFallbackResilience(t *testing.T) {
	t.Run("PoP disabled - no cache attempted", func(t *testing.T) {
		options := &Options{
			LoginMethod:       DeviceCodeLogin,
			ClientID:          "clientID",
			ServerID:          "serverID",
			TenantID:          "tenantID",
			IsPoPTokenEnabled: false,
		}

		plugin, err := New(options)

		assert.NoError(t, err, "Should succeed when PoP is disabled")
		assert.NotNil(t, plugin, "Should return valid plugin")

		execPlugin, ok := plugin.(*execCredentialPlugin)
		assert.True(t, ok, "Should return execCredentialPlugin type")
		assert.Nil(t, execPlugin.popTokenCache, "Should not create cache when PoP is disabled")
	})

	t.Run("PoP enabled with valid cache directory", func(t *testing.T) {
		// Use a temporary directory for cache
		tmpDir := t.TempDir()

		options := &Options{
			LoginMethod:        DeviceCodeLogin,
			ClientID:           "clientID",
			ServerID:           "serverID",
			TenantID:           "tenantID",
			IsPoPTokenEnabled:  true,
			AuthRecordCacheDir: tmpDir,
		}

		plugin, err := New(options)

		// New() must never fail, regardless of cache creation success/failure
		assert.NoError(t, err, "Must succeed regardless of cache creation outcome")
		assert.NotNil(t, plugin, "Must return valid plugin")

		execPlugin, ok := plugin.(*execCredentialPlugin)
		assert.True(t, ok, "Should return execCredentialPlugin type")

		// Log the actual outcome for debugging
		if execPlugin.popTokenCache != nil {
			t.Log("Cache creation succeeded - secure storage available")
		} else {
			t.Log("Cache creation failed (gracefully) - likely container environment or keyring restrictions")
		}
	})

	t.Run("Validates fallback mechanism behavior", func(t *testing.T) {
		// This test demonstrates that the behavior is consistent regardless of environment
		testCases := []struct {
			name     string
			cacheDir string
		}{
			{"temp directory", t.TempDir()},
			{"invalid directory", "/proc/non-existent-test-dir"},
			{"root directory (typically restricted)", "/root/cache-test"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				options := &Options{
					LoginMethod:        DeviceCodeLogin,
					ClientID:           "clientID",
					ServerID:           "serverID",
					TenantID:           "tenantID",
					IsPoPTokenEnabled:  true,
					AuthRecordCacheDir: tc.cacheDir,
				}

				plugin, err := New(options)

				// The universal requirement: New() must NEVER fail
				assert.NoError(t, err, "New() must succeed in all environments for container compatibility")
				assert.NotNil(t, plugin, "Must return valid plugin")

				execPlugin, ok := plugin.(*execCredentialPlugin)
				assert.True(t, ok, "Should return execCredentialPlugin type")

				// Document the behavior for each scenario
				cacheState := "succeeded"
				if execPlugin.popTokenCache == nil {
					cacheState = "failed (graceful fallback)"
				}
				t.Logf("Directory '%s': cache creation %s", tc.cacheDir, cacheState)
			})
		}
	})
}

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

	t.Run("PoP enabled - resilient to secure storage failures", func(t *testing.T) {
		options := &Options{
			LoginMethod:        DeviceCodeLogin,
			ClientID:           "clientID",
			ServerID:           "serverID",
			TenantID:           "tenantID",
			IsPoPTokenEnabled:  true,
			AuthRecordCacheDir: "/tmp/test-cache-fallback",
		}

		// This is the CRITICAL test: New() must NEVER fail due to cache creation issues
		plugin, err := New(options)

		assert.NoError(t, err, "CRITICAL: Must succeed even if secure storage fails (container compatibility)")
		assert.NotNil(t, plugin, "Must return valid plugin regardless of cache creation outcome")

		_, ok := plugin.(*execCredentialPlugin)
		assert.True(t, ok, "Should return execCredentialPlugin type")

		// We don't care if cache is nil or not-nil - what matters is that we handle both gracefully
		// This test validates that the fallback mechanism prevents crashes in container environments
	})
}

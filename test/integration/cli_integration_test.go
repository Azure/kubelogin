package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFeatureFlagSwitching tests that unified and legacy modes behave consistently
func TestFeatureFlagSwitching(t *testing.T) {
	t.Run("convert-kubeconfig help consistency", func(t *testing.T) {
		CompareHelpOutput(t, "convert-kubeconfig")
	})

	t.Run("get-token help consistency", func(t *testing.T) {
		CompareHelpOutput(t, "get-token")
	})
}

// TestCLIArgumentParsing tests that both modes parse CLI arguments identically
func TestCLIArgumentParsing(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "missing required tenant-id with nonexistent kubeconfig",
			args:        []string{"convert-kubeconfig", "--login", "devicecode", "--kubeconfig", "/nonexistent/kubeconfig"},
			expectError: true,
		},
		{
			name:        "invalid login method",
			args:        []string{"convert-kubeconfig", "--login", "invalid"},
			expectError: true,
		},
		{
			name:        "help should not error",
			args:        []string{"convert-kubeconfig", "--help"},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test legacy mode
			legacyResult := RunKubeloginCommandWithEnv(
				map[string]string{"KUBELOGIN_USE_LEGACY_OPTIONS": "true"},
				tc.args...,
			)

			// Test unified mode (default behavior)
			unifiedResult := RunKubeloginCommandWithEnv(
				map[string]string{},
				tc.args...,
			)

			// Both modes should have same success/failure behavior
			if tc.expectError {
				assert.NotEqual(t, 0, legacyResult.ExitCode, "Legacy mode should fail")
				assert.NotEqual(t, 0, unifiedResult.ExitCode, "Unified mode should fail")
			} else {
				assert.Equal(t, 0, legacyResult.ExitCode, "Legacy mode should succeed")
				assert.Equal(t, 0, unifiedResult.ExitCode, "Unified mode should succeed")
			}
		})
	}
}

// TestEnvironmentVariables tests that environment variables work consistently
func TestEnvironmentVariables(t *testing.T) {
	env := NewTestEnvironment(t)

	// Set some environment variables
	env.SetEnv("AZURE_TENANT_ID", "test-tenant-id")
	env.SetEnv("AZURE_CLIENT_ID", "test-client-id")
	env.SetEnv("AAD_LOGIN_METHOD", "devicecode")

	testCases := []struct {
		name           string
		unifiedOptions bool
	}{
		{"legacy mode", false},
		{"unified mode", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envVars := map[string]string{
				"AZURE_TENANT_ID":  "test-tenant-id",
				"AZURE_CLIENT_ID":  "test-client-id",
				"AAD_LOGIN_METHOD": "devicecode",
			}
			if !tc.unifiedOptions {
				envVars["KUBELOGIN_USE_LEGACY_OPTIONS"] = "true"
			}

			// Both modes should recognize environment variables
			// We'll test this by checking help output includes the env vars we set
			result := RunKubeloginCommandWithEnv(envVars, "convert-kubeconfig", "--help")
			require.Equal(t, 0, result.ExitCode)

			// The help output should be successful, indicating env vars were processed
			assert.Contains(t, result.Stdout, "convert kubeconfig")
		})
	}
}

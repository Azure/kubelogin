package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKubeconfigFlagIntegration(t *testing.T) {
	t.Run("legacy mode should support --kubeconfig flag", func(t *testing.T) {
		// Test that legacy mode recognizes --kubeconfig flag
		result := RunKubeloginCommandWithEnv(
			map[string]string{"KUBELOGIN_USE_LEGACY_OPTIONS": "true"},
			"convert-kubeconfig", "--kubeconfig", "/nonexistent/path", "--help",
		)
		// Should not fail with "unknown flag" error
		assert.NotContains(t, result.Stderr, "unknown flag: --kubeconfig")
	})

	t.Run("unified mode should support --kubeconfig flag", func(t *testing.T) {
		// Test that unified mode recognizes --kubeconfig flag
		result := RunKubeloginCommandWithEnv(
			map[string]string{}, // Default is unified mode
			"convert-kubeconfig", "--kubeconfig", "/nonexistent/path", "--help",
		)
		// Should not fail with "unknown flag" error (this test should initially fail)
		assert.NotContains(t, result.Stderr, "unknown flag: --kubeconfig",
			"Unified mode should support --kubeconfig flag. Stderr: %s", result.Stderr)
		assert.NotEqual(t, 1, result.ExitCode, "Command should not exit with error code 1 due to unknown flag")
	})

	t.Run("kubeconfig flag should appear in help output", func(t *testing.T) {
		// Test that --kubeconfig appears in help
		result := RunKubeloginCommandWithEnv(
			map[string]string{}, // Default is unified mode
			"convert-kubeconfig", "--help",
		)
		assert.Equal(t, 0, result.ExitCode)
		// Debug: print the actual help output
		t.Logf("Unified help output:\n%s", result.Stdout)
		assert.Contains(t, result.Stdout, "--kubeconfig",
			"--kubeconfig flag should appear in convert-kubeconfig help")
	})

	t.Run("legacy mode has kubeconfig flag in help", func(t *testing.T) {
		// Verify legacy mode has the flag (for comparison)
		result := RunKubeloginCommandWithEnv(
			map[string]string{"KUBELOGIN_USE_LEGACY_OPTIONS": "true"},
			"convert-kubeconfig", "--help",
		)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "--kubeconfig",
			"Legacy mode should have --kubeconfig flag in help")
	})
}

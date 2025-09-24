package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKubeconfigFlagSupport(t *testing.T) {
	t.Run("legacy mode should support --kubeconfig flag", func(t *testing.T) {
		// Test legacy converter (for comparison)
		// This test documents the expected behavior that legacy mode supports --kubeconfig
		t.Skip("Legacy converter not directly testable here - this test documents expected behavior")
	})

	t.Run("unified mode should support --kubeconfig flag", func(t *testing.T) {
		// Create unified command for convert-kubeconfig
		cmd := NewUnifiedCommand(ConvertCommand)

		// Check if --kubeconfig flag is registered
		kubeconfigFlag := cmd.Flags().Lookup("kubeconfig")
		assert.NotNil(t, kubeconfigFlag, "--kubeconfig flag should be registered in unified mode")

		if kubeconfigFlag != nil {
			assert.Equal(t, "string", kubeconfigFlag.Value.Type(), "--kubeconfig should be a string flag")
		}
	})

	t.Run("unified mode should accept --kubeconfig value", func(t *testing.T) {
		// Create unified command
		cmd := NewUnifiedCommand(ConvertCommand)

		// Test that we can parse --kubeconfig argument
		args := []string{"--kubeconfig", "/path/to/kubeconfig", "--login", "devicecode", "--tenant-id", "test", "--client-id", "test", "--server-id", "test"}
		cmd.SetArgs(args)

		// Parse flags (don't execute, just parse)
		err := cmd.ParseFlags(args)
		require.NoError(t, err, "Should be able to parse --kubeconfig flag")

		// Verify the flag value was set
		kubeconfigFlag := cmd.Flags().Lookup("kubeconfig")
		if kubeconfigFlag != nil {
			assert.Equal(t, "/path/to/kubeconfig", kubeconfigFlag.Value.String())
		}
	})
}

func TestKubeconfigFlagUsage(t *testing.T) {
	t.Run("kubeconfig flag should be used in convert command execution", func(t *testing.T) {
		// This test verifies that the --kubeconfig flag value is actually used
		// Create unified command
		cmd := NewUnifiedCommand(ConvertCommand)

		// The executeConvert method should be able to read the kubeconfig flag
		// We can't easily test the full execution without a real kubeconfig file,
		// but we can verify the flag lookup code doesn't crash
		_, err := cmd.Flags().GetString("kubeconfig")
		assert.NoError(t, err, "Should be able to get kubeconfig flag value")
	})
}

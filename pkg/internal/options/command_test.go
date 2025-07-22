package options

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUnifiedCommand(t *testing.T) {
	tests := []struct {
		name     string
		cmdType  CommandType
		wantUse  string
		wantArgs bool // Whether command requires args
	}{
		{
			name:     "convert command",
			cmdType:  ConvertCommand,
			wantUse:  "convert-kubeconfig",
			wantArgs: false,
		},
		{
			name:     "token command",
			cmdType:  TokenCommand,
			wantUse:  "get-token",
			wantArgs: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewUnifiedCommand(tt.cmdType)

			assert.Equal(t, tt.wantUse, cmd.Use)
			assert.NotEmpty(t, cmd.Short)
			assert.NotNil(t, cmd.RunE)

			// Check that flags are registered
			flags := cmd.Flags()
			assert.NotNil(t, flags.Lookup("login"))
			assert.NotNil(t, flags.Lookup("tenant-id"))

			// Test command-specific flags
			if tt.cmdType == ConvertCommand {
				assert.NotNil(t, flags.Lookup("context"), "convert command should have context flag")
				assert.NotNil(t, flags.Lookup("azure-config-dir"), "convert command should have azure-config-dir flag")
			}
			if tt.cmdType == TokenCommand {
				assert.NotNil(t, flags.Lookup("server-id"), "token command should have server-id flag")
				assert.Nil(t, flags.Lookup("context"), "token command should NOT have context flag")
				assert.Nil(t, flags.Lookup("azure-config-dir"), "token command should NOT have azure-config-dir flag")
			}
		})
	}
}

func TestUseUnifiedOptions(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name:     "legacy options enabled",
			envValue: "true",
			expected: false,
		},
		{
			name:     "legacy options disabled",
			envValue: "false",
			expected: true,
		},
		{
			name:     "legacy options unset - defaults to unified",
			envValue: "",
			expected: true,
		},
		{
			name:     "legacy options invalid value - defaults to unified",
			envValue: "invalid",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment
			os.Unsetenv("KUBELOGIN_USE_LEGACY_OPTIONS")

			if tt.envValue != "" {
				os.Setenv("KUBELOGIN_USE_LEGACY_OPTIONS", tt.envValue)
			}

			// Clean up after test
			defer func() {
				os.Unsetenv("KUBELOGIN_USE_LEGACY_OPTIONS")
			}()

			result := UseUnifiedOptions()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCommandUse(t *testing.T) {
	tests := []struct {
		name     string
		cmdType  CommandType
		expected string
	}{
		{
			name:     "convert command",
			cmdType:  ConvertCommand,
			expected: "convert-kubeconfig",
		},
		{
			name:     "token command",
			cmdType:  TokenCommand,
			expected: "get-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCommandUse(tt.cmdType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCommandShort(t *testing.T) {
	tests := []struct {
		name     string
		cmdType  CommandType
		expected string
	}{
		{
			name:     "convert command",
			cmdType:  ConvertCommand,
			expected: "convert kubeconfig to use exec auth module",
		},
		{
			name:     "token command",
			cmdType:  TokenCommand,
			expected: "get AAD token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCommandShort(tt.cmdType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommandTypeConstants(t *testing.T) {
	// Test that command type constants have expected values
	assert.Equal(t, CommandType(0), ConvertCommand)
	assert.Equal(t, CommandType(1), TokenCommand)

	// Test that they are different
	assert.NotEqual(t, ConvertCommand, TokenCommand)
}

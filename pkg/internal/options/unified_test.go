package options

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUnifiedOptions(t *testing.T) {
	tests := []struct {
		name        string
		cmdType     CommandType
		wantCommand CommandType
	}{
		{
			name:        "convert command",
			cmdType:     ConvertCommand,
			wantCommand: ConvertCommand,
		},
		{
			name:        "token command",
			cmdType:     TokenCommand,
			wantCommand: TokenCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewUnifiedOptions(tt.cmdType)
			assert.Equal(t, tt.wantCommand, opts.command)
		})
	}
}

func TestRegisterFlags(t *testing.T) {
	opts := NewUnifiedOptions(TokenCommand)
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)

	err := opts.RegisterFlags(fs)
	require.NoError(t, err)

	// Test that flags are registered
	assert.NotNil(t, fs.Lookup("login"))
	assert.NotNil(t, fs.Lookup("client-id"))
	assert.NotNil(t, fs.Lookup("tenant-id"))
	assert.NotNil(t, fs.Lookup("server-id"))
	assert.NotNil(t, fs.Lookup("environment"))
	assert.NotNil(t, fs.Lookup("timeout"))

	// Test shorthand flags
	loginFlag := fs.Lookup("login")
	assert.Equal(t, "l", loginFlag.Shorthand)

	tenantFlag := fs.Lookup("tenant-id")
	assert.Equal(t, "t", tenantFlag.Shorthand)
}

func TestLoadFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected map[string]interface{}
	}{
		{
			name: "Azure environment variables",
			envVars: map[string]string{
				"AZURE_CLIENT_ID":     "test-client-id",
				"AZURE_CLIENT_SECRET": "test-secret",
				"AZURE_TENANT_ID":     "test-tenant",
			},
			expected: map[string]interface{}{
				"ClientID":     "test-client-id",
				"ClientSecret": "test-secret",
				"TenantID":     "test-tenant",
			},
		},
		{
			name: "AAD environment variables",
			envVars: map[string]string{
				"AAD_SERVICE_PRINCIPAL_CLIENT_ID":     "aad-client-id",
				"AAD_SERVICE_PRINCIPAL_CLIENT_SECRET": "aad-secret",
			},
			expected: map[string]interface{}{
				"ClientID":     "aad-client-id",
				"ClientSecret": "aad-secret",
			},
		},
		{
			name: "Login method override",
			envVars: map[string]string{
				"AAD_LOGIN_METHOD": "spn",
			},
			expected: map[string]interface{}{
				"LoginMethod": "spn",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			for key := range tt.envVars {
				os.Unsetenv(key)
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Clean up after test
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			opts := NewUnifiedOptions(TokenCommand)
			opts.LoadFromEnv()

			// Check expected values
			for field, expectedValue := range tt.expected {
				switch field {
				case "ClientID":
					assert.Equal(t, expectedValue, opts.ClientID)
				case "ClientSecret":
					assert.Equal(t, expectedValue, opts.ClientSecret)
				case "TenantID":
					assert.Equal(t, expectedValue, opts.TenantID)
				case "LoginMethod":
					assert.Equal(t, expectedValue, opts.LoginMethod)
				}
			}
		})
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name      string
		setupOpts func(*UnifiedOptions)
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid devicecode options",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "devicecode"
				opts.ClientID = "test-client"
				opts.TenantID = "test-tenant"
				opts.ServerID = "test-server"
			},
			wantError: false,
		},
		{
			name: "missing required fields",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "devicecode"
				// Missing ClientID, TenantID, ServerID
			},
			wantError: true,
			errorMsg:  "validation failed",
		},
		{
			name: "invalid login method",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "invalid"
				opts.ClientID = "test-client"
				opts.TenantID = "test-tenant"
				opts.ServerID = "test-server"
			},
			wantError: true,
			errorMsg:  "not a supported login method",
		},
		{
			name: "pop token validation",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "spn"
				opts.ClientID = "test-client"
				opts.TenantID = "test-tenant"
				opts.ServerID = "test-server"
				opts.ClientSecret = "test-secret"
				opts.IsPoPTokenEnabled = true
				// Missing PoPTokenClaims
			},
			wantError: true,
			errorMsg:  "pop-claims flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewUnifiedOptions(TokenCommand)
			tt.setupOpts(opts)

			err := opts.ValidateForTokenExecution()

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestToString(t *testing.T) {
	opts := NewUnifiedOptions(TokenCommand)
	opts.ClientID = "test-client"
	opts.ClientSecret = "secret-value"
	opts.TenantID = "test-tenant"

	result := opts.ToString()

	// Should contain non-sensitive values
	assert.Contains(t, result, "test-client")
	assert.Contains(t, result, "test-tenant")

	// Should mask sensitive values
	assert.NotContains(t, result, "secret-value")
	assert.Contains(t, result, "***")
}

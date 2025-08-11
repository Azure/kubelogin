package options

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidationRules(t *testing.T) {
	tests := []struct {
		name          string
		setupOpts     func(*UnifiedOptions)
		wantError     bool
		errorContains string
	}{
		{
			name: "valid devicecode configuration",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "devicecode"
				opts.ClientID = "test-client"
				opts.TenantID = "test-tenant"
				opts.ServerID = "test-server"
			},
			wantError: false,
		},
		{
			name: "valid spn with secret configuration",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "spn"
				opts.ClientID = "test-client"
				opts.ClientSecret = "test-secret"
				opts.TenantID = "test-tenant"
				opts.ServerID = "test-server"
			},
			wantError: false,
		},
		{
			name: "valid spn with certificate configuration",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "spn"
				opts.ClientID = "test-client"
				opts.ClientCert = "/path/to/cert.pem"
				opts.TenantID = "test-tenant"
				opts.ServerID = "test-server"
			},
			wantError: false,
		},
		{
			name: "invalid login method",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "invalid"
				opts.ClientID = "test-client"
				opts.TenantID = "test-tenant"
				opts.ServerID = "test-server"
			},
			wantError:     true,
			errorContains: "not a supported login method",
		},
		{
			name: "missing tenant id",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "devicecode"
				opts.ClientID = "test-client"
				opts.ServerID = "test-server"
				// Missing TenantID
			},
			wantError:     true,
			errorContains: "tenantid is required",
		},
		{
			name: "missing server id for token command",
			setupOpts: func(opts *UnifiedOptions) {
				opts.command = TokenCommand
				opts.LoginMethod = "devicecode"
				opts.ClientID = "test-client"
				opts.TenantID = "test-tenant"
				// Missing ServerID
			},
			wantError:     true,
			errorContains: "serverid is required",
		},
		{
			name: "server id not required for convert command",
			setupOpts: func(opts *UnifiedOptions) {
				opts.command = ConvertCommand
				opts.LoginMethod = "devicecode"
				opts.ClientID = "test-client"
				opts.TenantID = "test-tenant"
				// Missing ServerID - should be OK for convert
			},
			wantError: false,
		},
		{
			name: "spn missing both secret and certificate",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "spn"
				opts.ClientID = "test-client"
				opts.TenantID = "test-tenant"
				opts.ServerID = "test-server"
				// Missing both ClientSecret and ClientCert
			},
			wantError:     true,
			errorContains: "clientsecret is required",
		},
		{
			name: "pop token requires claims",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "spn"
				opts.ClientID = "test-client"
				opts.ClientSecret = "test-secret"
				opts.TenantID = "test-tenant"
				opts.ServerID = "test-server"
				opts.IsPoPTokenEnabled = true
				// Missing PoPTokenClaims
			},
			wantError:     true,
			errorContains: "pop-claims flag",
		},
		{
			name: "ropc requires username and password",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "ropc"
				opts.ClientID = "test-client"
				opts.TenantID = "test-tenant"
				opts.ServerID = "test-server"
				// Missing Username and Password
			},
			wantError:     true,
			errorContains: "username is required",
		},
		{
			name: "workload identity requires federated token file",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "workloadidentity"
				opts.ClientID = "test-client"
				opts.TenantID = "test-tenant"
				opts.ServerID = "test-server"
				// Missing FederatedTokenFile
			},
			wantError:     true,
			errorContains: "federatedtokenfile is required",
		},
		{
			name: "invalid authority host URL",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "devicecode"
				opts.ClientID = "test-client"
				opts.TenantID = "test-tenant"
				opts.ServerID = "test-server"
				opts.AuthorityHost = "not-a-url"
			},
			wantError:     true,
			errorContains: "authority host",
		},
		{
			name: "invalid timeout",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "devicecode"
				opts.ClientID = "test-client"
				opts.TenantID = "test-tenant"
				opts.ServerID = "test-server"
				opts.Timeout = -1 * time.Second
			},
			wantError:     true,
			errorContains: "timeout must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewUnifiedOptions(TokenCommand)
			tt.setupOpts(opts)

			err := opts.ValidateForTokenExecution()

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateLoginMethod(t *testing.T) {
	tests := []struct {
		name        string
		loginMethod string
		wantError   bool
	}{
		{"devicecode", "devicecode", false},
		{"interactive", "interactive", false},
		{"spn", "spn", false},
		{"ropc", "ropc", false},
		{"msi", "msi", false},
		{"azurecli", "azurecli", false},
		{"azd", "azd", false},
		{"workloadidentity", "workloadidentity", false},
		{"invalid", "invalid", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewUnifiedOptions(TokenCommand)
			opts.LoginMethod = tt.loginMethod

			err := opts.ValidateForTokenExecution()

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not a supported login method")
			} else if tt.loginMethod != "" {
				// If we have other missing required fields, we expect validation to fail for those
				// but not for login method
				if err != nil {
					assert.NotContains(t, err.Error(), "not a supported login method")
				}
			}
		})
	}
}

func TestValidateRequiredFields(t *testing.T) {
	tests := []struct {
		name      string
		setupOpts func(*UnifiedOptions)
		cmdType   CommandType
		wantError bool
	}{
		{
			name: "token command with all required",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "devicecode"
				opts.ClientID = "test-client"
				opts.TenantID = "test-tenant"
				opts.ServerID = "test-server"
			},
			cmdType:   TokenCommand,
			wantError: false,
		},
		{
			name: "token command missing server id",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "devicecode"
				opts.ClientID = "test-client"
				opts.TenantID = "test-tenant"
			},
			cmdType:   TokenCommand,
			wantError: true,
		},
		{
			name: "convert command without server id",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "devicecode"
				opts.ClientID = "test-client"
				opts.TenantID = "test-tenant"
			},
			cmdType:   ConvertCommand,
			wantError: false,
		},
		{
			name: "missing tenant id",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "devicecode"
				opts.ClientID = "test-client"
				opts.ServerID = "test-server"
			},
			cmdType:   TokenCommand,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewUnifiedOptions(tt.cmdType)
			tt.setupOpts(opts)

			err := opts.ValidateForTokenExecution()

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateLoginMethodSpecific(t *testing.T) {
	tests := []struct {
		name      string
		setupOpts func(*UnifiedOptions)
		wantError bool
	}{
		{
			name: "spn with secret",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "spn"
				opts.ClientID = "test-client"
				opts.ClientSecret = "test-secret"
				opts.TenantID = "test-tenant"
				opts.ServerID = "test-server"
			},
			wantError: false,
		},
		{
			name: "spn with certificate",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "spn"
				opts.ClientID = "test-client"
				opts.ClientCert = "/path/to/cert.pem"
				opts.TenantID = "test-tenant"
				opts.ServerID = "test-server"
			},
			wantError: false,
		},
		{
			name: "spn missing both secret and cert",
			setupOpts: func(opts *UnifiedOptions) {
				opts.LoginMethod = "spn"
				opts.ClientID = "test-client"
				opts.TenantID = "test-tenant"
				opts.ServerID = "test-server"
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewUnifiedOptions(TokenCommand)
			tt.setupOpts(opts)

			err := opts.ValidateForTokenExecution()

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

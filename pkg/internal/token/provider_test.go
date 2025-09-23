package token

import (
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/kubelogin/pkg/internal/env"
	"github.com/stretchr/testify/assert"
)

func TestNewAzIdentityCredential(t *testing.T) {
	certFile := "fixtures/cert.pem"

	// Set up environment variables for Azure Pipelines test
	os.Setenv(env.SystemAccessToken, "test-system-access-token")
	os.Setenv(env.SystemOIDCRequestURI, "https://test.oidc.request.uri")
	defer func() {
		os.Unsetenv(env.SystemAccessToken)
		os.Unsetenv(env.SystemOIDCRequestURI)
	}()

	tests := []struct {
		name       string
		options    *Options
		wantErr    bool
		errMessage string
	}{
		{
			name: "Azure CLI login",
			options: &Options{
				LoginMethod: AzureCLILogin,
				ServerID:    "server-id",
				TenantID:    "tenant-id",
			},
			wantErr: false,
		},
		{
			name: "Device code login",
			options: &Options{
				LoginMethod: DeviceCodeLogin,
				ServerID:    "server-id",
				TenantID:    "tenant-id",
				ClientID:    "client-id",
			},
			wantErr: false,
		},
		{
			name: "Legacy device code login",
			options: &Options{
				LoginMethod: DeviceCodeLogin,
				ServerID:    "server-id",
				TenantID:    "tenant-id",
				ClientID:    "client-id",
				IsLegacy:    true,
			},
			wantErr: false,
		},
		{
			name: "Interactive login with PoP",
			options: &Options{
				LoginMethod:       InteractiveLogin,
				ServerID:          "server-id",
				TenantID:          "tenant-id",
				ClientID:          "client-id",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=test-cluster",
			},
			wantErr: false,
		},
		{
			name: "MSI login",
			options: &Options{
				LoginMethod: MSILogin,
				ServerID:    "server-id",
			},
			wantErr: false,
		},
		{
			name: "Service Principal with client cert and PoP",
			options: &Options{
				LoginMethod:       ServicePrincipalLogin,
				ServerID:          "server-id",
				TenantID:          "tenant-id",
				ClientID:          "client-id",
				ClientCert:        certFile,
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=test-cluster",
			},
			wantErr: false,
		},
		{
			name: "Service Principal with client secret",
			options: &Options{
				LoginMethod:  ServicePrincipalLogin,
				ServerID:     "server-id",
				TenantID:     "tenant-id",
				ClientID:     "client-id",
				ClientSecret: "secret",
			},
			wantErr: false,
		},
		{
			name: "Unsupported login method",
			options: &Options{
				LoginMethod: "unsupported",
				ServerID:    "server-id",
			},
			wantErr:    true,
			errMessage: "unsupported token provider",
		},
		{
			name: "Azure Pipelines login",
			options: &Options{
				LoginMethod:                       AzurePipelinesLogin,
				ServerID:                          "server-id",
				TenantID:                          "tenant-id", 
				ClientID:                          "client-id",
				AzurePipelinesServiceConnectionID: "service-connection-id",
			},
			wantErr: false,
		},
		{
			name: "Chained login",
			options: &Options{
				LoginMethod: ChainedLogin,
				ServerID:    "server-id",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := azidentity.AuthenticationRecord{}
			provider, err := NewAzIdentityCredential(record, tt.options)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Equal(t, tt.errMessage, err.Error())
				}
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestNewAzIdentityCredentialWithWorkloadIdentity(t *testing.T) {
	// Setup environment variables
	os.Setenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN", "token")
	os.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "url")
	defer func() {
		os.Unsetenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")
		os.Unsetenv("ACTIONS_ID_TOKEN_REQUEST_URL")
	}()

	tests := []struct {
		name    string
		options *Options
		wantErr bool
	}{
		{
			name: "GitHub Actions Workload Identity",
			options: &Options{
				LoginMethod: WorkloadIdentityLogin,
				ServerID:    "server-id",
				TenantID:    "tenant-id",
				ClientID:    "client-id",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := azidentity.AuthenticationRecord{}
			provider, err := NewAzIdentityCredential(record, tt.options)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

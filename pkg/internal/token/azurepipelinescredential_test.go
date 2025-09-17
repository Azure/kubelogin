package token

import (
	"fmt"
	"os"
	"testing"

	"github.com/Azure/kubelogin/pkg/internal/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAzurePipelinesCredential(t *testing.T) {
	// Clean up environment variables after test
	defer func() {
		os.Unsetenv(env.SystemAccessToken)
		os.Unsetenv(env.SystemOIDCRequestURI)
	}()

	tests := []struct {
		name                 string
		opts                 *Options
		systemAccessToken    string
		systemOIDCRequestURI string
		expectError          bool
		expectErrorSubstring string
	}{
		{
			name: "valid credentials",
			opts: &Options{
				TenantID:                          "test-tenant-id",
				ClientID:                          "test-client-id",
				AzurePipelinesServiceConnectionID: "test-service-connection-id",
			},
			systemAccessToken:    "test-system-access-token",
			systemOIDCRequestURI: "https://test.oidc.request.uri",
			expectError:          false,
		},
		{
			name: "missing system access token",
			opts: &Options{
				TenantID:                          "test-tenant-id",
				ClientID:                          "test-client-id",
				AzurePipelinesServiceConnectionID: "test-service-connection-id",
			},
			systemAccessToken:    "",
			systemOIDCRequestURI: "https://test.oidc.request.uri",
			expectError:          true,
			expectErrorSubstring: fmt.Sprintf("%s environment variable not set", env.SystemAccessToken),
		},
		{
			name: "missing tenant ID",
			opts: &Options{
				ClientID:                          "test-client-id",
				AzurePipelinesServiceConnectionID: "test-service-connection-id",
			},
			systemAccessToken:    "test-system-access-token",
			systemOIDCRequestURI: "https://test.oidc.request.uri",
			expectError:          true,
			expectErrorSubstring: "failed to create azure pipelines credential",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.systemAccessToken != "" {
				os.Setenv(env.SystemAccessToken, test.systemAccessToken)
			} else {
				os.Unsetenv(env.SystemAccessToken)
			}

			if test.systemOIDCRequestURI != "" {
				os.Setenv(env.SystemOIDCRequestURI, test.systemOIDCRequestURI)
			} else {
				os.Unsetenv(env.SystemOIDCRequestURI)
			}

			cred, err := newAzurePipelinesCredential(test.opts)

			if test.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectErrorSubstring)
				assert.Nil(t, cred)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, cred)
				assert.Equal(t, "AzurePipelinesCredential", cred.Name())
				assert.False(t, cred.NeedAuthenticate())
			}
		})
	}
}
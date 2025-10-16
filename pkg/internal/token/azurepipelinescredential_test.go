package token

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
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

func Test_newAzurePipelinesCredential(t *testing.T) {
	type args struct {
		opts *Options
	}
	tests := []struct {
		name    string
		args    args
		want    CredentialProvider
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newAzurePipelinesCredential(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("newAzurePipelinesCredential() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newAzurePipelinesCredential() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzurePipelinesCredential_Name(t *testing.T) {
	tests := []struct {
		name string
		c    *AzurePipelinesCredential
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Name(); got != tt.want {
				t.Errorf("AzurePipelinesCredential.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzurePipelinesCredential_Authenticate(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts *policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *AzurePipelinesCredential
		args    args
		want    azidentity.AuthenticationRecord
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Authenticate(tt.args.ctx, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("AzurePipelinesCredential.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AzurePipelinesCredential.Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzurePipelinesCredential_GetToken(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *AzurePipelinesCredential
		args    args
		want    azcore.AccessToken
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.GetToken(tt.args.ctx, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("AzurePipelinesCredential.GetToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AzurePipelinesCredential.GetToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzurePipelinesCredential_NeedAuthenticate(t *testing.T) {
	tests := []struct {
		name string
		c    *AzurePipelinesCredential
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.NeedAuthenticate(); got != tt.want {
				t.Errorf("AzurePipelinesCredential.NeedAuthenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClientSecretCredentialWithPoP(t *testing.T) {
	testCases := []struct {
		name           string
		opts           *Options
		expectErrorMsg string
		expectName     string
	}{
		{
			name: "valid options",
			opts: &Options{
				ClientID:          "test-client-id",
				TenantID:          "test-tenant-id",
				ClientSecret:      "test-secret",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=test-cluster",
				AuthorityHost:     "https://login.microsoftonline.com/",
			},
			expectName: "ClientSecretCredentialWithPoP",
		},
		{
			name: "missing client ID",
			opts: &Options{
				TenantID:          "test-tenant-id",
				ClientSecret:      "test-secret",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=test-cluster",
				AuthorityHost:     "https://login.microsoftonline.com/",
			},
			expectErrorMsg: "client ID cannot be empty",
		},
		{
			name: "missing tenant ID",
			opts: &Options{
				ClientID:          "test-client-id",
				ClientSecret:      "test-secret",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=test-cluster",
				AuthorityHost:     "https://login.microsoftonline.com/",
			},
			expectErrorMsg: "tenant ID cannot be empty",
		},
		{
			name: "missing client secret",
			opts: &Options{
				ClientID:          "test-client-id",
				TenantID:          "test-tenant-id",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=test-cluster",
				AuthorityHost:     "https://login.microsoftonline.com/",
			},
			expectErrorMsg: "client secret cannot be empty",
		},
		{
			name: "missing PoP claims",
			opts: &Options{
				ClientID:          "test-client-id",
				TenantID:          "test-tenant-id",
				ClientSecret:      "test-secret",
				IsPoPTokenEnabled: true,
				AuthorityHost:     "https://login.microsoftonline.com/",
			},
			expectErrorMsg: "unable to parse PoP claims: failed to parse PoP token claims: no claims provided",
		},
		{
			name: "invalid PoP claims format",
			opts: &Options{
				ClientID:          "test-client-id",
				TenantID:          "test-tenant-id",
				ClientSecret:      "test-secret",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "invalid-format",
				AuthorityHost:     "https://login.microsoftonline.com/",
			},
			expectErrorMsg: "unable to parse PoP claims: failed to parse PoP token claims. Ensure the claims are formatted as `key=value` with no extra whitespace",
		},
		{
			name: "missing required u-claim",
			opts: &Options{
				ClientID:          "test-client-id",
				TenantID:          "test-tenant-id",
				ClientSecret:      "test-secret",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "key=value",
				AuthorityHost:     "https://login.microsoftonline.com/",
			},
			expectErrorMsg: "unable to parse PoP claims: required u-claim not provided for PoP token flow. Please provide the ARM ID of the cluster in the format `u=<ARM_ID>`",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cred, err := newClientSecretCredentialWithPoP(tc.opts)
			if tc.expectErrorMsg != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.expectErrorMsg, err.Error())
				assert.Nil(t, cred)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cred)
				assert.Equal(t, tc.expectName, cred.Name())
			}
		})
	}
}

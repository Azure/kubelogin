package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUsernamePasswordCredentialWithPoP(t *testing.T) {
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
				Username:          "test-user",
				Password:          "test-password",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=test-cluster",
				AuthorityHost:     "https://login.microsoftonline.com/",
			},
			expectName: "UsernamePasswordCredentialWithPoP",
		},
		{
			name: "missing client ID",
			opts: &Options{
				TenantID:          "test-tenant-id",
				Username:          "test-user",
				Password:          "test-password",
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
				Username:          "test-user",
				Password:          "test-password",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=test-cluster",
				AuthorityHost:     "https://login.microsoftonline.com/",
			},
			expectErrorMsg: "tenant ID cannot be empty",
		},
		{
			name: "missing username",
			opts: &Options{
				ClientID:          "test-client-id",
				TenantID:          "test-tenant-id",
				Password:          "test-password",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=test-cluster",
				AuthorityHost:     "https://login.microsoftonline.com/",
			},
			expectErrorMsg: "username cannot be empty",
		},
		{
			name: "missing password",
			opts: &Options{
				ClientID:          "test-client-id",
				TenantID:          "test-tenant-id",
				Username:          "test-user",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=test-cluster",
				AuthorityHost:     "https://login.microsoftonline.com/",
			},
			expectErrorMsg: "password cannot be empty",
		},
		{
			name: "missing PoP claims",
			opts: &Options{
				ClientID:          "test-client-id",
				TenantID:          "test-tenant-id",
				Username:          "test-user",
				Password:          "test-password",
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
				Username:          "test-user",
				Password:          "test-password",
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
				Username:          "test-user",
				Password:          "test-password",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "key=value",
				AuthorityHost:     "https://login.microsoftonline.com/",
			},
			expectErrorMsg: "unable to parse PoP claims: required u-claim not provided for PoP token flow. Please provide the ARM ID of the cluster in the format `u=<ARM_ID>`",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cred, err := newUsernamePasswordCredentialWithPoP(tc.opts, nil)
			if tc.expectErrorMsg != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.expectErrorMsg, err.Error())
				assert.Nil(t, cred)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cred)
				assert.Equal(t, tc.expectName, cred.Name())
				assert.False(t, cred.NeedAuthenticate())
			}
		})
	}
}

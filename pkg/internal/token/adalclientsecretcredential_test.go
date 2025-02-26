package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewADALClientSecretCredential(t *testing.T) {
	testCases := []struct {
		name           string
		opts           *Options
		expectErrorMsg string
		expectName     string
	}{
		{
			name: "valid options",
			opts: &Options{
				ClientID:     "test-client-id",
				TenantID:     "test-tenant-id",
				ClientSecret: "test-client-secret",
				IsLegacy:     true,
			},
			expectName: "ADALClientSecretCredential",
		},
		{
			name: "missing client ID",
			opts: &Options{
				TenantID:     "test-tenant-id",
				ClientSecret: "test-client-secret",
				IsLegacy:     true,
			},
			expectErrorMsg: "client ID cannot be empty",
		},
		{
			name: "missing tenant ID",
			opts: &Options{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				IsLegacy:     true,
			},
			expectErrorMsg: "tenant ID cannot be empty",
		},
		{
			name: "missing client secret",
			opts: &Options{
				ClientID: "test-client-id",
				TenantID: "test-tenant-id",
				IsLegacy: true,
			},
			expectErrorMsg: "client secret cannot be empty",
		},
		{
			name: "non-legacy mode",
			opts: &Options{
				ClientID:     "test-client-id",
				TenantID:     "test-tenant-id",
				ClientSecret: "test-client-secret",
				IsLegacy:     false,
			},
			expectErrorMsg: "ADALClientSecretCredential is not supported in non-legacy mode",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cred, err := newADALClientSecretCredential(tc.opts)
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

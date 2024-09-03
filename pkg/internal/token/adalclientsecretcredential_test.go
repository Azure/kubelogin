package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewADALClientSecretCredential(t *testing.T) {
	testCases := []struct {
		name           string
		opts           *Options
		expectError    bool
		expectErrorMsg string
		expectNil      bool
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
			expectError: false,
			expectNil:   false,
			expectName:  "ADALClientSecretCredential",
		},
		{
			name: "missing client ID",
			opts: &Options{
				TenantID:     "test-tenant-id",
				ClientSecret: "test-client-secret",
				IsLegacy:     true,
			},
			expectError:    true,
			expectErrorMsg: "client ID cannot be empty",
			expectNil:      true,
		},
		{
			name: "missing tenant ID",
			opts: &Options{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				IsLegacy:     true,
			},
			expectError:    true,
			expectErrorMsg: "tenant ID cannot be empty",
			expectNil:      true,
		},
		{
			name: "missing client secret",
			opts: &Options{
				ClientID: "test-client-id",
				TenantID: "test-tenant-id",
				IsLegacy: true,
			},
			expectError:    true,
			expectErrorMsg: "client secret cannot be empty",
			expectNil:      true,
		},
		{
			name: "non-legacy mode",
			opts: &Options{
				ClientID:     "test-client-id",
				TenantID:     "test-tenant-id",
				ClientSecret: "test-client-secret",
				IsLegacy:     false,
			},
			expectError:    true,
			expectErrorMsg: "ADALClientSecretCredential is not supported in non-legacy mode",
			expectNil:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cred, err := newADALClientSecretCredential(tc.opts)
			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, tc.expectErrorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
			if tc.expectNil {
				assert.Nil(t, cred)
			} else {
				assert.NotNil(t, cred)
				assert.Equal(t, tc.expectName, cred.Name())
			}
		})
	}
}

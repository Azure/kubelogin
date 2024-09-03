package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewADALClientCertCredential(t *testing.T) {
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
				ClientID:           "test-client-id",
				TenantID:           "test-tenant-id",
				ClientCert:         "test-cert-path",
				ClientCertPassword: "test-cert-password",
				IsLegacy:           true,
			},
			expectError: false,
			expectNil:   false,
			expectName:  "ADALClientCertCredential",
		},
		{
			name: "missing client ID",
			opts: &Options{
				TenantID:           "test-tenant-id",
				ClientCert:         "test-cert-path",
				ClientCertPassword: "test-cert-password",
				IsLegacy:           true,
			},
			expectError:    true,
			expectErrorMsg: "client ID cannot be empty",
			expectNil:      true,
		},
		{
			name: "missing tenant ID",
			opts: &Options{
				ClientID:           "test-client-id",
				ClientCert:         "test-cert-path",
				ClientCertPassword: "test-cert-password",
				IsLegacy:           true,
			},
			expectError:    true,
			expectErrorMsg: "tenant ID cannot be empty",
			expectNil:      true,
		},
		{
			name: "missing client certificate",
			opts: &Options{
				ClientID:           "test-client-id",
				TenantID:           "test-tenant-id",
				ClientCertPassword: "test-cert-password",
				IsLegacy:           true,
			},
			expectError:    true,
			expectErrorMsg: "client certificate cannot be empty",
			expectNil:      true,
		},
		{
			name: "non-legacy mode",
			opts: &Options{
				ClientID:           "test-client-id",
				TenantID:           "test-tenant-id",
				ClientCert:         "test-cert-path",
				ClientCertPassword: "test-cert-password",
				IsLegacy:           false,
			},
			expectError:    true,
			expectErrorMsg: "ADALClientCertCredential is not supported in non-legacy mode",
			expectNil:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cred, err := newADALClientCertCredential(tc.opts)
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

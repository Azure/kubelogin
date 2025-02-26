package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewADALClientCertCredential(t *testing.T) {
	testCases := []struct {
		name           string
		opts           *Options
		expectErrorMsg string
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
			expectName: "ADALClientCertCredential",
		},
		{
			name: "missing client ID",
			opts: &Options{
				TenantID:           "test-tenant-id",
				ClientCert:         "test-cert-path",
				ClientCertPassword: "test-cert-password",
				IsLegacy:           true,
			},
			expectErrorMsg: "client ID cannot be empty",
		},
		{
			name: "missing tenant ID",
			opts: &Options{
				ClientID:           "test-client-id",
				ClientCert:         "test-cert-path",
				ClientCertPassword: "test-cert-password",
				IsLegacy:           true,
			},
			expectErrorMsg: "tenant ID cannot be empty",
		},
		{
			name: "missing client certificate",
			opts: &Options{
				ClientID:           "test-client-id",
				TenantID:           "test-tenant-id",
				ClientCertPassword: "test-cert-password",
				IsLegacy:           true,
			},
			expectErrorMsg: "client certificate cannot be empty",
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
			expectErrorMsg: "ADALClientCertCredential is not supported in non-legacy mode",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cred, err := newADALClientCertCredential(tc.opts)
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

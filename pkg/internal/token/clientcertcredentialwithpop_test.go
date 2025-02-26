package token

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClientCertificateCredentialWithPoP(t *testing.T) {
	certFile := os.Getenv("KUBELOGIN_LIVETEST_CERTIFICATE_FILE")
	if certFile == "" {
		certFile = "fixtures/cert.pem"
	}

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
				ClientCert:        certFile,
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=test-cluster",
			},
			expectName: "ClientCertificateCredentialWithPoP",
		},
		{
			name: "missing client ID",
			opts: &Options{
				TenantID:          "test-tenant-id",
				ClientCert:        certFile,
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=test-cluster",
			},
			expectErrorMsg: "client ID cannot be empty",
		},
		{
			name: "missing tenant ID",
			opts: &Options{
				ClientID:          "test-client-id",
				ClientCert:        certFile,
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=test-cluster",
			},
			expectErrorMsg: "tenant ID cannot be empty",
		},
		{
			name: "missing client certificate",
			opts: &Options{
				ClientID:          "test-client-id",
				TenantID:          "test-tenant-id",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=test-cluster",
			},
			expectErrorMsg: "client certificate cannot be empty",
		},
		{
			name: "missing PoP claims",
			opts: &Options{
				ClientID:          "test-client-id",
				TenantID:          "test-tenant-id",
				ClientCert:        certFile,
				IsPoPTokenEnabled: true,
			},
			expectErrorMsg: "unable to parse PoP claims: failed to parse PoP token claims: no claims provided",
		},
		{
			name: "invalid PoP claims format",
			opts: &Options{
				ClientID:          "test-client-id",
				TenantID:          "test-tenant-id",
				ClientCert:        certFile,
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "invalid-format",
			},
			expectErrorMsg: "unable to parse PoP claims: failed to parse PoP token claims. Ensure the claims are formatted as `key=value` with no extra whitespace",
		},
		{
			name: "missing required u-claim",
			opts: &Options{
				ClientID:          "test-client-id",
				TenantID:          "test-tenant-id",
				ClientCert:        certFile,
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "key=value",
			},
			expectErrorMsg: "unable to parse PoP claims: required u-claim not provided for PoP token flow. Please provide the ARM ID of the cluster in the format `u=<ARM_ID>`",
		},
		{
			name: "invalid certificate file",
			opts: &Options{
				ClientID:          "test-client-id",
				TenantID:          "test-tenant-id",
				ClientCert:        "nonexistent.pem",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=test-cluster",
			},
			expectErrorMsg: "failed to read certificate: failed to read the certificate file (nonexistent.pem):",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cred, err := newClientCertificateCredentialWithPoP(tc.opts)

			if tc.expectErrorMsg != "" {
				assert.Error(t, err)
				if tc.expectErrorMsg != "" {
					assert.Contains(t, err.Error(), tc.expectErrorMsg)
				}
				assert.Nil(t, cred)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cred)
				assert.Equal(t, tc.expectName, cred.Name())
			}
		})
	}
}

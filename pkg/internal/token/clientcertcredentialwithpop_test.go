package token

import (
	"context"
	"os"
	"testing"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/stretchr/testify/assert"

	popcache "github.com/Azure/kubelogin/pkg/internal/pop/cache"
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
				AuthorityHost:     "https://login.microsoftonline.com/",
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
				AuthorityHost:     "https://login.microsoftonline.com/",
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
				AuthorityHost:     "https://login.microsoftonline.com/",
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
				AuthorityHost:     "https://login.microsoftonline.com/",
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
				AuthorityHost:     "https://login.microsoftonline.com/",
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
				AuthorityHost:     "https://login.microsoftonline.com/",
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
				AuthorityHost:     "https://login.microsoftonline.com/",
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
				AuthorityHost:     "https://login.microsoftonline.com/",
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

func TestNewClientCertificateCredentialWithPoP_CacheScenarios(t *testing.T) {
	certFile := os.Getenv("KUBELOGIN_LIVETEST_CERTIFICATE_FILE")
	if certFile == "" {
		certFile = "fixtures/cert.pem"
	}

	validOpts := &Options{
		ClientID:           "test-client-id",
		TenantID:           "test-tenant-id",
		ClientCert:         certFile,
		IsPoPTokenEnabled:  true,
		PoPTokenClaims:     "u=test-cluster",
		AuthRecordCacheDir: "/tmp/test-cache",
		AuthorityHost:      "https://login.microsoftonline.com/",
	}

	testCases := []struct {
		name                    string
		cacheProvided           bool
		expectUsePersistentKeys bool
		expectCacheDir          string
		description             string
	}{
		{
			name:                    "with cache - should use persistent keys",
			cacheProvided:           true,
			expectUsePersistentKeys: true,
			expectCacheDir:          "/tmp/test-cache",
			description:             "When cache is available, should use persistent key storage",
		},
		{
			name:                    "nil cache - should use ephemeral keys",
			cacheProvided:           false,
			expectUsePersistentKeys: false,
			expectCacheDir:          "",
			description:             "When cache is nil, should use ephemeral keys",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up the cache in Options
			testOpts := *validOpts // Copy the options

			if tc.cacheProvided {
				// Try to create a real cache for testing, fallback to nil on error
				popCache, _ := popcache.NewCache("/tmp/test-cache")
				testOpts.setPoPTokenCache(popCache)
			} else {
				testOpts.setPoPTokenCache(nil)
			}

			cred, err := newClientCertificateCredentialWithPoP(&testOpts)

			assert.NoError(t, err, tc.description)
			assert.NotNil(t, cred, tc.description)

			// Check internal state via type assertion
			if certCred, ok := cred.(*ClientCertificateCredentialWithPoP); ok {
				assert.Equal(t, tc.expectUsePersistentKeys, certCred.usePersistentKeys, tc.description)
				assert.Equal(t, tc.expectCacheDir, certCred.cacheDir, tc.description)
			}
		})
	}
}

// mockCertCacheExportReplace is a simple mock implementation for testing
type mockCertCacheExportReplace struct{}

func (m *mockCertCacheExportReplace) Export(ctx context.Context, marshaler cache.Marshaler, hints cache.ExportHints) error {
	return nil
}

func (m *mockCertCacheExportReplace) Replace(ctx context.Context, unmarshaler cache.Unmarshaler, hints cache.ReplaceHints) error {
	return nil
}

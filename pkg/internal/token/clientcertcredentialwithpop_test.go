package token

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	popcache "github.com/Azure/kubelogin/pkg/internal/pop/cache"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
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
		cacheProvided  bool
		expectCacheDir string
		description    string
	}{
		{
			name:           "with cache - should use persistent keys",
			cacheProvided:  true,
			expectCacheDir: "/tmp/test-cache",
			description:    "When cache is available, should use persistent key storage",
		},
		{
			name:           "nil cache - should use ephemeral keys",
			cacheProvided:  false,
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

func Test_newClientCertificateCredentialWithPoP(t *testing.T) {
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
			got, err := newClientCertificateCredentialWithPoP(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("newClientCertificateCredentialWithPoP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newClientCertificateCredentialWithPoP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientCertificateCredentialWithPoP_Name(t *testing.T) {
	tests := []struct {
		name string
		c    *ClientCertificateCredentialWithPoP
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Name(); got != tt.want {
				t.Errorf("ClientCertificateCredentialWithPoP.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientCertificateCredentialWithPoP_Authenticate(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts *policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *ClientCertificateCredentialWithPoP
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
				t.Errorf("ClientCertificateCredentialWithPoP.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientCertificateCredentialWithPoP.Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientCertificateCredentialWithPoP_GetToken(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *ClientCertificateCredentialWithPoP
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
				t.Errorf("ClientCertificateCredentialWithPoP.GetToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientCertificateCredentialWithPoP.GetToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientCertificateCredentialWithPoP_NeedAuthenticate(t *testing.T) {
	tests := []struct {
		name string
		c    *ClientCertificateCredentialWithPoP
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.NeedAuthenticate(); got != tt.want {
				t.Errorf("ClientCertificateCredentialWithPoP.NeedAuthenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

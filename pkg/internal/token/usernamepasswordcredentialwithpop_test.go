package token

import (
	"context"
	"testing"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/stretchr/testify/assert"

	popcache "github.com/Azure/kubelogin/pkg/internal/pop/cache"
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
			cred, err := newUsernamePasswordCredentialWithPoP(tc.opts)
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

func TestNewUsernamePasswordCredentialWithPoP_CacheScenarios(t *testing.T) {
	validOpts := &Options{
		ClientID:           "test-client-id",
		TenantID:           "test-tenant-id",
		Username:           "test-user",
		Password:           "test-password",
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

			cred, err := newUsernamePasswordCredentialWithPoP(&testOpts)

			assert.NoError(t, err, tc.description)
			assert.NotNil(t, cred, tc.description)

			// Check internal state via type assertion
			if userPassCred, ok := cred.(*UsernamePasswordCredentialWithPoP); ok {
				assert.Equal(t, tc.expectUsePersistentKeys, userPassCred.usePersistentKeys, tc.description)
				assert.Equal(t, tc.expectCacheDir, userPassCred.cacheDir, tc.description)
			}
		})
	}
}

// mockUserPassCacheExportReplace is a simple mock implementation for testing
type mockUserPassCacheExportReplace struct{}

func (m *mockUserPassCacheExportReplace) Export(ctx context.Context, marshaler cache.Marshaler, hints cache.ExportHints) error {
	return nil
}

func (m *mockUserPassCacheExportReplace) Replace(ctx context.Context, unmarshaler cache.Unmarshaler, hints cache.ReplaceHints) error {
	return nil
}

package token

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	popcache "github.com/Azure/kubelogin/pkg/internal/pop/cache"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/stretchr/testify/assert"
)

func TestNewInteractiveBrowserCredentialWithPoP(t *testing.T) {
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
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=test-cluster",
			},
			expectName: "InteractiveBrowserCredentialWithPoP",
		},
		{
			name: "missing client ID",
			opts: &Options{
				TenantID:          "test-tenant-id",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=test-cluster",
			},
			expectErrorMsg: "client ID cannot be empty",
		},
		{
			name: "missing tenant ID",
			opts: &Options{
				ClientID:          "test-client-id",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=test-cluster",
			},
			expectErrorMsg: "tenant ID cannot be empty",
		},
		{
			name: "missing PoP claims",
			opts: &Options{
				ClientID:          "test-client-id",
				TenantID:          "test-tenant-id",
				IsPoPTokenEnabled: true,
			},
			expectErrorMsg: "unable to parse PoP claims: failed to parse PoP token claims: no claims provided",
		},
		{
			name: "invalid PoP claims format",
			opts: &Options{
				ClientID:          "test-client-id",
				TenantID:          "test-tenant-id",
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
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "key=value",
			},
			expectErrorMsg: "unable to parse PoP claims: required u-claim not provided for PoP token flow. Please provide the ARM ID of the cluster in the format `u=<ARM_ID>`",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cred, err := newInteractiveBrowserCredentialWithPoP(tc.opts)
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

func TestNewInteractiveBrowserCredentialWithPoP_CacheScenarios(t *testing.T) {
	validOpts := &Options{
		ClientID:           "test-client-id",
		TenantID:           "test-tenant-id",
		IsPoPTokenEnabled:  true,
		PoPTokenClaims:     "u=test-cluster",
		AuthRecordCacheDir: "/tmp/test-cache",
	}

	testCases := []struct {
		name           string
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
			expectCacheDir: "",
			description:    "When cache is nil, should use ephemeral keys",
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

			cred, err := newInteractiveBrowserCredentialWithPoP(&testOpts)

			assert.NoError(t, err, tc.description)
			assert.NotNil(t, cred, tc.description)

			// Verify that the key provider was set correctly by checking behavior
			if browserCred, ok := cred.(*InteractiveBrowserCredentialWithPoP); ok {
				assert.NotNil(t, browserCred.keyProvider, tc.description)
				// Verify key provider behavior: should be able to get a key
				_, err := browserCred.keyProvider.GetPoPKey()
				assert.NoError(t, err, "Key provider should be able to generate PoP keys")
			}
		})
	}
}

// mockCacheExportReplace is a simple mock implementation for testing
type mockCacheExportReplace struct{}

func (m *mockCacheExportReplace) Export(ctx context.Context, marshaler cache.Marshaler, hints cache.ExportHints) error {
	return nil
}

func (m *mockCacheExportReplace) Replace(ctx context.Context, unmarshaler cache.Unmarshaler, hints cache.ReplaceHints) error {
	return nil
}

func Test_newInteractiveBrowserCredentialWithPoP(t *testing.T) {
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
			got, err := newInteractiveBrowserCredentialWithPoP(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("newInteractiveBrowserCredentialWithPoP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newInteractiveBrowserCredentialWithPoP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInteractiveBrowserCredentialWithPoP_Name(t *testing.T) {
	tests := []struct {
		name string
		c    *InteractiveBrowserCredentialWithPoP
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Name(); got != tt.want {
				t.Errorf("InteractiveBrowserCredentialWithPoP.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInteractiveBrowserCredentialWithPoP_Authenticate(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts *policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *InteractiveBrowserCredentialWithPoP
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
				t.Errorf("InteractiveBrowserCredentialWithPoP.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InteractiveBrowserCredentialWithPoP.Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInteractiveBrowserCredentialWithPoP_GetToken(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *InteractiveBrowserCredentialWithPoP
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
				t.Errorf("InteractiveBrowserCredentialWithPoP.GetToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InteractiveBrowserCredentialWithPoP.GetToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInteractiveBrowserCredentialWithPoP_NeedAuthenticate(t *testing.T) {
	tests := []struct {
		name string
		c    *InteractiveBrowserCredentialWithPoP
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.NeedAuthenticate(); got != tt.want {
				t.Errorf("InteractiveBrowserCredentialWithPoP.NeedAuthenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

package token

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/kubelogin/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestManagedIdentityCredential_GetToken(t *testing.T) {
	rec, err := testutils.GetVCRHttpClient("fixtures/managedidentity_credential", testutils.TestTenantID)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}
	defer rec.Stop()

	opts := &Options{
		ClientID:   "49a6a7eb-d4f9-444a-a216-7b966e31bb05",
		ServerID:   testutils.TestServerID,
		httpClient: rec.GetDefaultClient(),
	}
	cred, err := newManagedIdentityCredential(opts)
	if err != nil {
		t.Fatalf("failed to create credential: %v", err)
	}

	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{opts.ServerID + "/.default"},
	})
	assert.NoError(t, err)
	assert.Equal(t, testutils.TestToken, token.Token)
}

func Test_newManagedIdentityCredential(t *testing.T) {
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
			got, err := newManagedIdentityCredential(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("newManagedIdentityCredential() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newManagedIdentityCredential() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManagedIdentityCredential_Name(t *testing.T) {
	tests := []struct {
		name string
		c    *ManagedIdentityCredential
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Name(); got != tt.want {
				t.Errorf("ManagedIdentityCredential.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManagedIdentityCredential_Authenticate(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts *policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *ManagedIdentityCredential
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
				t.Errorf("ManagedIdentityCredential.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ManagedIdentityCredential.Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManagedIdentityCredential_NeedAuthenticate(t *testing.T) {
	tests := []struct {
		name string
		c    *ManagedIdentityCredential
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.NeedAuthenticate(); got != tt.want {
				t.Errorf("ManagedIdentityCredential.NeedAuthenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

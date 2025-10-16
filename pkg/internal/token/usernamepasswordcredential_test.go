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

func TestUsernamePasswordCredential_GetToken(t *testing.T) {
	rec, err := testutils.GetVCRHttpClient("fixtures/usernamepassword_credential", testutils.TestTenantID)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}
	defer rec.Stop()

	opts := &Options{
		ClientID:   testutils.TestClientID,
		ServerID:   testutils.TestServerID,
		Username:   "user@example.come",
		Password:   "password",
		TenantID:   testutils.TestTenantID,
		httpClient: rec.GetDefaultClient(),
	}
	record := azidentity.AuthenticationRecord{}
	cred, err := newUsernamePasswordCredential(opts, record)
	if err != nil {
		t.Fatalf("failed to create credential: %v", err)
	}

	_, err = cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{opts.ServerID + "/.default"},
	})
	// our test environment requires MFA
	assert.ErrorContains(t, err, "AADSTS50076: Due to a configuration change made by your administrator, or because you moved to a new location, you must use multi-factor authentication to access")
}

func Test_newUsernamePasswordCredential(t *testing.T) {
	type args struct {
		opts   *Options
		record azidentity.AuthenticationRecord
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
			got, err := newUsernamePasswordCredential(tt.args.opts, tt.args.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("newUsernamePasswordCredential() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newUsernamePasswordCredential() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUsernamePasswordCredential_Name(t *testing.T) {
	tests := []struct {
		name string
		c    *UsernamePasswordCredential
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Name(); got != tt.want {
				t.Errorf("UsernamePasswordCredential.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUsernamePasswordCredential_Authenticate(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts *policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *UsernamePasswordCredential
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
				t.Errorf("UsernamePasswordCredential.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UsernamePasswordCredential.Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUsernamePasswordCredential_NeedAuthenticate(t *testing.T) {
	tests := []struct {
		name string
		c    *UsernamePasswordCredential
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.NeedAuthenticate(); got != tt.want {
				t.Errorf("UsernamePasswordCredential.NeedAuthenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

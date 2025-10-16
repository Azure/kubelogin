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

func TestClientSecretCredential_GetToken(t *testing.T) {
	rec, err := testutils.GetVCRHttpClient("fixtures/client_secret_credential", testutils.TestTenantID)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}
	defer rec.Stop()

	opts := &Options{
		ClientID:     testutils.TestClientID,
		ServerID:     testutils.TestServerID,
		ClientSecret: "password",
		TenantID:     testutils.TestTenantID,
		httpClient:   rec.GetDefaultClient(),
	}

	cred, err := newClientSecretCredential(opts)
	if err != nil {
		t.Fatalf("failed to create credential: %v", err)
	}

	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{opts.ServerID + "/.default"},
	})
	assert.NoError(t, err)
	assert.Equal(t, testutils.TestToken, token.Token)
}

func Test_newClientSecretCredential(t *testing.T) {
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
			got, err := newClientSecretCredential(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("newClientSecretCredential() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newClientSecretCredential() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientSecretCredential_Name(t *testing.T) {
	tests := []struct {
		name string
		c    *ClientSecretCredential
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Name(); got != tt.want {
				t.Errorf("ClientSecretCredential.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientSecretCredential_Authenticate(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts *policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *ClientSecretCredential
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
				t.Errorf("ClientSecretCredential.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientSecretCredential.Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientSecretCredential_NeedAuthenticate(t *testing.T) {
	tests := []struct {
		name string
		c    *ClientSecretCredential
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.NeedAuthenticate(); got != tt.want {
				t.Errorf("ClientSecretCredential.NeedAuthenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

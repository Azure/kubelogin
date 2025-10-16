package token

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/stretchr/testify/assert"
)

func TestInteractiveBrowserCredential_GetToken(t *testing.T) {
	if _, ok := os.LookupEnv("KUBELOGIN_MANUAL_TEST"); !ok {
		t.Skip("skipping test because KUBELOGIN_MANUAL_TEST is not set")
	}

	liveTestTenantID := os.Getenv("KUBELOGIN_LIVETEST_TENANT_ID")

	if liveTestTenantID == "" {
		t.Skip("skipping test because KUBELOGIN_LIVETEST_TENANT_ID is not set")
	}

	opts := &Options{
		ClientID: "80faf920-1908-4b52-b5ef-a8e7bedfc67a",
		ServerID: "6dae42f8-4368-4678-94ff-3960e28e3630",
		TenantID: liveTestTenantID,
	}
	record := azidentity.AuthenticationRecord{}
	cred, err := newInteractiveBrowserCredential(opts, record)
	if err != nil {
		t.Fatalf("failed to create credential: %v", err)
	}

	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{opts.ServerID + "/.default"},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, token.Token)
}

func Test_newInteractiveBrowserCredential(t *testing.T) {
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
			got, err := newInteractiveBrowserCredential(tt.args.opts, tt.args.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("newInteractiveBrowserCredential() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newInteractiveBrowserCredential() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInteractiveBrowserCredential_Name(t *testing.T) {
	tests := []struct {
		name string
		c    *InteractiveBrowserCredential
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Name(); got != tt.want {
				t.Errorf("InteractiveBrowserCredential.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInteractiveBrowserCredential_Authenticate(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts *policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *InteractiveBrowserCredential
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
				t.Errorf("InteractiveBrowserCredential.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InteractiveBrowserCredential.Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInteractiveBrowserCredential_NeedAuthenticate(t *testing.T) {
	tests := []struct {
		name string
		c    *InteractiveBrowserCredential
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.NeedAuthenticate(); got != tt.want {
				t.Errorf("InteractiveBrowserCredential.NeedAuthenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

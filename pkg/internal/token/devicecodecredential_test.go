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

func TestDeviceCodeCredential_GetToken(t *testing.T) {
	rec, err := testutils.GetVCRHttpClient("fixtures/device_code_credential", testutils.TestTenantID)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}
	defer rec.Stop()

	opts := &Options{
		ClientID:   testutils.TestClientID,
		ServerID:   testutils.TestServerID,
		TenantID:   testutils.TestTenantID,
		httpClient: rec.GetDefaultClient(),
	}

	record := azidentity.AuthenticationRecord{}
	cred, err := newDeviceCodeCredential(opts, record)
	if err != nil {
		t.Fatalf("failed to create credential: %v", err)
	}

	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{opts.ServerID + "/.default"},
	})
	assert.NoError(t, err)
	assert.Equal(t, testutils.TestToken, token.Token)
}

func Test_newDeviceCodeCredential(t *testing.T) {
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
			got, err := newDeviceCodeCredential(tt.args.opts, tt.args.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("newDeviceCodeCredential() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newDeviceCodeCredential() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeviceCodeCredential_Name(t *testing.T) {
	tests := []struct {
		name string
		c    *DeviceCodeCredential
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Name(); got != tt.want {
				t.Errorf("DeviceCodeCredential.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeviceCodeCredential_Authenticate(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts *policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *DeviceCodeCredential
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
				t.Errorf("DeviceCodeCredential.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeviceCodeCredential.Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeviceCodeCredential_NeedAuthenticate(t *testing.T) {
	tests := []struct {
		name string
		c    *DeviceCodeCredential
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.NeedAuthenticate(); got != tt.want {
				t.Errorf("DeviceCodeCredential.NeedAuthenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

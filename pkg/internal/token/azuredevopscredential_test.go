package token

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/stretchr/testify/assert"
)

func TestNewAzureDeveloperCLICredential(t *testing.T) {
	testCases := []struct {
		name           string
		opts           *Options
		expectErrorMsg string
		expectName     string
	}{
		{
			name: "valid options",
			opts: &Options{
				TenantID: "test-tenant-id",
			},
			expectName: "AzureDeveloperCLICredential",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cred, err := newAzureDeveloperCLICredential(tc.opts)
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

func Test_newAzureDeveloperCLICredential(t *testing.T) {
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
			got, err := newAzureDeveloperCLICredential(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("newAzureDeveloperCLICredential() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newAzureDeveloperCLICredential() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzureDeveloperCLICredential_Name(t *testing.T) {
	tests := []struct {
		name string
		c    *AzureDeveloperCLICredential
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Name(); got != tt.want {
				t.Errorf("AzureDeveloperCLICredential.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzureDeveloperCLICredential_Authenticate(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts *policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *AzureDeveloperCLICredential
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
				t.Errorf("AzureDeveloperCLICredential.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AzureDeveloperCLICredential.Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzureDeveloperCLICredential_GetToken(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *AzureDeveloperCLICredential
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
				t.Errorf("AzureDeveloperCLICredential.GetToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AzureDeveloperCLICredential.GetToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzureDeveloperCLICredential_NeedAuthenticate(t *testing.T) {
	tests := []struct {
		name string
		c    *AzureDeveloperCLICredential
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.NeedAuthenticate(); got != tt.want {
				t.Errorf("AzureDeveloperCLICredential.NeedAuthenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

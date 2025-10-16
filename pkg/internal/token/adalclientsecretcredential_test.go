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

func TestNewADALClientSecretCredential(t *testing.T) {
	testCases := []struct {
		name           string
		opts           *Options
		expectErrorMsg string
		expectName     string
	}{
		{
			name: "valid options",
			opts: &Options{
				ClientID:     "test-client-id",
				TenantID:     "test-tenant-id",
				ClientSecret: "test-client-secret",
				IsLegacy:     true,
			},
			expectName: "ADALClientSecretCredential",
		},
		{
			name: "missing client ID",
			opts: &Options{
				TenantID:     "test-tenant-id",
				ClientSecret: "test-client-secret",
				IsLegacy:     true,
			},
			expectErrorMsg: "client ID cannot be empty",
		},
		{
			name: "missing tenant ID",
			opts: &Options{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				IsLegacy:     true,
			},
			expectErrorMsg: "tenant ID cannot be empty",
		},
		{
			name: "missing client secret",
			opts: &Options{
				ClientID: "test-client-id",
				TenantID: "test-tenant-id",
				IsLegacy: true,
			},
			expectErrorMsg: "client secret cannot be empty",
		},
		{
			name: "non-legacy mode",
			opts: &Options{
				ClientID:     "test-client-id",
				TenantID:     "test-tenant-id",
				ClientSecret: "test-client-secret",
				IsLegacy:     false,
			},
			expectErrorMsg: "ADALClientSecretCredential is not supported in non-legacy mode",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cred, err := newADALClientSecretCredential(tc.opts)
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

func Test_newADALClientSecretCredential(t *testing.T) {
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
			got, err := newADALClientSecretCredential(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("newADALClientSecretCredential() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newADALClientSecretCredential() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestADALClientSecretCredential_Name(t *testing.T) {
	tests := []struct {
		name string
		c    *ADALClientSecretCredential
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Name(); got != tt.want {
				t.Errorf("ADALClientSecretCredential.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestADALClientSecretCredential_Authenticate(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts *policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *ADALClientSecretCredential
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
				t.Errorf("ADALClientSecretCredential.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ADALClientSecretCredential.Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestADALClientSecretCredential_GetToken(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *ADALClientSecretCredential
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
				t.Errorf("ADALClientSecretCredential.GetToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ADALClientSecretCredential.GetToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestADALClientSecretCredential_NeedAuthenticate(t *testing.T) {
	tests := []struct {
		name string
		c    *ADALClientSecretCredential
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.NeedAuthenticate(); got != tt.want {
				t.Errorf("ADALClientSecretCredential.NeedAuthenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

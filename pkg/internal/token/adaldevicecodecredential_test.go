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

func TestNewADALDeviceCodeCredential(t *testing.T) {
	testCases := []struct {
		name     string
		opts     *Options
		expected string
	}{
		{
			name: "valid options",
			opts: &Options{
				ClientID: "test-client-id",
				TenantID: "test-tenant-id",
				IsLegacy: true,
			},
			expected: "ADALDeviceCodeCredential",
		},
		{
			name: "missing client ID",
			opts: &Options{
				TenantID: "test-tenant-id",
				IsLegacy: true,
			},
			expected: "client ID cannot be empty",
		},
		{
			name: "missing tenant ID",
			opts: &Options{
				ClientID: "test-client-id",
				IsLegacy: true,
			},
			expected: "tenant ID cannot be empty",
		},
		{
			name: "non-legacy mode",
			opts: &Options{
				ClientID: "test-client-id",
				TenantID: "test-tenant-id",
				IsLegacy: false,
			},
			expected: "ADALDeviceCodeCredential is not supported in non-legacy mode",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cred, err := newADALDeviceCodeCredential(tc.opts)
			if err != nil {
				assert.EqualError(t, err, tc.expected)
				assert.Nil(t, cred)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cred)
				assert.Equal(t, tc.expected, cred.Name())
			}
		})
	}
}

func Test_newADALDeviceCodeCredential(t *testing.T) {
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
			got, err := newADALDeviceCodeCredential(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("newADALDeviceCodeCredential() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newADALDeviceCodeCredential() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestADALDeviceCodeCredential_Name(t *testing.T) {
	tests := []struct {
		name string
		c    *ADALDeviceCodeCredential
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Name(); got != tt.want {
				t.Errorf("ADALDeviceCodeCredential.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestADALDeviceCodeCredential_Authenticate(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts *policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *ADALDeviceCodeCredential
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
				t.Errorf("ADALDeviceCodeCredential.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ADALDeviceCodeCredential.Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestADALDeviceCodeCredential_GetToken(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *ADALDeviceCodeCredential
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
				t.Errorf("ADALDeviceCodeCredential.GetToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ADALDeviceCodeCredential.GetToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestADALDeviceCodeCredential_NeedAuthenticate(t *testing.T) {
	tests := []struct {
		name string
		c    *ADALDeviceCodeCredential
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.NeedAuthenticate(); got != tt.want {
				t.Errorf("ADALDeviceCodeCredential.NeedAuthenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

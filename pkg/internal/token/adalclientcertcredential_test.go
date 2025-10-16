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

func TestNewADALClientCertCredential(t *testing.T) {
	testCases := []struct {
		name           string
		opts           *Options
		expectErrorMsg string
		expectName     string
	}{
		{
			name: "valid options",
			opts: &Options{
				ClientID:           "test-client-id",
				TenantID:           "test-tenant-id",
				ClientCert:         "test-cert-path",
				ClientCertPassword: "test-cert-password",
				IsLegacy:           true,
			},
			expectName: "ADALClientCertCredential",
		},
		{
			name: "missing client ID",
			opts: &Options{
				TenantID:           "test-tenant-id",
				ClientCert:         "test-cert-path",
				ClientCertPassword: "test-cert-password",
				IsLegacy:           true,
			},
			expectErrorMsg: "client ID cannot be empty",
		},
		{
			name: "missing tenant ID",
			opts: &Options{
				ClientID:           "test-client-id",
				ClientCert:         "test-cert-path",
				ClientCertPassword: "test-cert-password",
				IsLegacy:           true,
			},
			expectErrorMsg: "tenant ID cannot be empty",
		},
		{
			name: "missing client certificate",
			opts: &Options{
				ClientID:           "test-client-id",
				TenantID:           "test-tenant-id",
				ClientCertPassword: "test-cert-password",
				IsLegacy:           true,
			},
			expectErrorMsg: "client certificate cannot be empty",
		},
		{
			name: "non-legacy mode",
			opts: &Options{
				ClientID:           "test-client-id",
				TenantID:           "test-tenant-id",
				ClientCert:         "test-cert-path",
				ClientCertPassword: "test-cert-password",
				IsLegacy:           false,
			},
			expectErrorMsg: "ADALClientCertCredential is not supported in non-legacy mode",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cred, err := newADALClientCertCredential(tc.opts)
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

func Test_newADALClientCertCredential(t *testing.T) {
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
			got, err := newADALClientCertCredential(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("newADALClientCertCredential() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newADALClientCertCredential() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestADALClientCertCredential_Name(t *testing.T) {
	tests := []struct {
		name string
		c    *ADALClientCertCredential
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Name(); got != tt.want {
				t.Errorf("ADALClientCertCredential.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestADALClientCertCredential_Authenticate(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts *policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *ADALClientCertCredential
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
				t.Errorf("ADALClientCertCredential.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ADALClientCertCredential.Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestADALClientCertCredential_GetToken(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *ADALClientCertCredential
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
				t.Errorf("ADALClientCertCredential.GetToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ADALClientCertCredential.GetToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestADALClientCertCredential_NeedAuthenticate(t *testing.T) {
	tests := []struct {
		name string
		c    *ADALClientCertCredential
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.NeedAuthenticate(); got != tt.want {
				t.Errorf("ADALClientCertCredential.NeedAuthenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

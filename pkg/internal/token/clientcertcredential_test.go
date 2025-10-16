package token

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"os"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/kubelogin/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestClientCertCredential_GetToken(t *testing.T) {
	certFile := os.Getenv("KUBELOGIN_LIVETEST_CERTIFICATE_FILE")
	if certFile == "" {
		certFile = "fixtures/cert.pem"
	}

	rec, err := testutils.GetVCRHttpClient("fixtures/client_cert_credential", testutils.TestTenantID)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}
	defer rec.Stop()

	opts := &Options{
		ClientID:   testutils.TestClientID,
		ServerID:   testutils.TestServerID,
		ClientCert: certFile,
		TenantID:   testutils.TestTenantID,
		httpClient: rec.GetDefaultClient(),
	}

	cred, err := newClientCertificateCredential(opts)
	if err != nil {
		t.Fatalf("failed to create credential: %v", err)
	}

	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{opts.ServerID + "/.default"},
	})
	assert.NoError(t, err)
	assert.Equal(t, testutils.TestToken, token.Token)
}

func Test_newClientCertificateCredential(t *testing.T) {
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
			got, err := newClientCertificateCredential(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("newClientCertificateCredential() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newClientCertificateCredential() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientCertificateCredential_Name(t *testing.T) {
	tests := []struct {
		name string
		c    *ClientCertificateCredential
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Name(); got != tt.want {
				t.Errorf("ClientCertificateCredential.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientCertificateCredential_Authenticate(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts *policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *ClientCertificateCredential
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
				t.Errorf("ClientCertificateCredential.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientCertificateCredential.Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientCertificateCredential_GetToken(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *ClientCertificateCredential
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
				t.Errorf("ClientCertificateCredential.GetToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientCertificateCredential.GetToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientCertificateCredential_NeedAuthenticate(t *testing.T) {
	tests := []struct {
		name string
		c    *ClientCertificateCredential
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.NeedAuthenticate(); got != tt.want {
				t.Errorf("ClientCertificateCredential.NeedAuthenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isPublicKeyEqual(t *testing.T) {
	type args struct {
		key1 *rsa.PublicKey
		key2 *rsa.PublicKey
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPublicKeyEqual(tt.args.key1, tt.args.key2); got != tt.want {
				t.Errorf("isPublicKeyEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_splitPEMBlock(t *testing.T) {
	type args struct {
		pemBlock []byte
	}
	tests := []struct {
		name        string
		args        args
		wantCertPEM []byte
		wantKeyPEM  []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCertPEM, gotKeyPEM := splitPEMBlock(tt.args.pemBlock)
			if !reflect.DeepEqual(gotCertPEM, tt.wantCertPEM) {
				t.Errorf("splitPEMBlock() gotCertPEM = %v, want %v", gotCertPEM, tt.wantCertPEM)
			}
			if !reflect.DeepEqual(gotKeyPEM, tt.wantKeyPEM) {
				t.Errorf("splitPEMBlock() gotKeyPEM = %v, want %v", gotKeyPEM, tt.wantKeyPEM)
			}
		})
	}
}

func Test_parseRsaPrivateKey(t *testing.T) {
	type args struct {
		privateKeyPEM []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *rsa.PrivateKey
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRsaPrivateKey(tt.args.privateKeyPEM)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRsaPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseRsaPrivateKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseKeyPairFromPEMBlock(t *testing.T) {
	type args struct {
		pemBlock []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *x509.Certificate
		want1   *rsa.PrivateKey
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := parseKeyPairFromPEMBlock(tt.args.pemBlock)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseKeyPairFromPEMBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseKeyPairFromPEMBlock() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("parseKeyPairFromPEMBlock() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_decodePkcs12(t *testing.T) {
	type args struct {
		pkcs     []byte
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    *x509.Certificate
		want1   *rsa.PrivateKey
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := decodePkcs12(tt.args.pkcs, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodePkcs12() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodePkcs12() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("decodePkcs12() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_readCertificate(t *testing.T) {
	type args struct {
		certFile string
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    *x509.Certificate
		want1   *rsa.PrivateKey
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := readCertificate(tt.args.certFile, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("readCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readCertificate() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("readCertificate() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

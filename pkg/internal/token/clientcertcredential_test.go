package token

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/kubelogin/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestClientCertCredential_GetToken(t *testing.T) {
	var certFile, tempDir string

	liveCertificate := os.Getenv("KUBELOGIN_LIVETEST_CERTIFICATE_FILE")
	if liveCertificate == "" {
		var err error
		tempDir, certFile, err = createSelfSignedCertificatePEM()
		if err != nil {
			t.Fatalf("failed to create certificate: %v", err)
		}
		defer os.RemoveAll(tempDir)
	} else {
		certFile = liveCertificate
	}

	rec, err := testutils.GetVCRHttpClient_v4("fixtures/client_cert_credential", testutils.TestTenantID)
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

func createSelfSignedCertificatePEM() (string, string, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Example Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", err
	}

	tempDir, err := os.MkdirTemp("", "kubelogin")
	if err != nil {
		return "", "", err
	}
	file := filepath.Join(tempDir, "cert.pem")
	outFile, err := os.Create(file)
	if err != nil {
		return "", "", err
	}
	defer outFile.Close()

	if err := pem.Encode(outFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return "", "", err
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", "", err
	}
	if err := pem.Encode(outFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return "", "", err
	}

	return tempDir, file, nil
}

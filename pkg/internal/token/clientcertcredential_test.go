package token

import (
	"context"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
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

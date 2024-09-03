package token

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/kubelogin/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestWorkloadIdentityCredential_GetToken(t *testing.T) {
	var tokenFile string

	liveTokenFile := os.Getenv("KUBELOGIN_LIVETEST_FEDERATED_TOKEN_FILE")
	if liveTokenFile == "" {
		tempDir, err := os.MkdirTemp("", "kubelogin")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}

		tokenFile = filepath.Join(tempDir, "token")
		outFile, err := os.Create(tokenFile)
		if err != nil {
			t.Fatalf("failed to create token file: %v", err)
		}
		_, err = outFile.WriteString("[REDACTED]")
		if err != nil {
			t.Fatalf("failed to write token file: %v", err)
		}
		outFile.Close()
	} else {
		tokenFile = liveTokenFile
	}

	rec, err := testutils.GetVCRHttpClient("fixtures/workloadidentity_credential", testutils.TestTenantID)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}
	defer rec.Stop()

	opts := &Options{
		ClientID:           testutils.TestClientID,
		ServerID:           testutils.TestServerID,
		TenantID:           testutils.TestTenantID,
		FederatedTokenFile: tokenFile,
		httpClient:         rec.GetDefaultClient(),
	}
	cred, err := newWorkloadIdentityCredential(opts)
	if err != nil {
		t.Fatalf("failed to create credential: %v", err)
	}

	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{opts.ServerID + "/.default"},
	})
	assert.NoError(t, err)
	assert.Equal(t, testutils.TestToken, token.Token)
}

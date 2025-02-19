package token

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/kubelogin/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestClientSecretCredential_GetToken(t *testing.T) {
	rec, err := testutils.GetVCRHttpClient_v4("fixtures/client_secret_credential", testutils.TestTenantID)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}
	defer rec.Stop()

	opts := &Options{
		ClientID:     testutils.TestClientID,
		ServerID:     testutils.TestServerID,
		ClientSecret: "password",
		TenantID:     testutils.TestTenantID,
		httpClient:   rec.GetDefaultClient(),
	}

	cred, err := newClientSecretCredential(opts)
	if err != nil {
		t.Fatalf("failed to create credential: %v", err)
	}

	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{opts.ServerID + "/.default"},
	})
	assert.NoError(t, err)
	assert.Equal(t, testutils.TestToken, token.Token)
}

package token

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/kubelogin/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestManagedIdentityCredential_GetToken(t *testing.T) {
	rec, err := testutils.GetVCRHttpClient("fixtures/managedidentity_credential", testutils.TestTenantID)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}
	defer rec.Stop()

	opts := &Options{
		ClientID:   "49a6a7eb-d4f9-444a-a216-7b966e31bb05",
		ServerID:   testutils.TestServerID,
		httpClient: rec.GetDefaultClient(),
	}
	cred, err := newManagedIdentityCredential(opts)
	if err != nil {
		t.Fatalf("failed to create credential: %v", err)
	}

	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{opts.ServerID + "/.default"},
	})
	assert.NoError(t, err)
	assert.Equal(t, testutils.TestToken, token.Token)
}

package token

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/kubelogin/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestUsernamePasswordCredential_GetToken(t *testing.T) {
	rec, err := testutils.GetVCRHttpClient("fixtures/usernamepassword_credential", testutils.TestTenantID)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}
	defer rec.Stop()

	opts := &Options{
		ClientID:   testutils.TestClientID,
		ServerID:   testutils.TestServerID,
		Username:   "user@example.come",
		Password:   "password",
		TenantID:   testutils.TestTenantID,
		httpClient: rec.GetDefaultClient(),
	}
	record := azidentity.AuthenticationRecord{}
	cred, err := newUsernamePasswordCredential(opts, record)
	if err != nil {
		t.Fatalf("failed to create credential: %v", err)
	}

	_, err = cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{opts.ServerID + "/.default"},
	})
	// our test environment requires MFA
	assert.ErrorContains(t, err, "AADSTS50076: Due to a configuration change made by your administrator, or because you moved to a new location, you must use multi-factor authentication to access")
}

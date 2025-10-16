package token

import (
	"context"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/stretchr/testify/assert"
)

func TestInteractiveBrowserCredential_GetToken(t *testing.T) {
	if _, ok := os.LookupEnv("KUBELOGIN_MANUAL_TEST"); !ok {
		t.Skip("skipping test because KUBELOGIN_MANUAL_TEST is not set")
	}

	liveTestTenantID := os.Getenv("KUBELOGIN_LIVETEST_TENANT_ID")

	if liveTestTenantID == "" {
		t.Skip("skipping test because KUBELOGIN_LIVETEST_TENANT_ID is not set")
	}

	opts := &Options{
		ClientID: "80faf920-1908-4b52-b5ef-a8e7bedfc67a",
		ServerID: "6dae42f8-4368-4678-94ff-3960e28e3630",
		TenantID: liveTestTenantID,
	}
	record := azidentity.AuthenticationRecord{}
	cred, err := newInteractiveBrowserCredential(opts, record)
	if err != nil {
		t.Fatalf("failed to create credential: %v", err)
	}

	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{opts.ServerID + "/.default"},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, token.Token)
}

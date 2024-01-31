package token

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/go-autorest/autorest/adal"
)

type AzureDeveloperCLIToken struct {
	resourceID string
	tenantID   string
	cred       *azidentity.AzureDeveloperCLICredential
	timeout    time.Duration
}

// newAzureDeveloperCLIToken returns a TokenProvider that will fetch a token for the user currently logged into the Azure Developer CLI.
// Required arguments the resourceID (which is used as the scope) and an optional tenantID.
func newAzureDeveloperCLIToken(resourceID string, tenantID string, timeout time.Duration) (TokenProvider, error) {
	if resourceID == "" {
		return nil, errors.New("resourceID cannot be empty")
	}

	if timeout <= 0 {
		timeout = defaultTimeout
	}

	// Request a new Azure Developer CLI token provider
	cred, err := azidentity.NewAzureDeveloperCLICredential(&azidentity.AzureDeveloperCLICredentialOptions{
		TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create credential. Received: %v", err)
	}

	return &AzureDeveloperCLIToken{
		resourceID: resourceID,
		tenantID:   tenantID,
		cred:       cred,
		timeout:    timeout,
	}, nil
}

// Token fetches an azcore.AccessToken from the Azure Developer CLI SDK and converts it to an adal.Token for use with kubelogin.
func (p *AzureDeveloperCLIToken) Token(ctx context.Context) (adal.Token, error) {
	emptyToken := adal.Token{}

	if p.cred == nil {
		return emptyToken, errors.New("credential is nil. Create new instance with newAzureDeveloperCLIToken function")
	}

	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	policyOptions := policy.TokenRequestOptions{
		TenantID: p.tenantID,
		Scopes:   []string{fmt.Sprintf("%s/.default", p.resourceID)},
	}

	// Use the token provider to get a new token with the new context
	azdAccessToken, err := p.cred.GetToken(ctx, policyOptions)
	if err != nil {
		return emptyToken, fmt.Errorf("expected an empty error but received: %v", err)
	}

	if azdAccessToken.Token == "" {
		return emptyToken, errors.New("did not receive a token")
	}

	// azurecore.AccessTokens have ExpiresOn as Time.Time. We need to convert it to JSON.Number
	// by fetching the time in seconds since the Unix epoch via Unix() and then converting to a
	// JSON.Number via formatting as a string using a base-10 int64 conversion.
	expiresOn := json.Number(strconv.FormatInt(azdAccessToken.ExpiresOn.Unix(), 10))

	// Re-wrap the azurecore.AccessToken into an adal.Token
	return adal.Token{
		AccessToken: azdAccessToken.Token,
		ExpiresOn:   expiresOn,
		Resource:    p.resourceID,
	}, nil
}

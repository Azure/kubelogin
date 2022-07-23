package token

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/go-autorest/autorest/adal"
)

type AzureCLIToken struct {
	resourceID  string
	oAuthConfig adal.OAuthConfig
}

// newAzureCLIToken returns a TokenProvider that will fetch a token for the user currently logged into the Azure CLI.
// Required arguments include an oAuthConfiguration object and the resourceID (which is used as the scope)
func newAzureCLIToken(oAuthConfig adal.OAuthConfig, resourceID string) (TokenProvider, error) {
	if resourceID == "" {
		return nil, errors.New("resourceID cannot be empty")
	}

	return &AzureCLIToken{
		resourceID:  resourceID,
		oAuthConfig: oAuthConfig,
	}, nil
}

// Token fetches an azcore.AccessToken from the Azure CLI SDK and converts it to an adal.Token for use with kubelogin.
func (p *AzureCLIToken) Token() (adal.Token, error) {
	emptyToken := adal.Token{}

	// Request a new Azure CLI token provider
	cred, err := azidentity.NewAzureCLICredential(&azidentity.AzureCLICredentialOptions{})
	if err != nil {
		return emptyToken, fmt.Errorf("unable to create credential. Received: %v", err)
	}

	// Use the token provider to get a new token
	cliAccessToken, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{Scopes: []string{p.resourceID}})
	if err != nil {
		return emptyToken, fmt.Errorf("expected an empty error but received: %v", err)
	}
	if len(cliAccessToken.Token) == 0 {
		return emptyToken, errors.New("did not receive a token")
	}

	// azurecore.AccessTokens have ExpiresOn as Time.Time. We need to convert it to JSON.Number
	// by fetching the time in seconds since the Unix epoch via Unix() and then converting to a
	// JSON.Number via formatting as a string using a base-10 int64 conversion.
	expiresOn := json.Number(strconv.FormatInt(cliAccessToken.ExpiresOn.Unix(), 10))

	// Re-wrap the azurecore.AccessToken into an adal.Token
	return adal.Token{
		AccessToken: cliAccessToken.Token,
		ExpiresOn:   expiresOn,
		Resource:    p.resourceID,
	}, nil
}

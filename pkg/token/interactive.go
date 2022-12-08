package token

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/go-autorest/autorest/adal"
)

type InteractiveToken struct {
	clientID    string
	resourceID  string
	tenantID    string
	oAuthConfig adal.OAuthConfig
}

// newInteractiveTokenProvider returns a TokenProvider that will fetch a token for the user currently logged into the Interactive.
// Required arguments include an oAuthConfiguration object and the resourceID (which is used as the scope)
func newInteractiveTokenProvider(oAuthConfig adal.OAuthConfig, clientID, resourceID, tenantID string) (TokenProvider, error) {
	if clientID == "" {
		return nil, errors.New("clientID cannot be empty")
	}
	if resourceID == "" {
		return nil, errors.New("resourceID cannot be empty")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}

	return &InteractiveToken{
		clientID:    clientID,
		resourceID:  resourceID,
		tenantID:    tenantID,
		oAuthConfig: oAuthConfig,
	}, nil
}

// Token fetches an azcore.AccessToken from the interactive browser SDK and converts it to an adal.Token for use with kubelogin.
func (p *InteractiveToken) Token() (adal.Token, error) {
	emptyToken := adal.Token{}

	// Request a new Interactive token provider
	authorityFromConfig := p.oAuthConfig.AuthorityEndpoint
	clientOpts := azcore.ClientOptions{Cloud: cloud.Configuration{
		ActiveDirectoryAuthorityHost: authorityFromConfig.String(),
	}}
	cred, err := azidentity.NewInteractiveBrowserCredential(&azidentity.InteractiveBrowserCredentialOptions{
		ClientOptions: clientOpts,
		TenantID:      p.tenantID,
		ClientID:      p.clientID,
	})
	if err != nil {
		return emptyToken, fmt.Errorf("unable to create credential. Received: %w", err)
	}

	// Use the token provider to get a new token
	interactiveToken, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{Scopes: []string{p.resourceID + "/.default"}})
	if err != nil {
		return emptyToken, fmt.Errorf("expected an empty error but received: %w", err)
	}
	if interactiveToken.Token == "" {
		return emptyToken, errors.New("did not receive a token")
	}

	// azurecore.AccessTokens have ExpiresOn as Time.Time. We need to convert it to JSON.Number
	// by fetching the time in seconds since the Unix epoch via Unix() and then converting to a
	// JSON.Number via formatting as a string using a base-10 int64 conversion.
	expiresOn := json.Number(strconv.FormatInt(interactiveToken.ExpiresOn.Unix(), 10))

	// Re-wrap the azurecore.AccessToken into an adal.Token
	return adal.Token{
		AccessToken: interactiveToken.Token,
		ExpiresOn:   expiresOn,
		Resource:    p.resourceID,
	}, nil
}

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
	"github.com/Azure/kubelogin/pkg/internal/pop"
)

type InteractiveToken struct {
	clientID    string
	resourceID  string
	tenantID    string
	oAuthConfig adal.OAuthConfig
	popClaims   map[string]string
}

// newInteractiveTokenProvider returns a TokenProvider that will fetch a token for the user currently logged into the Interactive.
// Required arguments include an oAuthConfiguration object and the resourceID (which is used as the scope)
func newInteractiveTokenProvider(oAuthConfig adal.OAuthConfig, clientID, resourceID, tenantID string, popClaims map[string]string) (TokenProvider, error) {
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
		popClaims:   popClaims,
	}, nil
}

// Token fetches an azcore.AccessToken from the interactive browser SDK and converts it to an adal.Token for use with kubelogin.
func (p *InteractiveToken) Token(ctx context.Context) (adal.Token, error) {
	return p.TokenWithOptions(ctx, nil)
}

func (p *InteractiveToken) TokenWithOptions(ctx context.Context, options *azcore.ClientOptions) (adal.Token, error) {
	emptyToken := adal.Token{}

	// Request a new Interactive token provider
	authorityFromConfig := p.oAuthConfig.AuthorityEndpoint
	scopes := []string{p.resourceID + defaultScope}
	clientOpts := azcore.ClientOptions{Cloud: cloud.Configuration{
		ActiveDirectoryAuthorityHost: authorityFromConfig.String(),
	}}
	if options != nil {
		clientOpts = *options
	}
	var token string
	var expirationTimeUnix int64
	var err error

	if len(p.popClaims) > 0 {
		// If PoP token support is enabled and the correct u-claim is provided, use the MSAL
		// token provider to acquire a new token
		token, expirationTimeUnix, err = pop.AcquirePoPTokenInteractive(
			ctx,
			p.popClaims,
			scopes,
			authorityFromConfig.String(),
			p.clientID,
			&clientOpts,
		)
		if err != nil {
			return emptyToken, fmt.Errorf("failed to create PoP token using interactive login: %w", err)
		}
	} else {
		cred, err := azidentity.NewInteractiveBrowserCredential(&azidentity.InteractiveBrowserCredentialOptions{
			ClientOptions: clientOpts,
			TenantID:      p.tenantID,
			ClientID:      p.clientID,
		})
		if err != nil {
			return emptyToken, fmt.Errorf("unable to create credential. Received: %w", err)
		}

		// Use the token provider to get a new token
		interactiveToken, err := cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: scopes})
		if err != nil {
			return emptyToken, fmt.Errorf("expected an empty error but received: %w", err)
		}
		token = interactiveToken.Token
		if token == "" {
			return emptyToken, errors.New("did not receive a token")
		}
		expirationTimeUnix = interactiveToken.ExpiresOn.Unix()
	}

	// azurecore.AccessTokens have ExpiresOn as Time.Time. We need to convert it to JSON.Number
	// by fetching the time in seconds since the Unix epoch via Unix() and then converting to a
	// JSON.Number via formatting as a string using a base-10 int64 conversion
	expiresOn := json.Number(strconv.FormatInt(expirationTimeUnix, 10))

	// re-wrap the azurecore.AccessToken into an adal.Token
	return adal.Token{
		AccessToken: token,
		ExpiresOn:   expiresOn,
		Resource:    p.resourceID,
	}, nil
}

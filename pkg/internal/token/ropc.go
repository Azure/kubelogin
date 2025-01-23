package token

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/kubelogin/pkg/internal/pop"
)

type resourceOwnerToken struct {
	clientID                 string
	username                 string
	password                 string
	resourceID               string
	tenantID                 string
	oAuthConfig              adal.OAuthConfig
	popClaims                map[string]string
	disableInstanceDiscovery bool
}

func newResourceOwnerTokenProvider(
	oAuthConfig adal.OAuthConfig,
	clientID,
	username,
	password,
	resourceID,
	tenantID string,
	popClaims map[string]string,
	disableInstanceDiscovery bool,
) (TokenProvider, error) {
	if clientID == "" {
		return nil, errors.New("clientID cannot be empty")
	}
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	if password == "" {
		return nil, errors.New("password cannot be empty")
	}
	if resourceID == "" {
		return nil, errors.New("resourceID cannot be empty")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}

	return &resourceOwnerToken{
		clientID:                 clientID,
		username:                 username,
		password:                 password,
		resourceID:               resourceID,
		tenantID:                 tenantID,
		oAuthConfig:              oAuthConfig,
		popClaims:                popClaims,
		disableInstanceDiscovery: disableInstanceDiscovery,
	}, nil
}

// Token fetches an azcore.AccessToken from the Azure SDK and converts it to an adal.Token for use with kubelogin.
func (p *resourceOwnerToken) Token(ctx context.Context) (adal.Token, error) {
	return p.tokenWithOptions(ctx, nil)
}

func (p *resourceOwnerToken) tokenWithOptions(ctx context.Context, options *azcore.ClientOptions) (adal.Token, error) {
	emptyToken := adal.Token{}
	authorityFromConfig := p.oAuthConfig.AuthorityEndpoint
	clientOpts := azcore.ClientOptions{Cloud: cloud.Configuration{
		ActiveDirectoryAuthorityHost: authorityFromConfig.String(),
	}}
	if options != nil {
		clientOpts = *options
	}
	var err error
	scopes := []string{p.resourceID + defaultScope}
	if len(p.popClaims) > 0 {
		// If PoP token support is enabled and the correct u-claim is provided, use the MSAL
		// token provider to acquire a new token
		token, expirationTimeUnix, err := pop.AcquirePoPTokenByUsernamePassword(
			ctx,
			p.popClaims,
			scopes,
			p.username,
			p.password,
			&pop.PublicClientOptions{
				Authority:                authorityFromConfig.String(),
				ClientID:                 p.clientID,
				DisableInstanceDiscovery: p.disableInstanceDiscovery,
				Options:                  &clientOpts,
			},
		)
		if err != nil {
			return emptyToken, fmt.Errorf("failed to create PoP token using resource owner flow: %w", err)
		}

		// azurecore.AccessTokens have ExpiresOn as Time.Time. We need to convert it to JSON.Number
		// by fetching the time in seconds since the Unix epoch via Unix() and then converting to a
		// JSON.Number via formatting as a string using a base-10 int64 conversion.
		expiresOn := json.Number(strconv.FormatInt(expirationTimeUnix, 10))

		// Re-wrap the azurecore.AccessToken into an adal.Token
		return adal.Token{
			AccessToken: token,
			ExpiresOn:   expiresOn,
			Resource:    p.resourceID,
		}, nil
	}

	// otherwise, if PoP token flow is not enabled, use the default flow
	callback := func(t adal.Token) error {
		return nil
	}
	spt, err := adal.NewServicePrincipalTokenFromUsernamePassword(
		p.oAuthConfig,
		p.clientID,
		p.username,
		p.password,
		p.resourceID,
		callback)
	if err != nil {
		return emptyToken, fmt.Errorf("failed to create service principal token from username password: %s", err)
	}

	err = spt.RefreshWithContext(ctx)
	if err != nil {
		return emptyToken, err
	}
	return spt.Token(), nil
}

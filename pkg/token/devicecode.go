package token

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/kubelogin/pkg/pop"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

type deviceCodeTokenProvider struct {
	clientID    string
	resourceID  string
	tenantID    string
	oAuthConfig adal.OAuthConfig
	popClaims   map[string]string
}

func newDeviceCodeTokenProvider(oAuthConfig adal.OAuthConfig, clientID, resourceID, tenantID string, popClaims map[string]string) (TokenProvider, error) {
	if clientID == "" {
		return nil, errors.New("clientID cannot be empty")
	}
	if resourceID == "" {
		return nil, errors.New("resourceID cannot be empty")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}

	return &deviceCodeTokenProvider{
		clientID:    clientID,
		resourceID:  resourceID,
		tenantID:    tenantID,
		oAuthConfig: oAuthConfig,
		popClaims:   popClaims,
	}, nil
}

func (p *deviceCodeTokenProvider) Token() (adal.Token, error) {
	emptyToken := adal.Token{}
	client := &autorest.Client{}
	var token *adal.Token
	if p.popClaims == nil || len(p.popClaims) == 0 {
		deviceCode, err := adal.InitiateDeviceAuth(client, p.oAuthConfig, p.clientID, p.resourceID)
		if err != nil {
			return emptyToken, fmt.Errorf("initialing the device code authentication: %s", err)
		}

		_, err = fmt.Fprintln(os.Stderr, *deviceCode.Message)
		if err != nil {
			return emptyToken, fmt.Errorf("prompting the device code message: %s", err)
		}

		token, err = adal.WaitForUserCompletion(client, deviceCode)
		if err != nil {
			return emptyToken, fmt.Errorf("waiting for device code authentication to complete: %s", err)
		}
	} else {
		// if pop token option is enabled, convert the access token into a PoP token before wrapping
		// it into the adal token
		scopes := []string{p.resourceID + "/.default"}
		client, err := public.New(
			p.clientID,
			public.WithAuthority(p.oAuthConfig.AuthorityEndpoint.String()),
		)
		if err != nil {
			log.Fatal(err)
		}
		deviceCode, err := client.AcquireTokenByDeviceCode(
			context.Background(),
			scopes,
		)
		if err != nil {
			log.Fatal(err)
		}
		result, err := deviceCode.AuthenticationResult(context.Background())
		if err != nil {
			return emptyToken, fmt.Errorf("waiting for device code authentication to complete: %s", err)
		}

		authnScheme := pop.PopAuthenticationScheme{
			Host:   p.popClaims["u"],
			PoPKey: pop.GetSwPoPKey(),
		}
		formatted, err := authnScheme.FormatAccessToken(result.AccessToken)
		if err != nil {
			return emptyToken, fmt.Errorf("waiting for device code authentication to complete: %s", err)
		}
		expiresOn := json.Number(strconv.FormatInt(result.IDToken.ExpirationTime, 10))

		// Re-wrap the azurecore.AccessToken into an adal.Token
		token = &adal.Token{
			AccessToken: formatted,
			ExpiresOn:   expiresOn,
			Resource:    p.resourceID,
		}
	}

	return *token, nil
}

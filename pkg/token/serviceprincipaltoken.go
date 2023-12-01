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
)

const (
	certificate = "CERTIFICATE"
	privateKey  = "PRIVATE KEY"
)

type servicePrincipalToken struct {
	clientID           string
	clientSecret       string
	clientCert         string
	clientCertPassword string
	resourceID         string
	tenantID           string
	cloud              cloud.Configuration
	popClaims          map[string]string
}

func newServicePrincipalTokenProvider(
	cloud cloud.Configuration,
	clientID,
	clientSecret,
	clientCert,
	clientCertPassword,
	resourceID,
	tenantID string,
	popClaims map[string]string,
) (TokenProvider, error) {
	if clientID == "" {
		return nil, errors.New("clientID cannot be empty")
	}
	if clientSecret == "" && clientCert == "" {
		return nil, errors.New("both clientSecret and clientcert cannot be empty. One must be specified")
	}
	if clientSecret != "" && clientCert != "" {
		return nil, errors.New("client secret and client certificate cannot be set at the same time. Only one can be specified")
	}
	if resourceID == "" {
		return nil, errors.New("resourceID cannot be empty")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}

	return &servicePrincipalToken{
		clientID:           clientID,
		clientSecret:       clientSecret,
		clientCert:         clientCert,
		clientCertPassword: clientCertPassword,
		resourceID:         resourceID,
		tenantID:           tenantID,
		cloud:              cloud,
		popClaims:          popClaims,
	}, nil
}

// Token fetches an azcore.AccessToken from the Azure SDK and converts it to an adal.Token for use with kubelogin.
func (p *servicePrincipalToken) Token() (adal.Token, error) {
	return p.TokenWithOptions(nil)
}

func (p *servicePrincipalToken) TokenWithOptions(options *azcore.ClientOptions) (adal.Token, error) {
	ctx := context.Background()
	emptyToken := adal.Token{}
	var accessToken string
	var expirationTimeUnix int64
	var err error
	scopes := []string{p.resourceID + defaultScope}

	// Request a new Azure token provider for service principal
	if p.clientSecret != "" {
		accessToken, expirationTimeUnix, err = p.getTokenWithClientSecret(ctx, scopes, options)
		if err != nil {
			return emptyToken, fmt.Errorf("failed to create service principal token using secret: %w", err)
		}
	} else if p.clientCert != "" {
		accessToken, expirationTimeUnix, err = p.getTokenWithClientCert(ctx, scopes, options)
		if err != nil {
			return emptyToken, fmt.Errorf("failed to create service principal token using certificate: %w", err)
		}
	} else {
		return emptyToken, errors.New("service principal token requires either client secret or certificate")
	}

	if accessToken == "" {
		return emptyToken, errors.New("unexpectedly got empty access token")
	}

	// azurecore.AccessTokens have ExpiresOn as Time.Time. We need to convert it to JSON.Number
	// by fetching the time in seconds since the Unix epoch via Unix() and then converting to a
	// JSON.Number via formatting as a string using a base-10 int64 conversion.
	expiresOn := json.Number(strconv.FormatInt(expirationTimeUnix, 10))

	// Re-wrap the azurecore.AccessToken into an adal.Token
	return adal.Token{
		AccessToken: accessToken,
		ExpiresOn:   expiresOn,
		Resource:    p.resourceID,
	}, nil
}

package token

import (
	"errors"
	"fmt"

	"github.com/Azure/go-autorest/autorest/adal"
)

type servicePrincipalToken struct {
	clientID     string
	clientSecret string
	resourceID   string
	tenantID     string
	oAuthConfig  adal.OAuthConfig
}

func newServicePrincipalToken(oAuthConfig adal.OAuthConfig, clientID, clientSecret, resourceID, tenantID string) (TokenProvider, error) {
	if clientID == "" {
		return nil, errors.New("clientID cannot be empty")
	}
	if clientSecret == "" {
		return nil, errors.New("clientSecret cannot be empty")
	}
	if resourceID == "" {
		return nil, errors.New("resourceID cannot be empty")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}

	return &servicePrincipalToken{
		clientID:     clientID,
		clientSecret: clientSecret,
		resourceID:   resourceID,
		tenantID:     tenantID,
		oAuthConfig:  oAuthConfig,
	}, nil
}

func (p *servicePrincipalToken) Token() (adal.Token, error) {
	emptyToken := adal.Token{}
	callback := func(t adal.Token) error {
		return nil
	}
	spt, err := adal.NewServicePrincipalToken(
		p.oAuthConfig,
		p.clientID,
		p.clientSecret,
		p.resourceID,
		callback)
	if err != nil {
		return emptyToken, fmt.Errorf("failed to create service principal token: %s", err)
	}

	err = spt.Refresh()
	if err != nil {
		return emptyToken, err
	}
	return spt.Token(), nil
}

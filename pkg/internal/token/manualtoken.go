package token

import (
	"errors"
	"fmt"

	"github.com/Azure/go-autorest/autorest/adal"
)

type manualToken struct {
	clientID    string
	resourceID  string
	tenantID    string
	oAuthConfig adal.OAuthConfig
	token       adal.Token
}

func newManualToken(oAuthConfig adal.OAuthConfig, clientID, resourceID, tenantID string, token *adal.Token) (TokenProvider, error) {
	if token == nil {
		return nil, errors.New("token cannot be nil")
	}
	if clientID == "" {
		return nil, errors.New("clientID cannot be empty")
	}
	if resourceID == "" {
		return nil, errors.New("resourceID cannot be empty")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}

	provider := &manualToken{
		clientID:    clientID,
		resourceID:  resourceID,
		tenantID:    tenantID,
		oAuthConfig: oAuthConfig,
		token:       *token,
	}

	return provider, nil
}

func (p *manualToken) Token() (adal.Token, error) {
	emptyToken := adal.Token{}
	callback := func(t adal.Token) error {
		return nil
	}
	spt, err := adal.NewServicePrincipalTokenFromManualToken(
		p.oAuthConfig,
		p.clientID,
		p.resourceID,
		p.token,
		callback)
	if err != nil {
		return emptyToken, fmt.Errorf("failed to create service principal from manual token for token refresh: %s", err)
	}

	err = spt.Refresh()
	if err != nil {
		return emptyToken, err
	}
	return spt.Token(), nil
}

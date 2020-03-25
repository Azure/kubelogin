package token

import (
	"errors"
	"fmt"

	"github.com/Azure/go-autorest/autorest/adal"
)

type resourceOwnerToken struct {
	clientID    string
	username    string
	password    string
	resourceID  string
	tenantID    string
	oAuthConfig adal.OAuthConfig
}

func newResourceOwnerToken(oAuthConfig adal.OAuthConfig, clientID, username, password, resourceID, tenantID string) (TokenProvider, error) {
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
		clientID:    clientID,
		username:    username,
		password:    password,
		resourceID:  resourceID,
		tenantID:    tenantID,
		oAuthConfig: oAuthConfig,
	}, nil
}

func (p *resourceOwnerToken) Token() (adal.Token, error) {
	emptyToken := adal.Token{}
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

	err = spt.Refresh()
	if err != nil {
		return emptyToken, err
	}
	return spt.Token(), nil
}

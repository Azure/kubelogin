package token

import (
	"errors"
	"fmt"

	"github.com/Azure/go-autorest/autorest/adal"
)

type managedIdentityToken struct {
	clientID   string
	resourceID string
}

func newManagedIdentityToken(clientID, resourceID string) (TokenProvider, error) {
	if resourceID == "" {
		return nil, errors.New("resourceID cannot be empty")
	}

	provider := &managedIdentityToken{
		clientID:   clientID,
		resourceID: resourceID,
	}

	return provider, nil
}

func (p *managedIdentityToken) Token() (adal.Token, error) {
	var (
		spt        *adal.ServicePrincipalToken
		err        error
		emptyToken adal.Token
	)
	callback := func(t adal.Token) error {
		return nil
	}
	msiEndpoint, _ := adal.GetMSIVMEndpoint()
	if p.clientID == "" {
		spt, err = adal.NewServicePrincipalTokenFromMSI(
			msiEndpoint,
			p.resourceID,
			callback)
		if err != nil {
			return emptyToken, fmt.Errorf("failed to create service principal from managed identity for token refresh: %s", err)
		}
	} else {
		spt, err = adal.NewServicePrincipalTokenFromMSIWithUserAssignedID(
			msiEndpoint,
			p.resourceID,
			p.clientID,
			callback)
		if err != nil {
			return emptyToken, fmt.Errorf("failed to create service principal from managed identity using user assigned ID for token refresh: %s", err)
		}
	}

	err = spt.Refresh()
	if err != nil {
		return emptyToken, err
	}
	return spt.Token(), nil
}

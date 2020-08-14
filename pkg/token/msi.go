package token

import (
	"errors"
	"fmt"

	"github.com/Azure/go-autorest/autorest/adal"
)

type managedIdentityToken struct {
	clientID           string
	identityResourceID string
	resourceID         string
}

func newManagedIdentityToken(clientID, identityResourceID, resourceID string) (TokenProvider, error) {
	if resourceID == "" {
		return nil, errors.New("resourceID cannot be empty")
	}

	provider := &managedIdentityToken{
		clientID:           clientID,
		identityResourceID: identityResourceID,
		resourceID:         resourceID,
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
		if p.identityResourceID == "" {
			// no identity specified, use whatever IMDS default to
			spt, err = adal.NewServicePrincipalTokenFromMSI(
				msiEndpoint,
				p.resourceID,
				callback)
			if err != nil {
				return emptyToken, fmt.Errorf("failed to create service principal from managed identity for token refresh: %s", err)
			}
		}

		// use a specified managedIdentity resource id
		spt, err = adal.NewServicePrincipalTokenFromMSIWithIdentityResourceID(
			msiEndpoint,
			p.resourceID,
			p.identityResourceID,
			callback)
		if err != nil {
			return emptyToken, fmt.Errorf("failed to create service principal from managed identity %s for token refresh: %s", p.identityResourceID, err)
		}
	} else {

		// use a specified clientId
		spt, err = adal.NewServicePrincipalTokenFromMSIWithUserAssignedID(
			msiEndpoint,
			p.resourceID,
			p.clientID,
			callback)
		if err != nil {
			return emptyToken, fmt.Errorf("failed to create service principal from managed identity %s for token refresh: %s", p.clientID, err)
		}
	}

	err = spt.Refresh()
	if err != nil {
		return emptyToken, err
	}
	return spt.Token(), nil
}

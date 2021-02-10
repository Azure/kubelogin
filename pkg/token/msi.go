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

	// there are multiple options to login with MSI: https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/how-to-use-vm-token#get-a-token-using-http
	// 1. use clientId if present
	// 2. use identityResourceID if present
	// 3. IMDS default

	if p.clientID == "" {
		if p.identityResourceID == "" {
			// no identity specified, use whatever IMDS default to
			spt, err = adal.NewServicePrincipalTokenFromManagedIdentity(
				p.resourceID,
				nil,
				callback)
			if err != nil {
				return emptyToken, fmt.Errorf("failed to create service principal from managed identity for token refresh: %s", err)
			}
		} else {
			// use a specified managedIdentity resource id
			spt, err = adal.NewServicePrincipalTokenFromManagedIdentity(
				p.resourceID,
				&adal.ManagedIdentityOptions{
					IdentityResourceID: p.identityResourceID,
				},
				callback)
			if err != nil {
				return emptyToken, fmt.Errorf("failed to create service principal from managed identity with identityResourceID %s for token refresh: %s", p.identityResourceID, err)
			}
		}
	} else {
		// use a specified clientId
		spt, err = adal.NewServicePrincipalTokenFromManagedIdentity(
			p.resourceID,
			&adal.ManagedIdentityOptions{
				ClientID: p.clientID,
			},
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

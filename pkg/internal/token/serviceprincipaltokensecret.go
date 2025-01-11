package token

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/kubelogin/pkg/internal/pop"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

// getTokenWithClientSecret requests a token using the configured client ID/secret
// and returns a PoP token if PoP claims are provided, otherwise returns a regular
// bearer token
func (p *servicePrincipalToken) getTokenWithClientSecret(
	context context.Context,
	scopes []string,
	options *azcore.ClientOptions,
) (string, int64, error) {
	clientOptions := &azidentity.ClientSecretCredentialOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: p.cloud,
		},
	}
	if options != nil {
		clientOptions.ClientOptions = *options
	}
	if len(p.popClaims) > 0 {
		// if PoP token support is enabled, use the PoP token flow to request the token
		return p.getPoPTokenWithClientSecret(context, scopes, options)
	}
	cred, err := azidentity.NewClientSecretCredential(
		p.tenantID,
		p.clientID,
		p.clientSecret,
		clientOptions,
	)
	if err != nil {
		return "", -1, fmt.Errorf("unable to create credential. Received: %w", err)
	}

	// Use the token provider to get a new token
	spnAccessToken, err := cred.GetToken(context, policy.TokenRequestOptions{Scopes: scopes})
	if err != nil {
		return "", -1, fmt.Errorf("failed to create service principal bearer token using secret: %w", err)
	}

	return spnAccessToken.Token, spnAccessToken.ExpiresOn.Unix(), nil
}

// getPoPTokenWithClientSecret requests a PoP token using the given client
// ID/secret and returns it
func (p *servicePrincipalToken) getPoPTokenWithClientSecret(
	context context.Context,
	scopes []string,
	options *azcore.ClientOptions,
) (string, int64, error) {
	cred, err := confidential.NewCredFromSecret(p.clientSecret)
	if err != nil {
		return "", -1, fmt.Errorf("unable to create credential. Received: %w", err)
	}
	accessToken, expiresOn, err := pop.AcquirePoPTokenConfidential(
		context,
		p.popClaims,
		scopes,
		cred,
		p.cloud.ActiveDirectoryAuthorityHost,
		p.clientID,
		p.tenantID,
		true,
		options,
		pop.GetSwPoPKey,
	)
	if err != nil {
		return "", -1, fmt.Errorf("failed to create service principal PoP token using secret: %w", err)
	}

	return accessToken, expiresOn, nil
}

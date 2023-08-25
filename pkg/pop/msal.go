// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pop

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

func AcquirePoPTokenInteractive(
	context context.Context,
	popClaims map[string]string,
	scopes []string,
	authority,
	clientID string,
) (string, int64, error) {
	client, err := public.New(clientID, public.WithAuthority(authority))
	if err != nil {
		return "", -1, fmt.Errorf("unable to create public client: %w", err)
	}
	result, err := client.AcquireTokenInteractive(
		context,
		scopes,
		public.WithAuthenticationScheme(
			&PopAuthenticationScheme{
				Host:   popClaims["u"],
				PoPKey: GetSwPoPKey(),
			},
		),
	)
	if err != nil {
		return "", -1, fmt.Errorf("failed to create PoP token with interactive flow: %w", err)
	}

	return result.AccessToken, result.ExpiresOn.Unix(), nil
}

func AcquirePoPTokenConfidential(
	context context.Context,
	popClaims map[string]string,
	scopes []string,
	cred confidential.Credential,
	authority,
	clientID,
	tenantID string,
	options *azcore.ClientOptions,
) (string, int64, error) {
	authnScheme := &PopAuthenticationScheme{
		Host:   popClaims["u"],
		PoPKey: GetSwPoPKey(),
	}
	client, err := confidential.New(
		authority,
		clientID,
		cred,
		confidential.WithHTTPClient(options.Transport.(*http.Client)),
	)
	if err != nil {
		return "", -1, fmt.Errorf("unable to create confidential client: %w", err)
	}
	result, err := client.AcquireTokenSilent(
		context,
		scopes,
		confidential.WithAuthenticationScheme(authnScheme),
		confidential.WithTenantID(tenantID),
	)
	if err != nil {
		result, err = client.AcquireTokenByCredential(
			context,
			scopes,
			confidential.WithAuthenticationScheme(authnScheme),
			confidential.WithTenantID(tenantID),
		)
		if err != nil {
			return "", -1, fmt.Errorf("failed to create service principal PoP token using secret: %w", err)
		}
	}

	return result.AccessToken, result.ExpiresOn.Unix(), nil
}

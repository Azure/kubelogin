// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pop

import (
	"context"
	"fmt"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

func AcquirePoPTokenInteractive(
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
		context.Background(),
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
	popClaims map[string]string,
	scopes []string,
	cred confidential.Credential,
	authority,
	clientID,
	tenantID string,
) (string, int64, error) {
	authnScheme := &PopAuthenticationScheme{
		Host:   popClaims["u"],
		PoPKey: GetSwPoPKey(),
	}
	client, err := confidential.New(
		authority,
		clientID,
		cred,
	)
	if err != nil {
		return "", -1, fmt.Errorf("unable to create confidential client: %w", err)
	}
	result, err := client.AcquireTokenSilent(
		context.Background(),
		scopes,
		confidential.WithAuthenticationScheme(authnScheme),
		confidential.WithTenantID(tenantID),
	)
	if err != nil {
		result, err = client.AcquireTokenByCredential(
			context.Background(),
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

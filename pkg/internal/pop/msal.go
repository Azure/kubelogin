package pop

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

// AcquirePoPTokenInteractive acquires a PoP token using MSAL's interactive login flow.
// Requires user to authenticate via browser
func AcquirePoPTokenInteractive(
	context context.Context,
	popClaims map[string]string,
	scopes []string,
	authority,
	clientID string,
	options *azcore.ClientOptions,
) (string, int64, error) {
	var client *public.Client
	var err error
	client, err = getPublicClient(authority, clientID, options)
	if err != nil {
		return "", -1, err
	}

	popKey, err := GetSwPoPKey()
	if err != nil {
		return "", -1, err
	}
	result, err := client.AcquireTokenInteractive(
		context,
		scopes,
		public.WithAuthenticationScheme(
			&PoPAuthenticationScheme{
				Host:   popClaims["u"],
				PoPKey: popKey,
			},
		),
	)
	if err != nil {
		return "", -1, fmt.Errorf("failed to create PoP token with interactive flow: %w", err)
	}

	return result.AccessToken, result.ExpiresOn.Unix(), nil
}

// AcquirePoPTokenByUsernamePassword acquires a PoP token using MSAL's username/password login flow
// This flow does not require user interaction as credentials have already been provided
func AcquirePoPTokenByUsernamePassword(
	context context.Context,
	popClaims map[string]string,
	scopes []string,
	authority,
	clientID,
	username,
	password string,
	options *azcore.ClientOptions,
) (string, int64, error) {
	client, err := getPublicClient(authority, clientID, options)
	if err != nil {
		return "", -1, err
	}

	popKey, err := GetSwPoPKey()
	if err != nil {
		return "", -1, err
	}
	result, err := client.AcquireTokenByUsernamePassword(
		context,
		scopes,
		username,
		password,
		public.WithAuthenticationScheme(
			&PoPAuthenticationScheme{
				Host:   popClaims["u"],
				PoPKey: popKey,
			},
		),
	)
	if err != nil {
		return "", -1, fmt.Errorf("failed to create PoP token with username/password flow: %w", err)
	}

	return result.AccessToken, result.ExpiresOn.Unix(), nil
}

// AcquirePoPTokenConfidential acquires a PoP token using MSAL's confidential login flow.
// This flow does not require user interaction as the credentials for the request have
// already been provided
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
	popKey, err := GetSwPoPKey()
	if err != nil {
		return "", -1, err
	}
	authnScheme := &PoPAuthenticationScheme{
		Host:   popClaims["u"],
		PoPKey: popKey,
	}
	var client confidential.Client
	if options != nil && options.Transport != nil {
		client, err = confidential.New(
			authority,
			clientID,
			cred,
			confidential.WithHTTPClient(options.Transport.(*http.Client)),
		)
	} else {
		client, err = confidential.New(
			authority,
			clientID,
			cred,
		)
	}
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

// getPublicClient returns an instance of the msal `public` client based on the provided options
func getPublicClient(
	authority,
	clientID string,
	options *azcore.ClientOptions,
) (*public.Client, error) {
	var client public.Client
	var err error
	if options != nil && options.Transport != nil {
		client, err = public.New(
			clientID,
			public.WithAuthority(authority),
			public.WithHTTPClient(options.Transport.(*http.Client)),
		)
	} else {
		client, err = public.New(
			clientID,
			public.WithAuthority(authority),
		)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to create public client: %w", err)
	}

	return &client, nil
}

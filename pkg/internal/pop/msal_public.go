package pop

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

// AcquirePoPTokenInteractive acquires a PoP token using MSAL's interactive login flow.
// Requires user to authenticate via browser
func AcquirePoPTokenInteractive(
	context context.Context,
	popClaims map[string]string,
	scopes []string,
	msalOptions *MsalClientOptions,
) (string, int64, error) {
	var client *public.Client
	var err error
	client, err = getPublicClient(msalOptions)
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
		public.WithTenantID(msalOptions.TenantID),
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
	username,
	password string,
	msalOptions *MsalClientOptions,
) (string, int64, error) {
	client, err := getPublicClient(msalOptions)
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
		public.WithTenantID(msalOptions.TenantID),
	)
	if err != nil {
		return "", -1, fmt.Errorf("failed to create PoP token with username/password flow: %w", err)
	}

	return result.AccessToken, result.ExpiresOn.Unix(), nil
}

// getPublicClient returns an instance of the msal `public` client based on the provided options
// The instance discovery should be disabled on private cloud
func getPublicClient(msalOptions *MsalClientOptions) (*public.Client, error) {
	var client public.Client
	var err error
	if msalOptions == nil {
		return nil, fmt.Errorf("unable to create public client: MsalClientOptions is empty")
	}
	if msalOptions.Options.Transport != nil {
		client, err = public.New(
			msalOptions.ClientID,
			public.WithAuthority(msalOptions.Authority),
			public.WithHTTPClient(msalOptions.Options.Transport.(*http.Client)),
			public.WithInstanceDiscovery(!msalOptions.DisableInstanceDiscovery),
		)
	} else {
		client, err = public.New(
			msalOptions.ClientID,
			public.WithAuthority(msalOptions.Authority),
			public.WithInstanceDiscovery(!msalOptions.DisableInstanceDiscovery),
		)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to create public client: %w", err)
	}

	return &client, nil
}

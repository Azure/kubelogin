package pop

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

type PublicClientOptions struct {
	Authority                string
	ClientID                 string
	DisableInstanceDiscovery bool
	Options                  *azcore.ClientOptions
}

// AcquirePoPTokenInteractive acquires a PoP token using MSAL's interactive login flow.
// Requires user to authenticate via browser
func AcquirePoPTokenInteractive(
	context context.Context,
	popClaims map[string]string,
	scopes []string,
	pcOptions *PublicClientOptions,
) (string, int64, error) {
	var client *public.Client
	var err error
	client, err = getPublicClient(pcOptions)
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
	username,
	password string,
	pcOptions *PublicClientOptions,
) (string, int64, error) {
	client, err := getPublicClient(pcOptions)
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

// getPublicClient returns an instance of the msal `public` client based on the provided options
// The instance discovery will be disable on private cloud
func getPublicClient(pcOptions *PublicClientOptions) (*public.Client, error) {
	var client public.Client
	var err error
	if pcOptions == nil {
		return nil, fmt.Errorf("unable to create public client: publicClientOptions is empty")
	}
	if pcOptions.Options != nil && pcOptions.Options.Transport != nil {
		client, err = public.New(
			pcOptions.ClientID,
			public.WithAuthority(pcOptions.Authority),
			public.WithHTTPClient(pcOptions.Options.Transport.(*http.Client)),
			public.WithInstanceDiscovery(!pcOptions.DisableInstanceDiscovery),
		)
	} else {
		client, err = public.New(
			pcOptions.ClientID,
			public.WithAuthority(pcOptions.Authority),
			public.WithInstanceDiscovery(!pcOptions.DisableInstanceDiscovery),
		)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to create public client: %w", err)
	}

	return &client, nil
}

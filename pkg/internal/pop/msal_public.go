package pop

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

// PublicClientOptions holds options for creating a public client
type PublicClientOptions struct {
	Cache cache.ExportReplace
}

// PublicClientOption defines a functional option for configuring a public client
type PublicClientOption func(*PublicClientOptions)

// WithCustomCachePublic adds a custom cache to the confidential client
func WithCustomCachePublic(cache cache.ExportReplace) PublicClientOption {
	return func(opts *PublicClientOptions) {
		opts.Cache = cache
	}
}

// NewPublicClient creates a new public client with default options
func NewPublicClient(
	msalOptions *MsalClientOptions,
	options ...PublicClientOption,
) (public.Client, error) {
	if msalOptions == nil {
		return public.Client{}, fmt.Errorf("unable to create public client: msalClientOptions is empty")
	}

	// Apply custom options
	clientOpts := &PublicClientOptions{}
	for _, option := range options {
		option(clientOpts)
	}

	// Build public options
	var publicOptions []public.Option
	publicOptions = append(publicOptions,
		public.WithInstanceDiscovery(!msalOptions.DisableInstanceDiscovery),
		public.WithAuthority(msalOptions.Authority),
	)

	// Add HTTP client if present in msalOptions
	if msalOptions.Options.Transport != nil {
		client, ok := msalOptions.Options.Transport.(*http.Client)
		if !ok {
			return public.Client{}, fmt.Errorf("unable to create public client: msalOptions.Options.Transport is not an *http.Client")
		}
		publicOptions = append(publicOptions,
			public.WithHTTPClient(client),
		)
	}

	// Add cache if specified
	if clientOpts.Cache != nil {
		publicOptions = append(publicOptions, public.WithCache(clientOpts.Cache))
	}

	client, err := public.New(
		msalOptions.ClientID,
		publicOptions...,
	)

	if err != nil {
		return public.Client{}, fmt.Errorf("unable to create public client: %w", err)
	}

	return client, nil
}

// AcquirePoPTokenInteractive acquires a PoP token using MSAL's interactive login flow.
// Requires user to authenticate via browser
func AcquirePoPTokenInteractive(
	context context.Context,
	popClaims map[string]string,
	scopes []string,
	client public.Client,
	msalOptions *MsalClientOptions,
) (string, int64, error) {

	var err error
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
	client public.Client,
	username,
	password string,
	msalOptions *MsalClientOptions,
) (string, int64, error) {
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

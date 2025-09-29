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

// AcquirePoPTokenInteractive acquires a PoP token using MSAL's interactive login flow with persistent key and single-user caching.
// It first tries to acquire a token silently from cache, and only falls back to interactive login if needed.
// Uses persistent PoP key for proper token caching and implements single-user cache (latest user wins).
// If silent token acquisition fails, the cache is automatically cleared to ensure clean state.
// Requires user to authenticate via browser only when no valid cached tokens exist.
func AcquirePoPTokenInteractive(
	context context.Context,
	popClaims map[string]string,
	scopes []string,
	client public.Client,
	msalOptions *MsalClientOptions,
	cacheDir string,
) (string, int64, error) {

	var err error
	popKey, err := GetSwPoPKeyPersistent(cacheDir)
	if err != nil {
		return "", -1, err
	}

	authnScheme := &PoPAuthenticationScheme{
		Host:   popClaims["u"],
		PoPKey: popKey,
	}

	// Try silent token acquisition first if accounts exist
	accounts, err := client.Accounts(context)
	if err == nil && len(accounts) > 0 {
		// Try silent acquisition with cached account
		result, err := client.AcquireTokenSilent(
			context,
			scopes,
			public.WithSilentAccount(accounts[0]),
			public.WithAuthenticationScheme(authnScheme),
			public.WithTenantID(msalOptions.TenantID),
		)
		if err == nil {
			return result.AccessToken, result.ExpiresOn.Unix(), nil
		}

		// Silent acquisition failed - clear cache to ensure single-user behavior
		// This handles token expiration, user switching, and cache corruption
		clearErr := clearAllAccounts(context, client)
		if clearErr != nil {
			return "", -1, fmt.Errorf("failed to clear cache after silent acquisition failure: %w", clearErr)
		}
	}

	// Interactive login (first time, cache cleared, or token refresh)
	result, err := client.AcquireTokenInteractive(
		context,
		scopes,
		public.WithAuthenticationScheme(authnScheme),
		public.WithTenantID(msalOptions.TenantID),
	)
	if err != nil {
		return "", -1, fmt.Errorf("failed to create PoP token with interactive flow: %w", err)
	}

	return result.AccessToken, result.ExpiresOn.Unix(), nil
}

// AcquirePoPTokenByUsernamePassword acquires a PoP token using MSAL's username/password login flow with persistent key and single-user caching.
// It first tries to acquire a token silently from cache, and only falls back to username/password login if needed.
// Uses persistent PoP key for proper token caching and implements single-user cache (latest user wins).
// This flow does not require user interaction as credentials have already been provided.
func AcquirePoPTokenByUsernamePassword(
	context context.Context,
	popClaims map[string]string,
	scopes []string,
	client public.Client,
	username,
	password string,
	msalOptions *MsalClientOptions,
	cacheDir string,
) (string, int64, error) {

	var err error
	popKey, err := GetSwPoPKeyPersistent(cacheDir)
	if err != nil {
		return "", -1, err
	}

	authnScheme := &PoPAuthenticationScheme{
		Host:   popClaims["u"],
		PoPKey: popKey,
	}

	// Try silent token acquisition first if accounts exist
	accounts, err := client.Accounts(context)
	if err == nil && len(accounts) > 0 {
		// Try silent acquisition with cached account
		result, err := client.AcquireTokenSilent(
			context,
			scopes,
			public.WithSilentAccount(accounts[0]),
			public.WithAuthenticationScheme(authnScheme),
			public.WithTenantID(msalOptions.TenantID),
		)
		if err == nil {
			return result.AccessToken, result.ExpiresOn.Unix(), nil
		}

		// Silent acquisition failed - clear cache to ensure single-user behavior
		// This handles token expiration, user switching, and cache corruption
		clearErr := clearAllAccounts(context, client)
		if clearErr != nil {
			return "", -1, fmt.Errorf("failed to clear cache after silent acquisition failure: %w", clearErr)
		}
	}

	// Username/password login (first time, user switch, or token refresh)
	result, err := client.AcquireTokenByUsernamePassword(
		context,
		scopes,
		username,
		password,
		public.WithAuthenticationScheme(authnScheme),
		public.WithTenantID(msalOptions.TenantID),
	)
	if err != nil {
		return "", -1, fmt.Errorf("failed to create PoP token with username/password flow: %w", err)
	}

	return result.AccessToken, result.ExpiresOn.Unix(), nil
}

// clearAllAccounts removes all cached accounts from the MSAL client.
// This is used to implement single-user caching where only the latest authenticated user is cached.
func clearAllAccounts(ctx context.Context, client public.Client) error {
	accounts, err := client.Accounts(ctx)
	if err != nil {
		return err
	}

	for _, account := range accounts {
		err = client.RemoveAccount(ctx, account)
		if err != nil {
			return fmt.Errorf("failed to remove account %s: %w", account.PreferredUsername, err)
		}
	}
	return nil
}

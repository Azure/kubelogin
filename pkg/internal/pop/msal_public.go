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

// AcquirePoPTokenInteractive acquires a PoP token using MSAL's interactive login flow with caching.
// First attempts silent token acquisition if a single account is cached.
// Uses the provided PoP key for proper token caching.
// Falls back to interactive authentication if silent acquisition fails or no accounts are cached.
func AcquirePoPTokenInteractive(
	context context.Context,
	popClaims map[string]string,
	scopes []string,
	client public.Client,
	msalOptions *MsalClientOptions,
	popKey PoPKey,
) (string, int64, error) {

	authnScheme := &PoPAuthenticationScheme{
		Host:   popClaims["u"],
		PoPKey: popKey,
	}

	// Try silent token acquisition first if accounts exist
	accounts, err := client.Accounts(context)
	if err == nil && len(accounts) > 0 {
		// Use the first account for silent acquisition (single-user cache)
		account := accounts[0]
		result, err := client.AcquireTokenSilent(
			context,
			scopes,
			public.WithSilentAccount(account),
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

	// Interactive login (first time or after cache cleared due to silent acquisition failure)
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

// AcquirePoPTokenByUsernamePassword acquires a PoP token using MSAL's username/password login flow with user-specific caching.
// It first tries to acquire a token silently from cache for the specific username, and only falls back to username/password login if needed.
// Uses the provided PoP key for proper token caching. If the cache contains tokens for a different user,
// it clears the cache and authenticates with the provided credentials.
// This flow does not require user interaction as credentials have already been provided.
func AcquirePoPTokenByUsernamePassword(
	context context.Context,
	popClaims map[string]string,
	scopes []string,
	client public.Client,
	username,
	password string,
	msalOptions *MsalClientOptions,
	popKey PoPKey,
) (string, int64, error) {

	authnScheme := &PoPAuthenticationScheme{
		Host:   popClaims["u"],
		PoPKey: popKey,
	}

	// Try silent token acquisition first if accounts exist for the specific username
	targetAccount, err := findAccountByUsername(context, client, username)
	if err == nil && targetAccount != nil {
		// Try silent acquisition with the matching account
		result, err := client.AcquireTokenSilent(
			context,
			scopes,
			public.WithSilentAccount(*targetAccount),
			public.WithAuthenticationScheme(authnScheme),
			public.WithTenantID(msalOptions.TenantID),
		)
		if err == nil {
			return result.AccessToken, result.ExpiresOn.Unix(), nil
		}

		// Silent acquisition failed - clear cache to ensure clean state for username/password authentication
		clearErr := clearAllAccounts(context, client)
		if clearErr != nil {
			return "", -1, fmt.Errorf("failed to clear cache before username/password authentication: %w", clearErr)
		}
	}

	// Username/password login (first time, user switch, or after cache cleared due to silent acquisition failure)
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

// findAccountByUsername searches for a cached account with the specified username.
// Returns the account if found, nil otherwise.
func findAccountByUsername(ctx context.Context, client public.Client, username string) (*public.Account, error) {
	accounts, err := client.Accounts(ctx)
	if err != nil {
		return nil, err
	}

	for _, account := range accounts {
		if account.PreferredUsername == username {
			return &account, nil
		}
	}
	return nil, nil
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

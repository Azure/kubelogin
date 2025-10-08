package pop

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

type MsalClientOptions struct {
	Authority                string
	ClientID                 string
	TenantID                 string
	DisableInstanceDiscovery bool
	Options                  azcore.ClientOptions
}

// ConfidentialClientOptions holds options for creating a confidential client
type ConfidentialClientOptions struct {
	Cache cache.ExportReplace
}

// ConfidentialClientOption defines a functional option for configuring a confidential client
type ConfidentialClientOption func(*ConfidentialClientOptions)

// WithCustomCacheConfidential adds a custom cache to the confidential client
func WithCustomCacheConfidential(cache cache.ExportReplace) ConfidentialClientOption {
	return func(opts *ConfidentialClientOptions) {
		opts.Cache = cache
	}
}

// NewConfidentialClient creates a new confidential client with default options
func NewConfidentialClient(
	cred confidential.Credential,
	msalOptions *MsalClientOptions,
	options ...ConfidentialClientOption,
) (confidential.Client, error) {
	if msalOptions == nil {
		return confidential.Client{}, fmt.Errorf("unable to create confidential client: msalClientOptions is empty")
	}

	// Apply custom options
	clientOpts := &ConfidentialClientOptions{}
	for _, option := range options {
		option(clientOpts)
	}

	// Build confidential options
	var confOptions []confidential.Option
	confOptions = append(confOptions,
		confidential.WithX5C(),
		confidential.WithInstanceDiscovery(!msalOptions.DisableInstanceDiscovery),
	)

	// Add HTTP client if present in msalOptions
	if msalOptions.Options.Transport != nil {
		client, ok := msalOptions.Options.Transport.(*http.Client)
		if !ok {
			return confidential.Client{}, fmt.Errorf("unable to create confidential client: msalOptions.Options.Transport is not an *http.Client")
		}
		confOptions = append(confOptions,
			confidential.WithHTTPClient(client),
		)
	}

	// Add cache if specified
	if clientOpts.Cache != nil {
		confOptions = append(confOptions, confidential.WithCache(clientOpts.Cache))
	}

	client, err := confidential.New(
		msalOptions.Authority,
		msalOptions.ClientID,
		cred,
		confOptions...,
	)

	if err != nil {
		return confidential.Client{}, fmt.Errorf("unable to create confidential client: %w", err)
	}

	return client, nil
}

// AcquirePoPTokenConfidential acquires a PoP token using MSAL's confidential login flow.
// It first tries to acquire a token silently from cache, and only falls back to credential-based login if needed.
// Uses the provided PoP key for token acquisition and caching.
// This flow does not require user interaction as the credentials for the request have already been provided.
func AcquirePoPTokenConfidential(
	ctx context.Context,
	popClaims map[string]string,
	scopes []string,
	client confidential.Client,
	tenantID string,
	popKey PoPKey,
) (string, int64, error) {

	authnScheme := &PoPAuthenticationScheme{
		Host:   popClaims["u"],
		PoPKey: popKey,
	}

	// Try silent token acquisition first
	result, err := client.AcquireTokenSilent(
		ctx,
		scopes,
		confidential.WithAuthenticationScheme(authnScheme),
		confidential.WithTenantID(tenantID),
	)
	if err == nil {
		return result.AccessToken, result.ExpiresOn.Unix(), nil
	}

	// Silent acquisition failed - proceed to credential-based acquisition
	// Note: For confidential clients (service principals), MSAL will handle cache updates automatically
	result, err = client.AcquireTokenByCredential(
		ctx,
		scopes,
		confidential.WithAuthenticationScheme(authnScheme),
		confidential.WithTenantID(tenantID),
	)
	if err != nil {
		return "", -1, fmt.Errorf("failed to create service principal PoP token using credential: %w", err)
	}

	return result.AccessToken, result.ExpiresOn.Unix(), nil
}

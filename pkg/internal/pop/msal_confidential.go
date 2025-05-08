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

// ClientOptions holds options for creating a confidential client
type ClientOptions struct {
	Cache cache.ExportReplace
}

// ConfidentialClientOption defines a functional option for configuring a confidential client
type ConfidentialClientOption func(*ClientOptions)

// WithCustomCache adds a custom cache to the confidential client
func WithCustomCache(cache cache.ExportReplace) ConfidentialClientOption {
	return func(opts *ClientOptions) {
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
	clientOpts := &ClientOptions{}
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
		confOptions = append(confOptions,
			confidential.WithHTTPClient(msalOptions.Options.Transport.(*http.Client)),
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
// This flow does not require user interaction as the credentials for the request have
// already been provided
// instanceDisovery is to be false only in disconnected clouds to disable instance discovery and authoority validation
func AcquirePoPTokenConfidential(
	context context.Context,
	popClaims map[string]string,
	scopes []string,
	client confidential.Client,
	tenantID string,
	popKeyFunc func() (*SwKey, error),
) (string, int64, error) {
	if popKeyFunc == nil {
		popKeyFunc = GetSwPoPKey
	}
	popKey, err := popKeyFunc()
	if err != nil {
		return "", -1, fmt.Errorf("unable to get PoP key: %w", err)
	}

	authnScheme := &PoPAuthenticationScheme{
		Host:   popClaims["u"],
		PoPKey: popKey,
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

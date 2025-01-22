package pop

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

type MsalClientOptions struct {
	Authority                string
	ClientID                 string
	TenantID                 string
	DisableInstanceDiscovery bool
	Options                  *azcore.ClientOptions
}

// AcquirePoPTokenConfidential acquires a PoP token using MSAL's confidential login flow.
// This flow does not require user interaction as the credentials for the request have
// already been provided
// instanceDisovery is to be false only in disconnected clouds to disable instance discovery and authoority validation
func AcquirePoPTokenConfidential(
	context context.Context,
	popClaims map[string]string,
	scopes []string,
	cred confidential.Credential,
	msalOptions *MsalClientOptions,
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
	var client confidential.Client

	if msalOptions == nil {
		return "", -1, fmt.Errorf("unable to create confidential client: msalClientOptions is empty")
	}

	if msalOptions.Options != nil && msalOptions.Options.Transport != nil {
		client, err = confidential.New(
			msalOptions.Authority,
			msalOptions.ClientID,
			cred,
			confidential.WithHTTPClient(msalOptions.Options.Transport.(*http.Client)),
			confidential.WithX5C(),
			confidential.WithInstanceDiscovery(!msalOptions.DisableInstanceDiscovery),
		)
	} else {
		client, err = confidential.New(
			msalOptions.Authority,
			msalOptions.ClientID,
			cred,
			confidential.WithX5C(),
			confidential.WithInstanceDiscovery(!msalOptions.DisableInstanceDiscovery),
		)
	}
	if err != nil {
		return "", -1, fmt.Errorf("unable to create confidential client: %w", err)
	}
	result, err := client.AcquireTokenSilent(
		context,
		scopes,
		confidential.WithAuthenticationScheme(authnScheme),
		confidential.WithTenantID(msalOptions.TenantID),
	)
	if err != nil {
		result, err = client.AcquireTokenByCredential(
			context,
			scopes,
			confidential.WithAuthenticationScheme(authnScheme),
			confidential.WithTenantID(msalOptions.TenantID),
		)
		if err != nil {
			return "", -1, fmt.Errorf("failed to create service principal PoP token using secret: %w", err)
		}
	}

	return result.AccessToken, result.ExpiresOn.Unix(), nil
}

package pop

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

// AcquirePoPTokenConfidential acquires a PoP token using MSAL's confidential login flow.
// This flow does not require user interaction as the credentials for the request have
// already been provided
// instanceDisovery is to be false only in disconnected clouds to disable instance discovery and authoority validation
func AcquirePoPTokenConfidential(
	context context.Context,
	popClaims map[string]string,
	scopes []string,
	cred confidential.Credential,
	authority,
	clientID,
	tenantID string,
	instanceDiscovery bool,
	options *azcore.ClientOptions,
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
	if options != nil && options.Transport != nil {
		client, err = confidential.New(
			authority,
			clientID,
			cred,
			confidential.WithHTTPClient(options.Transport.(*http.Client)),
			confidential.WithX5C(),
			confidential.WithInstanceDiscovery(instanceDiscovery),
		)
	} else {
		client, err = confidential.New(
			authority,
			clientID,
			cred,
			confidential.WithX5C(),
			confidential.WithInstanceDiscovery(instanceDiscovery),
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

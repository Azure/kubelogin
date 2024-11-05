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
func AcquirePoPTokenConfidential(
	context context.Context,
	popClaims map[string]string,
	scopes []string,
	cred confidential.Credential,
	authority,
	clientID,
	tenantID string,
	options *azcore.ClientOptions,
	popKey *SwKey,
) (string, int64, error) {
	authnScheme := &PoPAuthenticationScheme{
		Host:   popClaims["u"],
		PoPKey: popKey,
	}
	var client confidential.Client
	var err error
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

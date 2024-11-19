package pop

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/kubelogin/pkg/internal/pop"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

// AcquirePoPTokenConfidential retrieves a Proof of Possession (PoP) token using confidential client credentials.
// It utilizes the internal pop.AcquirePoPTokenConfidential function to obtain the token.
func AcquirePoPTokenConfidential(
	context context.Context,
	popClaims map[string]string,
	scopes []string,
	cred confidential.Credential,
	authority,
	clientID,
	tenantID string,
	options *azcore.ClientOptions,
	popKeyFn func() (*SwKey, error),
) (string, int64, error) {

	// This function is necessary to type cast the function from *SwKey to *pop.SwKey.
	internalPopKeyFn := func() (*pop.SwKey, error) {
		key, err := popKeyFn()
		if err != nil {
			return nil, err
		}
		return &key.SwKey, nil
	}

	return pop.AcquirePoPTokenConfidential(context, popClaims, scopes, cred, authority, clientID, tenantID, options, internalPopKeyFn)
}

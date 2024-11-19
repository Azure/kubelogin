package pop

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/kubelogin/pkg/internal/pop"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

func AcquirePoPTokenConfidential(
	context context.Context,
	popClaims map[string]string,
	scopes []string,
	cred confidential.Credential,
	authority,
	clientID,
	tenantID string,
	options *azcore.ClientOptions,
	popKeyFunc func() (*SwKey, error),
) (string, int64, error) {

	internalPopKeyFunc := func() (*pop.SwKey, error) {
		key, err := popKeyFunc()
		if err != nil {
			return nil, err
		}
		return &key.SwKey, nil
	}

	return pop.AcquirePoPTokenConfidential(context, popClaims, scopes, cred, authority, clientID, tenantID, options, internalPopKeyFunc)
}

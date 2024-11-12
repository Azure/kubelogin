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
	popKey *SwKey,
) (string, int64, error) {
	return pop.AcquirePoPTokenConfidential(context, popClaims, scopes, cred, authority, clientID, tenantID, options, &popKey.SwKey)
}

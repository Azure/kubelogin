package pop

import (
	"github.com/Azure/kubelogin/pkg/internal/pop"
)

// This is the MSAL implementation of AuthenticationScheme.
// For more details, see the MSAL repo interface:
// https://github.com/AzureAD/microsoft-authentication-library-for-go/blob/4a4dafcbcbd7d57a69ed3bc59760381232c2be9c/apps/internal/oauth/ops/authority/authority.go#L146
type PoPAuthenticationScheme = pop.PoPAuthenticationScheme

type SwKey = pop.SwKey

type MsalClientOptions = pop.MsalClientOptions

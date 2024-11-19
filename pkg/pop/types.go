package pop

import (
	"github.com/Azure/kubelogin/pkg/internal/pop"
)

// This is the MSAL implementation of AuthenticationScheme.
// For more details, see the MSAL repo interface:
// https://github.com/AzureAD/microsoft-authentication-library-for-go/blob/4a4dafcbcbd7d57a69ed3bc59760381232c2be9c/apps/internal/oauth/ops/authority/authority.go#L146
type PoPAuthenticationScheme struct {
	pop.PoPAuthenticationScheme
}

// FormatAccessToken takes an access token, formats it as a PoP token,
// and returns it as a base-64 encoded string
func (as *PoPAuthenticationScheme) FormatAccessToken(accessToken string) (string, error) {
	return as.PoPAuthenticationScheme.FormatAccessToken(accessToken)
}

type SwKey struct {
	pop.SwKey
}

func (swKey *SwKey) JWK() string {
	return swKey.SwKey.JWK()
}

func (swKey *SwKey) JWKThumbprint() string {
	return swKey.SwKey.JWKThumbprint()
}

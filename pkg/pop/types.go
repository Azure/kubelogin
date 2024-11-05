package pop

import (
	"github.com/Azure/kubelogin/pkg/internal/pop"
)

type PoPAuthenticationScheme struct {
	pop.PoPAuthenticationScheme
}

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

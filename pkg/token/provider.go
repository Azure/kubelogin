package token

import (
	"context"

	"github.com/Azure/kubelogin/pkg/internal/token"
)

type tokenProviderShim struct {
	impl token.TokenProvider
}

var _ TokenProvider = (*tokenProviderShim)(nil)

func (tp *tokenProviderShim) GetAccessToken(ctx context.Context) (AccessToken, error) {
	t, err := tp.impl.Token(ctx)
	if err != nil {
		return AccessToken{}, err
	}

	rv := AccessToken{
		Token:     t.AccessToken,
		ExpiresOn: t.Expires(),
	}

	return rv, nil
}

// GetTokenProvider returns a token provider based on the given options.
func GetTokenProvider(options *Options) (TokenProvider, error) {
	impl, err := token.NewTokenProvider(options.toInternalOptions())
	if err != nil {
		return nil, err
	}

	rv := &tokenProviderShim{
		impl: impl,
	}

	return rv, nil
}

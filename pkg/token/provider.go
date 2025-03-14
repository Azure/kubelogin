package token

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/kubelogin/pkg/internal/token"
)

type tokenProviderShim struct {
	opts *token.Options
	cred token.CredentialProvider
}

var _ TokenProvider = (*tokenProviderShim)(nil)

func (tp *tokenProviderShim) GetAccessToken(ctx context.Context) (AccessToken, error) {
	tro := policy.TokenRequestOptions{
		TenantID: tp.opts.TenantID,
		Scopes:   []string{token.GetScope(tp.opts.ServerID)},
	}
	return tp.cred.GetToken(ctx, tro)
}

// GetTokenProvider returns a token provider based on the given options.
func GetTokenProvider(options *Options) (TokenProvider, error) {
	opts := options.toInternalOptions()
	cred, err := token.NewAzIdentityCredential(azidentity.AuthenticationRecord{}, opts)
	if err != nil {
		return nil, err
	}

	return &tokenProviderShim{
		cred: cred,
		opts: opts,
	}, nil
}

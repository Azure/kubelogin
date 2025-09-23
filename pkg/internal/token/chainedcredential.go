package token

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type ChainedCredential struct {
	cred *azidentity.DefaultAzureCredential
}

var _ CredentialProvider = (*ChainedCredential)(nil)

func newChainedCredential(opts *Options) (CredentialProvider, error) {
	azOpts := &azidentity.DefaultAzureCredentialOptions{
		ClientOptions:            azcore.ClientOptions{Cloud: opts.GetCloudConfiguration()},
		DisableInstanceDiscovery: opts.DisableInstanceDiscovery,
	}

	if opts.httpClient != nil {
		azOpts.ClientOptions.Transport = opts.httpClient
	}

	cred, err := azidentity.NewDefaultAzureCredential(azOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create chained credential: %w", err)
	}
	return &ChainedCredential{cred: cred}, nil
}

func (c *ChainedCredential) Name() string {
	return "ChainedCredential"
}

func (c *ChainedCredential) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	return azidentity.AuthenticationRecord{}, errAuthenticateNotSupported
}

func (c *ChainedCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return c.cred.GetToken(ctx, opts)
}

func (c *ChainedCredential) NeedAuthenticate() bool {
	return false
}

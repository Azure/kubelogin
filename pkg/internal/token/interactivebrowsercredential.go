package token

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"
)

type InteractiveBrowserCredential struct {
	cred *azidentity.InteractiveBrowserCredential
}

var _ CredentialProvider = (*InteractiveBrowserCredential)(nil)

func newInteractiveBrowserCredential(opts *Options, record azidentity.AuthenticationRecord) (CredentialProvider, error) {
	if opts.ClientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}
	if opts.TenantID == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	var (
		c   azidentity.Cache
		err error
	)
	if opts.UsePersistentCache {
		c, err = cache.New(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create cache: %s", err)
		}
	}

	cred, err := azidentity.NewInteractiveBrowserCredential(&azidentity.InteractiveBrowserCredentialOptions{
		ClientOptions:        azcore.ClientOptions{Cloud: opts.GetCloudConfiguration()},
		AuthenticationRecord: record,
		Cache:                c,
		ClientID:             opts.ClientID,
		TenantID:             opts.TenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create interactive browser credential: %s", err)
	}
	return &InteractiveBrowserCredential{cred: cred}, nil
}

func (c *InteractiveBrowserCredential) Name() string {
	return "InteractiveBrowserCredential"
}

func (c *InteractiveBrowserCredential) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	return c.cred.Authenticate(ctx, opts)
}

func (c *InteractiveBrowserCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return c.cred.GetToken(ctx, opts)
}

func (c *InteractiveBrowserCredential) NeedAuthenticate() bool {
	return true
}

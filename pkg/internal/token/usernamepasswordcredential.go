package token

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"
	"k8s.io/klog/v2"
)

type UsernamePasswordCredential struct {
	cred *azidentity.UsernamePasswordCredential
}

var _ CredentialProvider = (*UsernamePasswordCredential)(nil)

func newUsernamePasswordCredential(opts *Options, record azidentity.AuthenticationRecord) (CredentialProvider, error) {
	if opts.ClientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}
	if opts.TenantID == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	if opts.Username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if opts.Password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}
	var (
		c   azidentity.Cache
		err error
	)
	if opts.UsePersistentCache {
		c, err = cache.New(nil)
		if err != nil {
			klog.V(5).Infof("failed to create cache: %v", err)
		}
	}

	azOpts := &azidentity.UsernamePasswordCredentialOptions{
		ClientOptions:            azcore.ClientOptions{Cloud: opts.GetCloudConfiguration()},
		AuthenticationRecord:     record,
		Cache:                    c,
		DisableInstanceDiscovery: opts.DisableInstanceDiscovery,
	}

	if opts.httpClient != nil {
		azOpts.ClientOptions.Transport = opts.httpClient
	}

	cred, err := azidentity.NewUsernamePasswordCredential(
		opts.TenantID, opts.ClientID, opts.Username, opts.Password,
		azOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create username password credential: %w", err)
	}
	return &UsernamePasswordCredential{cred: cred}, nil
}

func (c *UsernamePasswordCredential) Name() string {
	return "UsernamePasswordCredential"
}

func (c *UsernamePasswordCredential) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	return c.cred.Authenticate(ctx, opts)
}

func (c *UsernamePasswordCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return c.cred.GetToken(ctx, opts)
}

func (c *UsernamePasswordCredential) NeedAuthenticate() bool {
	return true
}

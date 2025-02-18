package token

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"
)

type DeviceCodeCredential struct {
	cred *azidentity.DeviceCodeCredential
}

var _ CredentialProvider = (*DeviceCodeCredential)(nil)

func newDeviceCodeCredential(opts *Options, record azidentity.AuthenticationRecord) (CredentialProvider, error) {
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

	azOpts := &azidentity.DeviceCodeCredentialOptions{
		ClientOptions:        azcore.ClientOptions{Cloud: opts.GetCloudConfiguration()},
		AuthenticationRecord: record,
		Cache:                c,
		ClientID:             opts.ClientID,
		TenantID:             opts.TenantID,
	}

	if opts.httpClient != nil {
		azOpts.ClientOptions.Transport = opts.httpClient
	}

	cred, err := azidentity.NewDeviceCodeCredential(azOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create device code credential: %s", err)
	}
	return &DeviceCodeCredential{cred: cred}, nil
}

func (c *DeviceCodeCredential) Name() string {
	return "DeviceCodeCredential"
}

func (c *DeviceCodeCredential) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	return c.cred.Authenticate(ctx, opts)
}

func (c *DeviceCodeCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return c.cred.GetToken(ctx, opts)
}

func (c *DeviceCodeCredential) NeedAuthenticate() bool {
	return true
}

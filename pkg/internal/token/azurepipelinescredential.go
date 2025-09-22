package token

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"
	"k8s.io/klog/v2"

	"github.com/Azure/kubelogin/pkg/internal/env"
)

type AzurePipelinesCredential struct {
	cred *azidentity.AzurePipelinesCredential
}

var _ CredentialProvider = (*AzurePipelinesCredential)(nil)

func newAzurePipelinesCredential(opts *Options) (CredentialProvider, error) {
	systemAccessToken := os.Getenv(env.SystemAccessToken)
	if systemAccessToken == "" {
		return nil, fmt.Errorf("%s environment variable not set", env.SystemAccessToken)
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

	azOpts := &azidentity.AzurePipelinesCredentialOptions{
		ClientOptions:            azcore.ClientOptions{Cloud: opts.GetCloudConfiguration()},
		Cache:                    c,
		DisableInstanceDiscovery: opts.DisableInstanceDiscovery,
	}

	if opts.httpClient != nil {
		azOpts.Transport = opts.httpClient
	}

	cred, err := azidentity.NewAzurePipelinesCredential(
		opts.TenantID,
		opts.ClientID,
		opts.AzurePipelinesServiceConnectionID,
		systemAccessToken,
		azOpts,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure pipelines credential: %w", err)
	}
	return &AzurePipelinesCredential{cred: cred}, nil
}

func (c *AzurePipelinesCredential) Name() string {
	return "AzurePipelinesCredential"
}

func (c *AzurePipelinesCredential) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	return azidentity.AuthenticationRecord{}, errAuthenticateNotSupported
}

func (c *AzurePipelinesCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return c.cred.GetToken(ctx, opts)
}

func (c *AzurePipelinesCredential) NeedAuthenticate() bool {
	return false
}

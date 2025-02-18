package token

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"
)

const (
	actionsIDTokenRequestToken = "ACTIONS_ID_TOKEN_REQUEST_TOKEN"
	actionsIDTokenRequestURL   = "ACTIONS_ID_TOKEN_REQUEST_URL"
	azureADAudience            = "api://AzureADTokenExchange"
	defaultScope               = "/.default"
)

type WorkloadIdentityCredential struct {
	cred *azidentity.WorkloadIdentityCredential
}

var _ CredentialProvider = (*WorkloadIdentityCredential)(nil)

func newWorkloadIdentityCredential(opts *Options) (CredentialProvider, error) {
	if opts.ClientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}
	if opts.TenantID == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	if opts.FederatedTokenFile == "" {
		return nil, fmt.Errorf("federated token file cannot be empty")
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

	azOpts := &azidentity.WorkloadIdentityCredentialOptions{
		ClientOptions: azcore.ClientOptions{Cloud: opts.GetCloudConfiguration()},
		Cache:         c,
		ClientID:      opts.ClientID,
		TenantID:      opts.TenantID,
		TokenFilePath: opts.FederatedTokenFile,
	}

	if opts.httpClient != nil {
		azOpts.ClientOptions.Transport = opts.httpClient
	}

	cred, err := azidentity.NewWorkloadIdentityCredential(azOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create workload identity credential: %s", err)
	}
	return &WorkloadIdentityCredential{cred: cred}, nil
}

func (c *WorkloadIdentityCredential) Name() string {
	return "WorkloadIdentityCredential"
}

func (c *WorkloadIdentityCredential) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	panic("not implemented")
}

func (c *WorkloadIdentityCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return c.cred.GetToken(ctx, opts)
}

func (c *WorkloadIdentityCredential) NeedAuthenticate() bool {
	return false
}

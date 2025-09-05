package token

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type AzurePipelinesCredential struct {
	cred *azidentity.AzurePipelinesCredential
}

var _ CredentialProvider = (*AzurePipelinesCredential)(nil)

func newAzurePipelinesCredential(opts *Options) (CredentialProvider, error) {
	systemAccessToken := os.Getenv("SYSTEM_ACCESSTOKEN")
	if systemAccessToken == "" {
		return nil, fmt.Errorf("SYSTEM_ACCESSTOKEN environment variable not set")
	}

	cred, err := azidentity.NewAzurePipelinesCredential(
		opts.TenantID,
		opts.ClientID,
		opts.AzurePipelinesServiceConnectionID,
		systemAccessToken,
		nil,
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

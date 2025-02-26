package token

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/go-autorest/autorest/adal"
)

type ADALClientSecretCredential struct {
	oAuthConfig  adal.OAuthConfig
	clientID     string
	clientSecret string
}

var _ CredentialProvider = (*ADALClientSecretCredential)(nil)

func newADALClientSecretCredential(opts *Options) (CredentialProvider, error) {
	if !opts.IsLegacy {
		return nil, fmt.Errorf("ADALClientSecretCredential is not supported in non-legacy mode")
	}
	if opts.ClientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}
	if opts.TenantID == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	if opts.ClientSecret == "" {
		return nil, fmt.Errorf("client secret cannot be empty")
	}
	cloud := opts.GetCloudConfiguration()
	oAuthConfig, err := adal.NewOAuthConfig(cloud.ActiveDirectoryAuthorityHost, opts.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth config: %w", err)
	}
	return &ADALClientSecretCredential{
		oAuthConfig:  *oAuthConfig,
		clientID:     opts.ClientID,
		clientSecret: opts.ClientSecret,
	}, nil
}

func (c *ADALClientSecretCredential) Name() string {
	return "ADALClientSecretCredential"
}

func (c *ADALClientSecretCredential) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	return azidentity.AuthenticationRecord{}, errAuthenticateNotSupported
}

func (c *ADALClientSecretCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	// to keep backward compatibility,
	// 1. we only support one resource
	// 2. we remove the "/.default" suffix from the resource
	resource := strings.Replace(opts.Scopes[0], "/.default", "", 1)
	spt, err := adal.NewServicePrincipalToken(
		c.oAuthConfig,
		c.clientID,
		c.clientSecret,
		resource)
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("failed to create service principal token using secret: %w", err)
	}

	if err := spt.EnsureFreshWithContext(ctx); err != nil {
		return azcore.AccessToken{}, err
	}

	token := spt.Token()
	return azcore.AccessToken{Token: token.AccessToken, ExpiresOn: token.Expires()}, nil
}

func (c *ADALClientSecretCredential) NeedAuthenticate() bool {
	return false
}

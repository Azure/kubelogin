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

type ADALClientCertCredential struct {
	oAuthConfig        adal.OAuthConfig
	clientID           string
	clientCert         string
	clientCertPassword string
}

var _ CredentialProvider = (*ADALClientCertCredential)(nil)

func newADALClientCertCredential(opts *Options) (CredentialProvider, error) {
	if !opts.IsLegacy {
		return nil, fmt.Errorf("ADALClientCertCredential is not supported in non-legacy mode")
	}
	if opts.ClientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}
	if opts.TenantID == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	if opts.ClientCert == "" {
		return nil, fmt.Errorf("client certificate cannot be empty")
	}
	cloud := opts.GetCloudConfiguration()
	oAuthConfig, err := adal.NewOAuthConfig(cloud.ActiveDirectoryAuthorityHost, opts.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth config: %w", err)
	}
	return &ADALClientCertCredential{
		oAuthConfig:        *oAuthConfig,
		clientID:           opts.ClientID,
		clientCert:         opts.ClientCert,
		clientCertPassword: opts.ClientCertPassword,
	}, nil
}

func (c *ADALClientCertCredential) Name() string {
	return "ADALClientCertCredential"
}

func (c *ADALClientCertCredential) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	return azidentity.AuthenticationRecord{}, errAuthenticateNotSupported
}

func (c *ADALClientCertCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	// Get the certificate and private key from cert file
	cert, rsaPrivateKey, err := readCertificate(c.clientCert, c.clientCertPassword)
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("failed to read certificate: %w", err)
	}

	// to keep backward compatibility,
	// 1. we only support one resource
	// 2. we remove the "/.default" suffix from the resource
	resource := strings.Replace(opts.Scopes[0], "/.default", "", 1)
	spt, err := adal.NewServicePrincipalTokenFromCertificate(
		c.oAuthConfig,
		c.clientID,
		cert,
		rsaPrivateKey,
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

func (c *ADALClientCertCredential) NeedAuthenticate() bool {
	return false
}

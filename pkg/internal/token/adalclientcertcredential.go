package token

import (
	"context"
	"fmt"
	"os"
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
		return nil, fmt.Errorf("failed to create OAuth config: %s", err)
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
	panic("not implemented")
}

func (c *ADALClientCertCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	certData, err := os.ReadFile(c.clientCert)
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("failed to read the certificate file (%s): %w", c.clientCert, err)
	}

	// Get the certificate and private key from pfx file
	cert, rsaPrivateKey, err := decodePkcs12(certData, c.clientCertPassword)
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("failed to decode pkcs12 certificate while creating spt: %w", err)
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
		return azcore.AccessToken{}, fmt.Errorf("failed to create service principal token using secret: %s", err)
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

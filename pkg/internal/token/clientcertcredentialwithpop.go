package token

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/url"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/kubelogin/pkg/internal/pop"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

type ClientCertificateCredentialWithPoP struct {
	popClaims map[string]string
	cred      confidential.Credential
	client    confidential.Client
	options   *pop.MsalClientOptions
	cacheDir  string
}

var _ CredentialProvider = (*ClientCertificateCredentialWithPoP)(nil)

func newClientCertificateCredentialWithPoP(opts *Options) (CredentialProvider, error) {
	if opts.ClientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}
	if opts.TenantID == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	if opts.ClientCert == "" {
		return nil, fmt.Errorf("client certificate cannot be empty")
	}
	popClaimsMap, err := parsePoPClaims(opts.PoPTokenClaims)
	if err != nil {
		return nil, fmt.Errorf("unable to parse PoP claims: %w", err)
	}
	if len(popClaimsMap) == 0 {
		return nil, fmt.Errorf("number of pop claims is invalid: %d", len(popClaimsMap))
	}

	// Get the certificate and private key from cert file
	cert, rsaPrivateKey, err := readCertificate(opts.ClientCert, opts.ClientCertPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	cred, err := confidential.NewCredFromCert([]*x509.Certificate{cert}, rsaPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("unable to create credential from certificate: %w", err)
	}

	// Construct authority URL properly to avoid malformation
	authorityURL, err := url.JoinPath(opts.GetCloudConfiguration().ActiveDirectoryAuthorityHost, opts.TenantID)
	if err != nil {
		return nil, fmt.Errorf("unable to construct authority URL: %w", err)
	}

	msalOpts := &pop.MsalClientOptions{
		Authority:                authorityURL,
		ClientID:                 opts.ClientID,
		TenantID:                 opts.TenantID,
		DisableInstanceDiscovery: opts.DisableInstanceDiscovery,
	}
	if opts.httpClient != nil {
		msalOpts.Options.Transport = opts.httpClient
	}
	// Get cache from Options
	popCache := opts.GetPoPTokenCache()

	client, err := pop.NewConfidentialClient(
		cred,
		msalOpts,
		pop.WithCustomCacheConfidential(popCache),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create confidential client: %w", err)
	}

	// Only set cacheDir when cache is available
	var cacheDir string
	if popCache != nil {
		cacheDir = opts.AuthRecordCacheDir
	}

	return &ClientCertificateCredentialWithPoP{
		popClaims: popClaimsMap,
		cred:      cred,
		client:    client,
		options:   msalOpts,
		cacheDir:  cacheDir,
	}, nil
}

func (c *ClientCertificateCredentialWithPoP) Name() string {
	return "ClientCertificateCredentialWithPoP"
}

func (c *ClientCertificateCredentialWithPoP) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	return azidentity.AuthenticationRecord{}, errAuthenticateNotSupported
}

func (c *ClientCertificateCredentialWithPoP) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	// Get PoP key using centralized logic
	popKey, err := pop.GetPoPKeyByPolicy(c.cacheDir)
	if err != nil {
		return azcore.AccessToken{}, err
	}

	accessToken, expiresOn, err := pop.AcquirePoPTokenConfidential(
		ctx,
		c.popClaims,
		opts.Scopes,
		c.client,
		c.options.TenantID,
		popKey,
	)
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("failed to create PoP token using client certificate credential: %w", err)
	}
	return azcore.AccessToken{Token: accessToken, ExpiresOn: time.Unix(expiresOn, 0)}, nil
}

func (c *ClientCertificateCredentialWithPoP) NeedAuthenticate() bool {
	return false
}

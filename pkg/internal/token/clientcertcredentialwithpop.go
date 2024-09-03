package token

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/kubelogin/pkg/internal/pop"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

type ClientCertificateCredentialWithPoP struct {
	popClaims map[string]string
	cloud     cloud.Configuration
	tenantID  string
	clientID  string
	cred      confidential.Credential
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
		return nil, fmt.Errorf("unable to parse PoP claims: %s", err)
	}
	if len(popClaimsMap) == 0 {
		return nil, fmt.Errorf("number of pop claims is invalid: %d", len(popClaimsMap))
	}

	certData, err := os.ReadFile(opts.ClientCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read the certificate file (%s): %w", opts.ClientCert, err)
	}

	// Get the certificate and private key from pfx file
	cert, rsaPrivateKey, err := decodePkcs12(certData, opts.ClientCertPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decode pkcs12 certificate while creating spt: %w", err)
	}

	cred, err := confidential.NewCredFromCert([]*x509.Certificate{cert}, rsaPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("unable to create credential from certificate. Received: %w", err)
	}
	return &ClientCertificateCredentialWithPoP{
		cloud:     opts.GetCloudConfiguration(),
		tenantID:  opts.TenantID,
		clientID:  opts.ClientID,
		popClaims: popClaimsMap,
		cred:      cred,
	}, nil
}

func (c *ClientCertificateCredentialWithPoP) Name() string {
	return "ClientCertificateCredentialWithPoP"
}

func (c *ClientCertificateCredentialWithPoP) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	panic("not implemented")
}

func (c *ClientCertificateCredentialWithPoP) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	accessToken, expiresOn, err := pop.AcquirePoPTokenConfidential(
		ctx,
		c.popClaims,
		opts.Scopes,
		c.cred,
		c.cloud.ActiveDirectoryAuthorityHost,
		c.clientID,
		c.tenantID,
		nil,
		pop.GetSwPoPKey,
	)
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("failed to create PoP token using client certificate credential: %w", err)
	}
	return azcore.AccessToken{Token: accessToken, ExpiresOn: time.Unix(expiresOn, 0)}, nil
}

func (c *ClientCertificateCredentialWithPoP) NeedAuthenticate() bool {
	return false
}

package token

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/kubelogin/pkg/internal/pop"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

type ClientSecretCredentialWithPoP struct {
	popClaims map[string]string
	cloud     cloud.Configuration
	tenantID  string
	clientID  string
	cred      confidential.Credential
}

var _ CredentialProvider = (*ClientSecretCredentialWithPoP)(nil)

func newClientSecretCredentialWithPoP(opts *Options) (CredentialProvider, error) {
	if opts.ClientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}
	if opts.TenantID == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	if opts.ClientSecret == "" {
		return nil, fmt.Errorf("client secret cannot be empty")
	}
	popClaimsMap, err := parsePoPClaims(opts.PoPTokenClaims)
	if err != nil {
		return nil, fmt.Errorf("unable to parse PoP claims: %s", err)
	}
	if len(popClaimsMap) == 0 {
		return nil, fmt.Errorf("number of pop claims is invalid: %d", len(popClaimsMap))
	}

	cred, err := confidential.NewCredFromSecret(opts.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("unable to create confidential credential: %w", err)
	}
	return &ClientSecretCredentialWithPoP{
		cloud:     opts.GetCloudConfiguration(),
		tenantID:  opts.TenantID,
		clientID:  opts.ClientID,
		popClaims: popClaimsMap,
		cred:      cred,
	}, nil
}

func (c *ClientSecretCredentialWithPoP) Name() string {
	return "ClientSecretCredentialWithPoP"
}

func (c *ClientSecretCredentialWithPoP) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	panic("not implemented")
}

func (c *ClientSecretCredentialWithPoP) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
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
		return azcore.AccessToken{}, fmt.Errorf("failed to create PoP token using client secret credential: %w", err)
	}
	return azcore.AccessToken{Token: accessToken, ExpiresOn: time.Unix(expiresOn, 0)}, nil
}

func (c *ClientSecretCredentialWithPoP) NeedAuthenticate() bool {
	return false
}

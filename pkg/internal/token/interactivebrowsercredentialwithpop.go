package token

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/kubelogin/pkg/internal/pop"
)

type InteractiveBrowserCredentialWithPoP struct {
	popClaims map[string]string
	options   *pop.MsalClientOptions
}

var _ CredentialProvider = (*InteractiveBrowserCredentialWithPoP)(nil)

func newInteractiveBrowserCredentialWithPoP(opts *Options) (CredentialProvider, error) {
	if opts.ClientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}
	if opts.TenantID == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	popClaimsMap, err := parsePoPClaims(opts.PoPTokenClaims)
	if err != nil {
		return nil, fmt.Errorf("unable to parse PoP claims: %w", err)
	}
	if len(popClaimsMap) == 0 {
		return nil, fmt.Errorf("number of pop claims is invalid: %d", len(popClaimsMap))
	}
	msalOpts := &pop.MsalClientOptions{
		Authority:                opts.GetCloudConfiguration().ActiveDirectoryAuthorityHost,
		ClientID:                 opts.ClientID,
		TenantID:                 opts.TenantID,
		DisableInstanceDiscovery: opts.DisableInstanceDiscovery,
	}
	if opts.httpClient != nil {
		msalOpts.Options.Transport = opts.httpClient
	}
	return &InteractiveBrowserCredentialWithPoP{
		options:   msalOpts,
		popClaims: popClaimsMap,
	}, nil
}

func (c *InteractiveBrowserCredentialWithPoP) Name() string {
	return "InteractiveBrowserCredentialWithPoP"
}

func (c *InteractiveBrowserCredentialWithPoP) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	return azidentity.AuthenticationRecord{}, errAuthenticateNotSupported
}

func (c *InteractiveBrowserCredentialWithPoP) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	token, expirationTimeUnix, err := pop.AcquirePoPTokenInteractive(
		ctx,
		c.popClaims,
		opts.Scopes,
		c.options,
	)
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("failed to create PoP token using interactive login: %w", err)
	}
	return azcore.AccessToken{Token: token, ExpiresOn: time.Unix(expirationTimeUnix, 0)}, nil
}

func (c *InteractiveBrowserCredentialWithPoP) NeedAuthenticate() bool {
	return false
}

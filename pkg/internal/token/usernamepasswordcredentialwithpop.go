package token

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/kubelogin/pkg/internal/pop"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

type UsernamePasswordCredentialWithPoP struct {
	popClaims         map[string]string
	username          string
	password          string
	client            public.Client
	options           *pop.MsalClientOptions
	cacheDir          string
	usePersistentKeys bool
}

var _ CredentialProvider = (*UsernamePasswordCredentialWithPoP)(nil)

func newUsernamePasswordCredentialWithPoP(opts *Options, cache cache.ExportReplace) (CredentialProvider, error) {
	if opts.ClientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}
	if opts.TenantID == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	if opts.Username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if opts.Password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}
	popClaimsMap, err := parsePoPClaims(opts.PoPTokenClaims)
	if err != nil {
		return nil, fmt.Errorf("unable to parse PoP claims: %w", err)
	}
	if len(popClaimsMap) == 0 {
		return nil, fmt.Errorf("number of pop claims is invalid: %d", len(popClaimsMap))
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
	client, err := pop.NewPublicClient(msalOpts, pop.WithCustomCachePublic(cache))
	if err != nil {
		return nil, fmt.Errorf("unable to create public client: %w", err)
	}
	// Only set cacheDir and use persistent keys when cache is available
	var cacheDir string
	usePersistentPoPKeys := false
	if cache != nil {
		cacheDir = opts.AuthRecordCacheDir
		usePersistentPoPKeys = true
	}

	return &UsernamePasswordCredentialWithPoP{
		options:           msalOpts,
		popClaims:         popClaimsMap,
		username:          opts.Username,
		password:          opts.Password,
		client:            client,
		cacheDir:          cacheDir,
		usePersistentKeys: usePersistentPoPKeys,
	}, nil
}

func (c *UsernamePasswordCredentialWithPoP) Name() string {
	return "UsernamePasswordCredentialWithPoP"
}

func (c *UsernamePasswordCredentialWithPoP) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	return azidentity.AuthenticationRecord{}, errAuthenticateNotSupported
}

func (c *UsernamePasswordCredentialWithPoP) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	var popKey *pop.SwKey
	var err error

	if c.usePersistentKeys {
		// Use persistent key storage when caching is available
		popKey, err = pop.GetSwPoPKeyPersistent(c.cacheDir)
		if err != nil {
			return azcore.AccessToken{}, fmt.Errorf("unable to get persistent PoP key: %w", err)
		}
	} else {
		// Use ephemeral keys when no caching is available
		popKey, err = pop.GetSwPoPKey()
		if err != nil {
			return azcore.AccessToken{}, fmt.Errorf("unable to generate PoP key: %w", err)
		}
	}

	token, expirationTimeUnix, err := pop.AcquirePoPTokenByUsernamePassword(
		ctx,
		c.popClaims,
		opts.Scopes,
		c.client,
		c.username,
		c.password,
		c.options,
		popKey,
	)
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("failed to create PoP token using username and password credential: %w", err)
	}
	return azcore.AccessToken{Token: token, ExpiresOn: time.Unix(expirationTimeUnix, 0)}, nil
}

func (c *UsernamePasswordCredentialWithPoP) NeedAuthenticate() bool {
	return false
}

package token

//go:generate sh -c "mockgen -destination mock_$GOPACKAGE/execCredentialPlugin.go github.com/Azure/kubelogin/pkg/internal/token ExecCredentialPlugin"

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Azure/go-autorest/autorest/adal"
	klog "k8s.io/klog/v2"
)

const (
	expirationDelta time.Duration = 60 * time.Second
)

type ExecCredentialPlugin interface {
	Do(ctx context.Context) error
}

type execCredentialPlugin struct {
	o                    *Options
	tokenCache           TokenCache
	execCredentialWriter ExecCredentialWriter
	provider             TokenProvider
	disableTokenCache    bool
	refresher            func(adal.OAuthConfig, string, string, string, *adal.Token) (TokenProvider, error)
}

func New(o *Options) (ExecCredentialPlugin, error) {

	klog.V(10).Info(o.ToString())
	provider, err := NewTokenProvider(o)
	if err != nil {
		return nil, err
	}
	disableTokenCache := false
	if o.LoginMethod == ServicePrincipalLogin || o.LoginMethod == MSILogin || o.LoginMethod == WorkloadIdentityLogin || o.LoginMethod == AzureCLILogin {
		disableTokenCache = true
	}
	return &execCredentialPlugin{
		o:                    o,
		tokenCache:           &defaultTokenCache{},
		execCredentialWriter: &execCredentialWriter{},
		provider:             provider,
		refresher:            newManualToken,
		disableTokenCache:    disableTokenCache,
	}, nil
}

func (p *execCredentialPlugin) Do(ctx context.Context) error {
	var (
		token adal.Token
		err   error
	)
	if !p.disableTokenCache {
		// get token from cache
		token, err = p.tokenCache.Read(p.o.tokenCacheFile)
		if err != nil {
			return fmt.Errorf("unable to read from token cache: %s, err: %s", p.o.tokenCacheFile, err)
		}
	}

	// verify resource
	targetAudience := p.o.ServerID
	if p.o.IsLegacy {
		targetAudience = fmt.Sprintf("spn:%s", p.o.ServerID)
	}
	if token.Resource == targetAudience && !token.IsZero() {
		// if not expired, return
		if !token.WillExpireIn(expirationDelta) {
			klog.V(10).Info("access token is still valid. will return")
			return p.execCredentialWriter.Write(token, os.Stdout)
		}

		// if expired, try refresh when refresh token exists
		if token.RefreshToken != "" {
			tokenRefreshed := false
			klog.V(10).Info("getting refresher")
			oAuthConfig, err := getOAuthConfig(p.o.Environment, p.o.TenantID, p.o.IsLegacy)
			if err != nil {
				return fmt.Errorf("unable to get oAuthConfig: %s", err)
			}
			refresher, err := p.refresher(*oAuthConfig, p.o.ClientID, p.o.ServerID, p.o.TenantID, &token)
			if err != nil {
				return fmt.Errorf("failed to get refresher: %s", err)
			}
			klog.V(5).Info("refresh token")
			token, err := refresher.Token(ctx)
			// if refresh fails, we will login using token provider
			if err != nil {
				klog.V(5).Infof("refresh failed, will continue to login: %s", err)
			} else {
				tokenRefreshed = true
			}

			if tokenRefreshed {
				klog.V(10).Info("token refreshed")

				// if refresh succeeds, save token, and return
				if err := p.tokenCache.Write(p.o.tokenCacheFile, token); err != nil {
					return fmt.Errorf("failed to write to store: %s", err)
				}

				return p.execCredentialWriter.Write(token, os.Stdout)
			}
		} else {
			klog.V(5).Info("there is no refresh token")
		}
	}

	klog.V(5).Info("acquire new token")
	// run the underlying provider
	token, err = p.provider.Token(ctx)
	if err != nil {
		return fmt.Errorf("failed to get token: %s", err)
	}

	if !p.disableTokenCache {
		// save token
		if err := p.tokenCache.Write(p.o.tokenCacheFile, token); err != nil {
			return fmt.Errorf("unable to write to token cache: %s, err: %s", p.o.tokenCacheFile, err)
		}
	}

	return p.execCredentialWriter.Write(token, os.Stdout)
}

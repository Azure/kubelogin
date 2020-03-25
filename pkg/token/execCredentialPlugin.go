package token

import (
	"fmt"
	"time"

	"github.com/Azure/go-autorest/autorest/adal"
	"k8s.io/klog"
)

const (
	expirationDelta time.Duration = 60 * time.Second
)

type ExecCredentialPlugin interface {
	Do() error
}

type execCredentialPlugin struct {
	o                    *Options
	tokenCache           TokenCache
	execCredentialWriter ExecCredentialWriter
	provider             TokenProvider
	refresher            func(adal.OAuthConfig, string, string, string, *adal.Token) (TokenProvider, error)
}

func New(o *Options) (ExecCredentialPlugin, error) {
	provider, err := newTokenProvider(o, nil)
	if err != nil {
		return nil, err
	}
	return &execCredentialPlugin{
		o:                    o,
		tokenCache:           &defaultTokenCache{},
		execCredentialWriter: &execCredentialWriter{},
		provider:             provider,
		refresher:            newManualToken,
	}, nil
}

func (p *execCredentialPlugin) Do() error {
	// get token from cache
	token, err := p.tokenCache.Read(p.o.TokenCacheFile)
	if err != nil {
		return fmt.Errorf("unable to read from token cache: %s, err: %s", p.o.TokenCacheFile, err)
	}

	// verify resource
	if token.Resource == p.o.ServerID && !token.IsZero() {
		// if not expired, return
		if !token.WillExpireIn(expirationDelta) {
			klog.V(10).Info("access token is still valid. will return")
			return p.execCredentialWriter.Write(token)
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
			token, err := refresher.Token()
			// if refresh fails, we will login using token provider
			if err != nil {
				klog.V(5).Infof("refresh failed, will continue to login: %s", err)
			} else {
				tokenRefreshed = true
			}

			if tokenRefreshed {
				klog.V(10).Info("token refreshed")

				// if refresh succeeds, save tooken, and return
				if err := p.tokenCache.Write(p.o.TokenCacheFile, token); err != nil {
					return fmt.Errorf("failed to write to store: %s", err)
				}

				return p.execCredentialWriter.Write(token)
			}
		} else {
			klog.V(5).Info("there is no refresh token")
		}
	}

	klog.V(5).Info("acquire new token")
	// run the underlying provider
	token, err = p.provider.Token()
	if err != nil {
		return fmt.Errorf("failed to get token: %s", err)
	}

	// save token
	if err := p.tokenCache.Write(p.o.TokenCacheFile, token); err != nil {
		return fmt.Errorf("unable to write to token cache: %s, err: %s", p.o.TokenCacheFile, err)
	}

	return p.execCredentialWriter.Write(token)
}

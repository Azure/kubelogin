package token

//go:generate sh -c "mockgen -destination mock_$GOPACKAGE/execCredentialPlugin.go github.com/Azure/kubelogin/pkg/internal/token ExecCredentialPlugin"

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	klog "k8s.io/klog/v2"
)

type ExecCredentialPlugin interface {
	Do(ctx context.Context) error
}

type execCredentialPlugin struct {
	o                    *Options
	cachedRecord         CachedRecordProvider
	execCredentialWriter ExecCredentialWriter
	newCredentialFunc    func(record azidentity.AuthenticationRecord, o *Options) (CredentialProvider, error)
}

var errAuthenticateNotSupported = errors.New("authenticate is not supported")

func New(o *Options) (ExecCredentialPlugin, error) {
	klog.V(10).Info(o.ToString())
	return &execCredentialPlugin{
		o:                    o,
		execCredentialWriter: &execCredentialWriter{},
		cachedRecord: &defaultCachedRecordProvider{
			file: o.authRecordCacheFile,
		},
		newCredentialFunc: NewAzIdentityCredential,
	}, nil
}

func (p *execCredentialPlugin) Do(ctx context.Context) error {
	if p.o.ServerID == "" {
		return errors.New("server-id is required")
	}

	ctx, cancel := context.WithTimeout(ctx, p.o.Timeout)
	defer cancel()

	record, err := p.cachedRecord.Retrieve()
	if err != nil {
		klog.V(5).Infof("failed to retrieve cached record: %s", err)
	}

	cred, err := p.newCredentialFunc(record, p.o)
	if err != nil {
		return fmt.Errorf("failed to create azidentity credential: %w", err)
	}

	klog.V(5).Infof("using credential: %s", cred.Name())
	scopes := []string{GetScope(p.o.ServerID)}
	tokenRequestOptions := policy.TokenRequestOptions{
		TenantID: p.o.TenantID,
		Scopes:   scopes,
	}

	if cred.NeedAuthenticate() && record == (azidentity.AuthenticationRecord{}) {
		// No stored record; call Authenticate to acquire one.
		// This will prompt the user to authenticate interactively.
		klog.V(5).Info("no stored record; calling Authenticate")
		record, err = cred.Authenticate(ctx, &tokenRequestOptions)
		if err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
		err = p.cachedRecord.Store(record)
		if err != nil {
			return fmt.Errorf("failed to store record: %w", err)
		}
	}
	klog.V(5).Infof("getting token with scopes: %v", scopes)
	token, err := cred.GetToken(ctx, tokenRequestOptions)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	return p.execCredentialWriter.Write(token, os.Stdout)
}

func GetScope(serverID string) string {
	scope := strings.TrimRight(serverID, "/")
	if !strings.HasSuffix(scope, defaultScope) {
		scope += defaultScope
	}
	return scope
}

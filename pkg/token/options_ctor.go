package token

import (
	"os"

	"github.com/Azure/kubelogin/pkg/internal/env"
	"github.com/Azure/kubelogin/pkg/internal/token"
)

// OptionsWithEnv loads options from environment variables.
func OptionsWithEnv() *Options {
	// initial default values
	rv := &Options{
		LoginMethod:        os.Getenv(env.LoginMethod),
		TenantID:           os.Getenv(env.AzureTenantID),
		ClientID:           os.Getenv(env.KubeloginClientID),
		ClientSecret:       os.Getenv(env.KubeloginClientSecret),
		ClientCert:         os.Getenv(env.KubeloginClientCertificatePath),
		ClientCertPassword: os.Getenv(env.KubeloginClientCertificatePassword),
		AuthorityHost:      os.Getenv(env.AzureAuthorityHost),
		FederatedTokenFile: os.Getenv(env.AzureFederatedTokenFile),
	}

	// azure variant overrides
	if v, ok := os.LookupEnv(env.AzureClientID); ok {
		rv.ClientID = v
	}
	if v, ok := os.LookupEnv(env.AzureClientSecret); ok {
		rv.ClientSecret = v
	}
	if v, ok := os.LookupEnv(env.AzureClientCertificatePath); ok {
		rv.ClientCert = v
	}
	if v, ok := os.LookupEnv(env.AzureClientCertificatePassword); ok {
		rv.ClientCertPassword = v
	}

	return rv
}

func (opts *Options) toInternalOptions() *token.Options {
	return &token.Options{
		LoginMethod:        opts.LoginMethod,
		Environment:        opts.Environment,
		TenantID:           opts.TenantID,
		ServerID:           opts.ServerID,
		ClientID:           opts.ClientID,
		ClientSecret:       opts.ClientSecret,
		ClientCert:         opts.ClientCert,
		ClientCertPassword: opts.ClientCertPassword,
		IsPoPTokenEnabled:  opts.IsPoPTokenEnabled,
		PoPTokenClaims:     opts.PoPTokenClaims,
		IdentityResourceID: opts.IdentityResourceID,
		AuthorityHost:      opts.AuthorityHost,
		FederatedTokenFile: opts.FederatedTokenFile,
	}
}

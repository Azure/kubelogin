package token

import (
	"os"

	"github.com/Azure/kubelogin/pkg/internal/env"
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
		Username:           os.Getenv(env.KubeloginROPCUsername),
		Password:           os.Getenv(env.KubeloginROPCPassword),
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
	if v, ok := os.LookupEnv(env.AzureUsername); ok {
		rv.Username = v
	}
	if v, ok := os.LookupEnv(env.AzurePassword); ok {
		rv.Password = v
	}

	return rv
}

package token

import (
	"testing"

	"github.com/Azure/kubelogin/pkg/internal/env"
	"github.com/stretchr/testify/assert"
)

func TestOptionsWithEnv(t *testing.T) {
	t.Run("no env vars", func(t *testing.T) {
		o := OptionsWithEnv()
		assert.Equal(t, &Options{}, o)
	})

	t.Run("with kubelogin variant env vars", func(t *testing.T) {
		for k, v := range map[string]string{
			env.LoginMethod:                        MSILogin,
			env.AzureTenantID:                      "tenant-id",
			env.KubeloginClientID:                  "client-id",
			env.KubeloginClientSecret:              "client-secret",
			env.KubeloginClientCertificatePath:     "client-cert-path",
			env.KubeloginClientCertificatePassword: "client-cert-password",
			env.KubeloginROPCUsername:              "username",
			env.KubeloginROPCPassword:              "password",
			env.AzureAuthorityHost:                 "authority-host",
			env.AzureFederatedTokenFile:            "federated-token-file",
		} {
			t.Setenv(k, v)
		}

		o := OptionsWithEnv()
		assert.Equal(t, &Options{
			LoginMethod:        MSILogin,
			TenantID:           "tenant-id",
			ClientID:           "client-id",
			ClientSecret:       "client-secret",
			ClientCert:         "client-cert-path",
			ClientCertPassword: "client-cert-password",
			Username:           "username",
			Password:           "password",
			AuthorityHost:      "authority-host",
			FederatedTokenFile: "federated-token-file",
		}, o)
	})

	t.Run("with azure variant env vars", func(t *testing.T) {
		for k, v := range map[string]string{
			env.LoginMethod:                        MSILogin,
			env.AzureTenantID:                      "tenant-id",
			env.KubeloginClientID:                  "client-id",
			env.AzureClientID:                      "azure-client-id",
			env.KubeloginClientSecret:              "client-secret",
			env.AzureClientSecret:                  "azure-client-secret",
			env.KubeloginClientCertificatePath:     "client-cert-path",
			env.AzureClientCertificatePath:         "azure-client-cert-path",
			env.KubeloginClientCertificatePassword: "client-cert-password",
			env.AzureClientCertificatePassword:     "azure-client-cert-password",
			env.KubeloginROPCUsername:              "username",
			env.AzureUsername:                      "azure-username",
			env.KubeloginROPCPassword:              "password",
			env.AzurePassword:                      "azure-password",
			env.AzureAuthorityHost:                 "authority-host",
			env.AzureFederatedTokenFile:            "federated-token-file",
		} {
			t.Setenv(k, v)
		}

		o := OptionsWithEnv()
		assert.Equal(t, &Options{
			LoginMethod:        MSILogin,
			TenantID:           "tenant-id",
			ClientID:           "azure-client-id",
			ClientSecret:       "azure-client-secret",
			ClientCert:         "azure-client-cert-path",
			ClientCertPassword: "azure-client-cert-password",
			Username:           "azure-username",
			Password:           "azure-password",
			AuthorityHost:      "authority-host",
			FederatedTokenFile: "federated-token-file",
		}, o)
	})
}

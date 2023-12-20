package token

import (
	"reflect"
	"testing"

	"github.com/Azure/kubelogin/pkg/internal/env"
	"github.com/Azure/kubelogin/pkg/internal/token"
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
			AuthorityHost:      "authority-host",
			FederatedTokenFile: "federated-token-file",
		}, o)
	})
}

func TestOptions_toInternalOptions(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		o := &Options{
			LoginMethod:        "login-method",
			Environment:        "environment",
			TenantID:           "tenant-id",
			ServerID:           "server-id",
			ClientID:           "client-id",
			ClientSecret:       "client-secret",
			ClientCert:         "client-cert",
			ClientCertPassword: "client-cert-password",
			IsPoPTokenEnabled:  true,
			PoPTokenClaims:     "pop-token-claims",
			IdentityResourceID: "identity-resource-id",
			AuthorityHost:      "authority-host",
			FederatedTokenFile: "federated-token-file",
		}
		assert.Equal(t, &token.Options{
			LoginMethod:        "login-method",
			Environment:        "environment",
			TenantID:           "tenant-id",
			ServerID:           "server-id",
			ClientID:           "client-id",
			ClientSecret:       "client-secret",
			ClientCert:         "client-cert",
			ClientCertPassword: "client-cert-password",
			IsPoPTokenEnabled:  true,
			PoPTokenClaims:     "pop-token-claims",
			IdentityResourceID: "identity-resource-id",
			AuthorityHost:      "authority-host",
			FederatedTokenFile: "federated-token-file",
		}, o.toInternalOptions())
	})

	// this test uses reflection to ensure all fields in *Options
	// are copied to *token.Options without modification.
	t.Run("fields assignment", func(t *testing.T) {
		boolValue := true
		stringValue := "string-value"

		o := &Options{}

		// fill up all fields in *Options
		oType := reflect.TypeOf(o).Elem()
		oValue := reflect.ValueOf(o).Elem()
		for i := 0; i < oValue.NumField(); i++ {
			fieldValue := oValue.Field(i)
			fieldType := oType.Field(i)
			switch k := fieldType.Type.Kind(); k {
			case reflect.Bool:
				// set bool value
				fieldValue.SetBool(boolValue)
			case reflect.String:
				fieldValue.SetString(stringValue)
			default:
				t.Errorf("unexpected type: %s", k)
			}
		}

		internalOpts := o.toInternalOptions()
		assert.NotNil(t, internalOpts)

		internalOptsValue := reflect.ValueOf(internalOpts).Elem()
		for i := 0; i < oValue.NumField(); i++ {
			fieldType := oType.Field(i)
			t.Log(fieldType.Name)
			internalOptsFieldValue := internalOptsValue.FieldByName(fieldType.Name)
			switch k := fieldType.Type.Kind(); k {
			case reflect.Bool:
				assert.Equal(t, boolValue, internalOptsFieldValue.Bool(), "field: %s", fieldType.Name)
			case reflect.String:
				assert.Equal(t, stringValue, internalOptsFieldValue.String(), "field: %s", fieldType.Name)
			default:
				t.Errorf("unexpected type: %s", k)
			}
		}
	})
}

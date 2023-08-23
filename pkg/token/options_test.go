package token

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/pflag"
)

func TestOptions(t *testing.T) {
	t.Run("Default option should produce token cache file under default token cache directory", func(t *testing.T) {
		o := NewOptions()
		o.AddFlags(&pflag.FlagSet{})
		o.UpdateFromEnv()
		if err := o.Validate(); err != nil {
			t.Fatalf("option validation failed: %s", err)
		}
		dir, _ := filepath.Split(o.tokenCacheFile)
		if dir != DefaultTokenCacheDir {
			t.Fatalf("token cache directory is expected to be %s, got %s", DefaultTokenCacheDir, dir)
		}
	})

	t.Run("option with customized token cache dir should produce token cache file under specified token cache directory", func(t *testing.T) {
		o := NewOptions()
		o.TokenCacheDir = "/tmp/foo/"
		o.AddFlags(&pflag.FlagSet{})
		o.UpdateFromEnv()
		if err := o.Validate(); err != nil {
			t.Fatalf("option validation failed: %s", err)
		}
		dir, _ := filepath.Split(o.tokenCacheFile)
		if dir != o.TokenCacheDir {
			t.Fatalf("token cache directory is expected to be %s, got %s", o.TokenCacheDir, dir)
		}
	})

	t.Run("invalid login method should return error", func(t *testing.T) {
		o := NewOptions()
		o.LoginMethod = "unsupported"
		if err := o.Validate(); err == nil || !strings.Contains(err.Error(), "is not a supported login method") {
			t.Fatalf("unsupported login method should return unsupported error. got: %s", err)
		}
	})

	t.Run("pop-enabled flag should return error if pop-claims are not provided", func(t *testing.T) {
		o := NewOptions()
		o.IsPoPTokenEnabled = true
		if err := o.Validate(); err == nil || !strings.Contains(err.Error(), "please provide the pop-claims flag") {
			t.Fatalf("pop-enabled with no pop claims should return missing pop-claims error. got: %s", err)
		}
	})

	t.Run("pop-claims flag should return error if pop-enabled is not provided", func(t *testing.T) {
		o := NewOptions()
		o.PoPTokenClaims = "u=testhost"
		if err := o.Validate(); err == nil || !strings.Contains(err.Error(), "pop-enabled flag is required to use the PoP token feature") {
			t.Fatalf("pop-claims provided with no pop-enabled flag should return missing pop-enabled error. got: %s", err)
		}
	})
}

func TestOptionsWithEnvVars(t *testing.T) {
	const (
		clientID      = "clientID"
		clientSecret  = "clientSecret"
		certPath      = "certPath"
		certPassword  = "password"
		username      = "username"
		password      = "password"
		tenantID      = "tenantID"
		tokenFile     = "tokenFile"
		authorityHost = "authorityHost"
	)
	testCases := []struct {
		name        string
		envVarMap   map[string]string
		isTerraform bool
		expected    Options
	}{
		{
			name: "setting env var using legacy env var format",
			envVarMap: map[string]string{
				kubeloginClientID:                  clientID,
				kubeloginClientSecret:              clientSecret,
				kubeloginClientCertificatePath:     certPath,
				kubeloginClientCertificatePassword: certPassword,
				kubeloginROPCUsername:              username,
				kubeloginROPCPassword:              password,
				azureTenantID:                      tenantID,
				loginMethod:                        DeviceCodeLogin,
			},
			expected: Options{
				ClientID:           clientID,
				ClientSecret:       clientSecret,
				ClientCert:         certPath,
				ClientCertPassword: certPassword,
				Username:           username,
				Password:           password,
				TenantID:           tenantID,
				LoginMethod:        DeviceCodeLogin,
				tokenCacheFile:     "---.json",
			},
		},
		{
			name:        "setting env var using terraform env var format",
			isTerraform: true,
			envVarMap: map[string]string{
				terraformClientID:                  clientID,
				terraformClientSecret:              clientSecret,
				terraformClientCertificatePath:     certPath,
				terraformClientCertificatePassword: certPassword,
				terraformTenantID:                  tenantID,
				loginMethod:                        DeviceCodeLogin,
			},
			expected: Options{
				UseAzureRMTerraformEnv: true,
				ClientID:               clientID,
				ClientSecret:           clientSecret,
				ClientCert:             certPath,
				ClientCertPassword:     certPassword,
				TenantID:               tenantID,
				LoginMethod:            DeviceCodeLogin,
				tokenCacheFile:         "---.json",
			},
		},
		{
			name: "setting env var using azure sdk env var format",
			envVarMap: map[string]string{
				azureClientID:                  clientID,
				azureClientSecret:              clientSecret,
				azureClientCertificatePath:     certPath,
				azureClientCertificatePassword: certPassword,
				azureUsername:                  username,
				azurePassword:                  password,
				azureTenantID:                  tenantID,
				loginMethod:                    WorkloadIdentityLogin,
				azureFederatedTokenFile:        tokenFile,
				azureAuthorityHost:             authorityHost,
			},
			expected: Options{
				ClientID:           clientID,
				ClientSecret:       clientSecret,
				ClientCert:         certPath,
				ClientCertPassword: certPassword,
				Username:           username,
				Password:           password,
				TenantID:           tenantID,
				LoginMethod:        WorkloadIdentityLogin,
				AuthorityHost:      authorityHost,
				FederatedTokenFile: tokenFile,
				tokenCacheFile:     "---.json",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVarMap {
				t.Setenv(k, v)
			}
			o := Options{}
			if tc.isTerraform {
				o.UseAzureRMTerraformEnv = true
			}
			o.AddFlags(&pflag.FlagSet{})
			o.UpdateFromEnv()
			if !cmp.Equal(o, tc.expected, cmp.AllowUnexported(Options{})) {
				t.Fatalf("expected option: %+v, got %+v", tc.expected, o)
			}
		})
	}
}

func TestParsePoPClaims(t *testing.T) {
	t.Run("pop-claim parsing should fail on empty string", func(t *testing.T) {
		popClaims := ""
		if _, err := parsePopClaims(popClaims); err == nil || !strings.Contains(err.Error(), "no claims provided") {
			t.Fatalf("parsing pop claims should return error if claims is an empty string. got: %s", err)
		}
	})

	t.Run("pop-claim parsing should fail on whitespace-only string", func(t *testing.T) {
		popClaims := "	    "
		if _, err := parsePopClaims(popClaims); err == nil || !strings.Contains(err.Error(), "no claims provided") {
			t.Fatalf("parsing pop claims should return error if claims is whitespace-only. got: %s", err)
		}
	})

	t.Run("pop-claim parsing should fail if claims are not provided in key=value format", func(t *testing.T) {
		popClaims := "claim1=val1,claim2"
		if _, err := parsePopClaims(popClaims); err == nil || !strings.Contains(err.Error(), "Ensure the claims are formatted as `key=value`") {
			t.Fatalf("parsing pop claims should return error if claims are not provided in key=value format. got: %s", err)
		}
	})

	t.Run("pop-claim parsing should fail if claims are malformed", func(t *testing.T) {
		popClaims := "claim1=  "
		if _, err := parsePopClaims(popClaims); err == nil || !strings.Contains(err.Error(), "Ensure the claims are formatted as `key=value`") {
			t.Fatalf("parsing pop claims should return error if claims are malformed. got: %s", err)
		}
	})

	t.Run("pop-claim parsing should fail if u-claim is not provided", func(t *testing.T) {
		popClaims := "claim1=val1, claim2=val2"
		if _, err := parsePopClaims(popClaims); err == nil || !strings.Contains(err.Error(), "required u-claim not provided") {
			t.Fatalf("parsing pop claims should return error if u-claim is not provided. got: %s", err)
		}
	})

	t.Run("pop-claim parsing should succeed with u-claim and additional claims", func(t *testing.T) {
		popClaims := "u=val1, claim2=val2, claim3=val3"
		if _, err := parsePopClaims(popClaims); err != nil {
			t.Fatalf("parsing pop claims should return successfully on valid claims. got: %s", err)
		}
	})
}

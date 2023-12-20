package token

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Azure/kubelogin/pkg/internal/env"
	"github.com/Azure/kubelogin/pkg/internal/testutils"
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
				env.KubeloginClientID:                  clientID,
				env.KubeloginClientSecret:              clientSecret,
				env.KubeloginClientCertificatePath:     certPath,
				env.KubeloginClientCertificatePassword: certPassword,
				env.KubeloginROPCUsername:              username,
				env.KubeloginROPCPassword:              password,
				env.AzureTenantID:                      tenantID,
				env.LoginMethod:                        DeviceCodeLogin,
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
				Timeout:            30 * time.Second,
			},
		},
		{
			name:        "setting env var using terraform env var format",
			isTerraform: true,
			envVarMap: map[string]string{
				env.TerraformClientID:                  clientID,
				env.TerraformClientSecret:              clientSecret,
				env.TerraformClientCertificatePath:     certPath,
				env.TerraformClientCertificatePassword: certPassword,
				env.TerraformTenantID:                  tenantID,
				env.LoginMethod:                        DeviceCodeLogin,
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
				Timeout:                30 * time.Second,
			},
		},
		{
			name: "setting env var using azure sdk env var format",
			envVarMap: map[string]string{
				env.AzureClientID:                  clientID,
				env.AzureClientSecret:              clientSecret,
				env.AzureClientCertificatePath:     certPath,
				env.AzureClientCertificatePassword: certPassword,
				env.AzureUsername:                  username,
				env.AzurePassword:                  password,
				env.AzureTenantID:                  tenantID,
				env.LoginMethod:                    WorkloadIdentityLogin,
				env.AzureFederatedTokenFile:        tokenFile,
				env.AzureAuthorityHost:             authorityHost,
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
				Timeout:            30 * time.Second,
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
	testCases := []struct {
		name           string
		popClaims      string
		expectedError  error
		expectedClaims map[string]string
	}{
		{
			name:           "pop-claim parsing should fail on empty string",
			popClaims:      "",
			expectedError:  fmt.Errorf("failed to parse PoP token claims: no claims provided"),
			expectedClaims: nil,
		},
		{
			name:           "pop-claim parsing should fail on whitespace-only string",
			popClaims:      "	    ",
			expectedError:  fmt.Errorf("failed to parse PoP token claims: no claims provided"),
			expectedClaims: nil,
		},
		{
			name:           "pop-claim parsing should fail if claims are not provided in key=value format",
			popClaims:      "claim1=val1,claim2",
			expectedError:  fmt.Errorf("failed to parse PoP token claims. Ensure the claims are formatted as `key=value` with no extra whitespace"),
			expectedClaims: nil,
		},
		{
			name:           "pop-claim parsing should fail if claims are malformed",
			popClaims:      "claim1=  ",
			expectedError:  fmt.Errorf("failed to parse PoP token claims. Ensure the claims are formatted as `key=value` with no extra whitespace"),
			expectedClaims: nil,
		},
		{
			name:           "pop-claim parsing should fail if claims are malformed/commas only",
			popClaims:      ",,,,,,,,",
			expectedError:  fmt.Errorf("failed to parse PoP token claims. Ensure the claims are formatted as `key=value` with no extra whitespace"),
			expectedClaims: nil,
		},
		{
			name:           "pop-claim parsing should fail if u-claim is not provided",
			popClaims:      "1=2,3=4",
			expectedError:  fmt.Errorf("required u-claim not provided for PoP token flow. Please provide the ARM ID of the cluster in the format `u=<ARM_ID>`"),
			expectedClaims: nil,
		},
		{
			name:          "pop-claim parsing should succeed with u-claim and additional claims",
			popClaims:     "u=val1, claim2=val2, claim3=val3",
			expectedError: nil,
			expectedClaims: map[string]string{
				"u":      "val1",
				"claim2": "val2",
				"claim3": "val3",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claimsMap, err := parsePoPClaims(tc.popClaims)
			if err != nil {
				if !testutils.ErrorContains(err, tc.expectedError.Error()) {
					t.Fatalf("expected error: %+v, got error: %+v", tc.expectedError, err)
				}
			} else {
				if err != tc.expectedError {
					t.Fatalf("expected error: %+v, got error: %+v", tc.expectedError, err)
				}
			}
			if !cmp.Equal(claimsMap, tc.expectedClaims) {
				t.Fatalf("expected claims map to be %s, got map: %s", tc.expectedClaims, claimsMap)
			}
		})
	}
}

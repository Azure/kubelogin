package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNeedAuthenticate locks in which credential types require the MSAL
// Authenticate() step before GetToken. Only the interactive / public-client
// flows (device code, interactive browser, username-password) report true;
// every other credential authenticates as part of GetToken and reports false.
//
// NeedAuthenticate is a constant predicate on each type, so a zero-value
// instance is sufficient to exercise it.
func TestNeedAuthenticate(t *testing.T) {
	testCases := []struct {
		name string
		cred interface{ NeedAuthenticate() bool }
		want bool
	}{
		{"ADALClientCertCredential", &ADALClientCertCredential{}, false},
		{"ADALClientSecretCredential", &ADALClientSecretCredential{}, false},
		{"ADALDeviceCodeCredential", &ADALDeviceCodeCredential{}, false},
		{"AzureCLICredential", &AzureCLICredential{}, false},
		{"AzureDeveloperCLICredential", &AzureDeveloperCLICredential{}, false},
		{"AzurePipelinesCredential", &AzurePipelinesCredential{}, false},
		{"ClientCertificateCredential", &ClientCertificateCredential{}, false},
		{"ClientCertificateCredentialWithPoP", &ClientCertificateCredentialWithPoP{}, false},
		{"ClientSecretCredential", &ClientSecretCredential{}, false},
		{"ClientSecretCredentialWithPoP", &ClientSecretCredentialWithPoP{}, false},
		{"DeviceCodeCredential", &DeviceCodeCredential{}, true},
		{"GithubActionsCredential", &GithubActionsCredential{}, false},
		{"InteractiveBrowserCredential", &InteractiveBrowserCredential{}, true},
		{"InteractiveBrowserCredentialWithPoP", &InteractiveBrowserCredentialWithPoP{}, false},
		{"ManagedIdentityCredential", &ManagedIdentityCredential{}, false},
		{"UsernamePasswordCredential", &UsernamePasswordCredential{}, true},
		{"UsernamePasswordCredentialWithPoP", &UsernamePasswordCredentialWithPoP{}, false},
		{"WorkloadIdentityCredential", &WorkloadIdentityCredential{}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.cred.NeedAuthenticate())
		})
	}
}

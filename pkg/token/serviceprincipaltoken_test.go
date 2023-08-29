package token

import (
	"fmt"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

const (
	clientID       = "AZURE_CLIENT_ID"
	clientSecret   = "AAD_SERVICE_PRINCIPAL_CLIENT_SECRET"
	clientCert     = "AZURE_CLIENT_CER"
	clientCertPass = "AZURE_CLIENT_CERTIFICATE_PASSWORD"
	resourceID     = "AZURE_RESOURCE_ID"
	tenantID       = "AZURE_TENANT_ID"
)

func TestMissingLoginMethods(t *testing.T) {
	p := &servicePrincipalToken{}
	expectedErr := "service principal token requires either client secret or certificate"

	_, err := p.Token()
	if !ErrorContains(err, expectedErr) {
		t.Errorf("expected error %s, but got %s", expectedErr, err)
	}
}

func TestMissingCertFile(t *testing.T) {
	p := &servicePrincipalToken{
		clientCert: "testdata/noCertHere.pfx",
	}
	expectedErr := "failed to read the certificate file"

	_, err := p.Token()
	if !ErrorContains(err, expectedErr) {
		t.Errorf("expected error %s, but got %s", expectedErr, err)
	}
}

func TestBadCertPassword(t *testing.T) {
	p := &servicePrincipalToken{
		clientCert:         "testdata/testCert.pfx",
		clientCertPassword: badSecret,
	}
	expectedErr := "failed to decode pkcs12 certificate while creating spt: pkcs12: decryption password incorrect"

	_, err := p.Token()
	if !ErrorContains(err, expectedErr) {
		t.Errorf("expected error %s, but got %s", expectedErr, err)
	}
}

func TestServicePrincipalTokenVCR(t *testing.T) {
	pEnv := &servicePrincipalToken{
		clientID:           os.Getenv(clientID),
		clientSecret:       os.Getenv(clientSecret),
		clientCert:         os.Getenv(clientCert),
		clientCertPassword: os.Getenv(clientCertPass),
		resourceID:         os.Getenv(resourceID),
		tenantID:           os.Getenv(tenantID),
	}
	// Use defaults if environmental variables are empty
	if pEnv.clientID == "" {
		pEnv.clientID = clientID
	}
	if pEnv.clientSecret == "" {
		pEnv.clientSecret = clientSecret
	}
	if pEnv.clientCert == "" {
		pEnv.clientCert = "testdata/testCert.pfx"
	}
	if pEnv.clientCertPassword == "" {
		pEnv.clientCertPassword = "TestPassword"
	}
	if pEnv.resourceID == "" {
		pEnv.resourceID = resourceID
	}
	if pEnv.tenantID == "" {
		pEnv.tenantID = "00000000-0000-0000-0000-000000000000"
	}
	var expectedToken string
	testCase := []struct {
		cassetteName      string
		p                 *servicePrincipalToken
		expectedError     error
		useSecret         bool
		expectedTokenType string
	}{
		{
			// Test using incorrect secret value
			cassetteName: "ServicePrincipalTokenFromBadSecretVCR",
			p: &servicePrincipalToken{
				clientID:     pEnv.clientID,
				clientSecret: badSecret,
				resourceID:   pEnv.resourceID,
				tenantID:     pEnv.tenantID,
			},
			expectedError: fmt.Errorf("ClientSecretCredential authentication failed"),
			useSecret:     true,
		},
		{
			// Test using service principal secret value to get token
			cassetteName: "ServicePrincipalTokenFromSecretVCR",
			p: &servicePrincipalToken{
				clientID:     pEnv.clientID,
				clientSecret: pEnv.clientSecret,
				resourceID:   pEnv.resourceID,
				tenantID:     pEnv.tenantID,
			},
			expectedError: nil,
			useSecret:     true,
		},
		{
			// Test using service principal certificate to get token
			cassetteName: "ServicePrincipalTokenFromCertVCR",
			p: &servicePrincipalToken{
				clientID:           pEnv.clientID,
				clientCert:         pEnv.clientCert,
				clientCertPassword: pEnv.clientCertPassword,
				resourceID:         pEnv.resourceID,
				tenantID:           pEnv.tenantID,
			},
			expectedError: nil,
			useSecret:     false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.cassetteName, func(t *testing.T) {
			if tc.expectedError == nil {
				expectedToken = uuid.New().String()
			}
			vcrRecorder, httpClient := GetVCRHttpClient(fmt.Sprintf("testdata/%s", tc.cassetteName), expectedToken)

			clientOpts := azcore.ClientOptions{
				Cloud:     cloud.AzurePublic,
				Transport: httpClient,
			}

			token, err := tc.p.TokenWithOptions(&clientOpts)
			defer vcrRecorder.Stop()
			if err != nil {
				if !ErrorContains(err, tc.expectedError.Error()) {
					t.Errorf("expected error %s, but got %s", tc.expectedError.Error(), err)
				}
			} else {
				if token.AccessToken == "" {
					t.Error("expected valid token, but received empty token.")
				}
				if vcrRecorder.Mode() == recorder.ModeReplayOnly {
					if token.AccessToken != expectedToken {
						t.Errorf("unexpected token returned (expected %s, but got %s)", expectedToken, token.AccessToken)
					}
				}
			}
		})
	}
}

func TestServicePrincipalPoPTokenVCR(t *testing.T) {
	pEnv := &servicePrincipalToken{
		clientID:           os.Getenv(clientID),
		clientSecret:       os.Getenv(clientSecret),
		clientCert:         os.Getenv(clientCert),
		clientCertPassword: os.Getenv(clientCertPass),
		resourceID:         os.Getenv(resourceID),
		tenantID:           os.Getenv(tenantID),
	}
	// Use defaults if environmental variables are empty
	if pEnv.clientID == "" {
		pEnv.clientID = clientID
	}
	if pEnv.clientSecret == "" {
		pEnv.clientSecret = clientSecret
	}
	if pEnv.clientCert == "" {
		pEnv.clientCert = "testdata/testCert.pfx"
	}
	if pEnv.clientCertPassword == "" {
		pEnv.clientCertPassword = "TestPassword"
	}
	if pEnv.resourceID == "" {
		pEnv.resourceID = resourceID
	}
	if pEnv.tenantID == "" {
		pEnv.tenantID = "00000000-0000-0000-0000-000000000000"
	}
	var expectedToken string
	var err error
	var token adal.Token
	expectedTokenType := "pop"
	testCase := []struct {
		cassetteName  string
		p             *servicePrincipalToken
		expectedError error
		useSecret     bool
	}{
		{
			// Test using malformed pop claims
			cassetteName: "ServicePrincipalPoPTokenFromBadPoPClaimsVCR",
			p: &servicePrincipalToken{
				clientID:     pEnv.clientID,
				clientSecret: badSecret,
				resourceID:   pEnv.resourceID,
				tenantID:     pEnv.tenantID,
				popClaims:    map[string]string{"1": "2"},
			},
			expectedError: fmt.Errorf("failed to create service principal PoP token using secret"),
			useSecret:     true,
		},
		{
			// Test using service principal secret value to get PoP token
			cassetteName: "ServicePrincipalPoPTokenFromSecretVCR",
			p: &servicePrincipalToken{
				clientID:     pEnv.clientID,
				clientSecret: pEnv.clientSecret,
				resourceID:   pEnv.resourceID,
				tenantID:     pEnv.tenantID,
				popClaims:    map[string]string{"u": "testhost"},
				cloud: cloud.Configuration{
					ActiveDirectoryAuthorityHost: "https://login.microsoftonline.com/AZURE_TENANT_ID",
				},
			},
			expectedError: nil,
			useSecret:     true,
		},
		{
			// Test using service principal secret value to get PoP token
			cassetteName: "ServicePrincipalPoPTokenFromCertVCR",
			p: &servicePrincipalToken{
				clientID:           pEnv.clientID,
				clientCert:         pEnv.clientCert,
				clientCertPassword: pEnv.clientCertPassword,
				resourceID:         pEnv.resourceID,
				tenantID:           pEnv.tenantID,
				popClaims:          map[string]string{"u": "testhost"},
				cloud: cloud.Configuration{
					ActiveDirectoryAuthorityHost: "https://login.microsoftonline.com/AZURE_TENANT_ID",
				},
			},
			expectedError: nil,
			useSecret:     false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.cassetteName, func(t *testing.T) {
			if tc.expectedError == nil {
				expectedToken = uuid.New().String()
			}
			vcrRecorder, httpClient := GetVCRHttpClient(fmt.Sprintf("testdata/%s", tc.cassetteName), expectedToken)

			clientOpts := azcore.ClientOptions{
				Cloud:     cloud.AzurePublic,
				Transport: httpClient,
			}

			token, err = tc.p.TokenWithOptions(&clientOpts)
			defer vcrRecorder.Stop()
			if err != nil {
				if !ErrorContains(err, tc.expectedError.Error()) {
					t.Errorf("expected error %s, but got %s", tc.expectedError.Error(), err)
				}
			} else {
				if token.AccessToken == "" {
					t.Error("expected valid token, but received empty token.")
				}
				claims := jwt.MapClaims{}
				parsed, _ := jwt.ParseWithClaims(token.AccessToken, &claims, nil)
				if vcrRecorder.Mode() == recorder.ModeReplayOnly {
					if claims["at"] != expectedToken {
						t.Errorf("unexpected token returned (expected %s, but got %s)", expectedToken, claims["at"])
					}
					if parsed.Header["typ"] != expectedTokenType {
						t.Errorf("unexpected token returned (expected %s, but got %s)", expectedTokenType, parsed.Header["typ"])
					}
				}
			}
		})
	}
}

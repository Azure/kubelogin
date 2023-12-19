package token

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/kubelogin/pkg/internal/testutils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

func TestNewServicePrincipalTokenProvider(t *testing.T) {
	testCases := []struct {
		name               string
		clientID           string
		clientSecret       string
		clientCert         string
		clientCertPassword string
		resourceID         string
		tenantID           string
		popClaims          map[string]string
		expectedError      string
	}{
		{
			name:          "test new service principal token provider with empty client ID should return error",
			expectedError: "clientID cannot be empty",
		},
		{
			name:          "test new service principal token provider with empty client secret and cert should return error",
			clientID:      "testclientid",
			expectedError: "both clientSecret and clientcert cannot be empty. One must be specified",
		},
		{
			name:          "test new service principal token provider with both client secret and cert provided should return error",
			clientID:      "testclientid",
			clientSecret:  "testsecret",
			clientCert:    "testcert",
			expectedError: "client secret and client certificate cannot be set at the same time. Only one can be specified",
		},
		{
			name:          "test new service principal token provider with empty resource ID should return error",
			clientID:      "testclientid",
			clientSecret:  "testsecret",
			expectedError: "resourceID cannot be empty",
		},
		{
			name:          "test new service principal token provider with empty tenant ID should return error",
			clientID:      "testclientid",
			clientCert:    "testcert",
			resourceID:    "testresource",
			expectedError: "tenantID cannot be empty",
		},
		{
			name:               "test new service principal token provider with cert fields should not return error",
			clientID:           "testclientid",
			clientCert:         "testcert",
			clientCertPassword: "testpass",
			resourceID:         "testresource",
			tenantID:           "testtenant",
			expectedError:      "",
		},
		{
			name:          "test new service principal token provider with secret and pop claims should not return error",
			clientID:      "testclientid",
			clientSecret:  "testsecret",
			resourceID:    "testresource",
			tenantID:      "testtenant",
			popClaims:     map[string]string{"u": "testhost"},
			expectedError: "",
		},
	}

	cloudConfig := cloud.Configuration{
		ActiveDirectoryAuthorityHost: "testendpoint",
	}
	var tokenProvider TokenProvider
	var err error
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenProvider, err = newServicePrincipalTokenProvider(
				cloudConfig,
				tc.clientID,
				tc.clientSecret,
				tc.clientCert,
				tc.clientCertPassword,
				tc.resourceID,
				tc.tenantID,
				tc.popClaims,
			)

			if tc.expectedError != "" {
				if !testutils.ErrorContains(err, tc.expectedError) {
					t.Errorf("expected error %s, but got %s", tc.expectedError, err)
				}
			} else {
				if tokenProvider == nil {
					t.Errorf("expected token provider creation to succeed, but got error: %s", err)
				}
				spnTokenProvider := tokenProvider.(*servicePrincipalToken)
				if spnTokenProvider.clientID != tc.clientID {
					t.Errorf("expected client ID: %s but got: %s", tc.clientID, spnTokenProvider.clientID)
				}
				if spnTokenProvider.clientSecret != tc.clientSecret {
					t.Errorf("expected client secret: %s but got: %s", tc.clientSecret, spnTokenProvider.clientSecret)
				}
				if spnTokenProvider.clientCert != tc.clientCert {
					t.Errorf("expected client cert: %s but got: %s", tc.clientCert, spnTokenProvider.clientCert)
				}
				if spnTokenProvider.clientCertPassword != tc.clientCertPassword {
					t.Errorf("expected client cert password: %s but got: %s", tc.clientCertPassword, spnTokenProvider.clientCertPassword)
				}
				if spnTokenProvider.resourceID != tc.resourceID {
					t.Errorf("expected resource ID: %s but got: %s", tc.resourceID, spnTokenProvider.resourceID)
				}
				if spnTokenProvider.tenantID != tc.tenantID {
					t.Errorf("expected tenant ID: %s but got: %s", tc.tenantID, spnTokenProvider.tenantID)
				}
				if !cmp.Equal(spnTokenProvider.cloud, cloudConfig) {
					t.Errorf("expected cloud config: %s but got: %s", tc.clientCertPassword, spnTokenProvider.clientCertPassword)
				}
				if !cmp.Equal(spnTokenProvider.popClaims, tc.popClaims) {
					t.Errorf("expected PoP claims: %s but got: %s", tc.popClaims, spnTokenProvider.popClaims)
				}
			}
		})
	}
}

func TestMissingLoginMethods(t *testing.T) {
	p := &servicePrincipalToken{}
	expectedErr := "service principal token requires either client secret or certificate"
	_, err := p.Token(context.TODO())
	if !testutils.ErrorContains(err, expectedErr) {
		t.Errorf("expected error %s, but got %s", expectedErr, err)
	}
}

func TestServicePrincipalTokenVCR(t *testing.T) {
	pEnv := &servicePrincipalToken{
		clientID:           os.Getenv(testutils.ClientID),
		clientSecret:       os.Getenv(testutils.ClientSecret),
		clientCert:         os.Getenv(testutils.ClientCert),
		clientCertPassword: os.Getenv(testutils.ClientCertPass),
		resourceID:         os.Getenv(testutils.ResourceID),
		tenantID:           os.Getenv(testutils.TenantID),
	}
	// Use defaults if environmental variables are empty
	if pEnv.clientID == "" {
		pEnv.clientID = testutils.ClientID
	}
	if pEnv.clientSecret == "" {
		pEnv.clientSecret = testutils.ClientSecret
	}
	if pEnv.clientCert == "" {
		pEnv.clientCert = "testdata/testCert.pfx"
	}
	if pEnv.clientCertPassword == "" {
		pEnv.clientCertPassword = "TestPassword"
	}
	if pEnv.resourceID == "" {
		pEnv.resourceID = testutils.ResourceID
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
				clientSecret: testutils.BadSecret,
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
			vcrRecorder, httpClient := testutils.GetVCRHttpClient(fmt.Sprintf("testdata/%s", tc.cassetteName), expectedToken)

			clientOpts := azcore.ClientOptions{
				Cloud:     cloud.AzurePublic,
				Transport: httpClient,
			}

			token, err := tc.p.TokenWithOptions(context.TODO(), &clientOpts)
			defer vcrRecorder.Stop()
			if err != nil {
				if !testutils.ErrorContains(err, tc.expectedError.Error()) {
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
		clientID:           os.Getenv(testutils.ClientID),
		clientSecret:       os.Getenv(testutils.ClientSecret),
		clientCert:         os.Getenv(testutils.ClientCert),
		clientCertPassword: os.Getenv(testutils.ClientCertPass),
		resourceID:         os.Getenv(testutils.ResourceID),
		tenantID:           os.Getenv(testutils.TenantID),
	}
	// Use defaults if environmental variables are empty
	if pEnv.clientID == "" {
		pEnv.clientID = testutils.ClientID
	}
	if pEnv.clientSecret == "" {
		pEnv.clientSecret = testutils.ClientSecret
	}
	if pEnv.clientCert == "" {
		pEnv.clientCert = "testdata/testCert.pfx"
	}
	if pEnv.clientCertPassword == "" {
		pEnv.clientCertPassword = "TestPassword"
	}
	if pEnv.resourceID == "" {
		pEnv.resourceID = testutils.ResourceID
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
			// Test using bad client secret
			cassetteName: "ServicePrincipalPoPTokenFromBadSecretVCR",
			p: &servicePrincipalToken{
				clientID:     pEnv.clientID,
				clientSecret: testutils.BadSecret,
				resourceID:   pEnv.resourceID,
				tenantID:     pEnv.tenantID,
				popClaims:    map[string]string{"u": "testhost"},
				cloud: cloud.Configuration{
					ActiveDirectoryAuthorityHost: "https://login.microsoftonline.com/AZURE_TENANT_ID",
				},
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
			// Test using service principal certificate to get PoP token
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
			vcrRecorder, httpClient := testutils.GetVCRHttpClient(fmt.Sprintf("testdata/%s", tc.cassetteName), expectedToken)

			clientOpts := azcore.ClientOptions{
				Cloud:     cloud.AzurePublic,
				Transport: httpClient,
			}

			token, err = tc.p.TokenWithOptions(context.TODO(), &clientOpts)
			defer vcrRecorder.Stop()
			if err != nil {
				if !testutils.ErrorContains(err, tc.expectedError.Error()) {
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

package pop

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/kubelogin/pkg/internal/testutils"
	"github.com/golang-jwt/jwt/v4"
)

type resourceOwnerTokenVars struct {
	clientID   string
	username   string
	password   string
	resourceID string
	tenantID   string
	popClaims  map[string]string
}

func TestAcquirePoPTokenByUsernamePassword(t *testing.T) {
	pEnv := &resourceOwnerTokenVars{
		clientID: os.Getenv(testutils.ClientID),
		username: os.Getenv(testutils.Username),
		password: os.Getenv(testutils.Password),
		tenantID: os.Getenv(testutils.TenantID),
	}
	// Use defaults if environmental variables are empty
	if pEnv.clientID == "" {
		pEnv.clientID = testutils.TestClientID
	}
	if pEnv.username == "" {
		pEnv.username = testutils.Username
	}
	if pEnv.password == "" {
		pEnv.password = testutils.Password
	}
	if pEnv.tenantID == "" {
		pEnv.tenantID = testutils.TestTenantID
	}

	ctx := context.Background()
	scopes := []string{testutils.TestServerID + "/.default"}
	authority := "https://login.microsoftonline.com/" + pEnv.tenantID
	var expectedToken string
	expectedTokenType := "pop"
	testCase := []struct {
		cassetteName  string
		p             *resourceOwnerTokenVars
		expectedError error
	}{
		{
			// Test using bad password
			cassetteName: "AcquirePoPTokenByUsernamePasswordFromBadPasswordVCR",
			p: &resourceOwnerTokenVars{
				clientID:   pEnv.clientID,
				username:   pEnv.username,
				password:   testutils.BadSecret,
				resourceID: testutils.TestServerID,
				tenantID:   pEnv.tenantID,
				popClaims:  map[string]string{"u": "testhost"},
			},
			expectedError: fmt.Errorf("failed to create PoP token with username/password flow"),
		},
		{
			// Test using username/password to get PoP token
			cassetteName: "AcquirePoPTokenByUsernamePasswordVCR",
			p: &resourceOwnerTokenVars{
				clientID:   pEnv.clientID,
				username:   pEnv.username,
				password:   pEnv.password,
				resourceID: testutils.TestServerID,
				tenantID:   pEnv.tenantID,
				popClaims:  map[string]string{"u": "testhost"},
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.cassetteName, func(t *testing.T) {
			if tc.expectedError == nil {
				expectedToken = testutils.TestToken
			}
			vcrRecorder, err := testutils.GetVCRHttpClient(fmt.Sprintf("testdata/%s", tc.cassetteName), pEnv.tenantID)
			if err != nil {
				t.Fatalf("failed to create vcr recorder: %s", err)
			}

			msalClientOptions := &MsalClientOptions{
				Authority: authority,
				ClientID:  tc.p.clientID,
				Options: azcore.ClientOptions{
					Cloud:     cloud.AzurePublic,
					Transport: vcrRecorder.GetDefaultClient(),
				},
				TenantID: tc.p.tenantID,
			}
			client, err := NewPublicClient(msalClientOptions)
			if err != nil {
				t.Errorf("expected no error creating client but got: %s", err)
			}

			popKey, err := GetSwPoPKeyPersistent("/tmp/test_cache")
			if err != nil {
				t.Errorf("expected no error getting PoP key but got: %s", err)
			}

			token, _, err := AcquirePoPTokenByUsernamePassword(
				ctx,
				tc.p.popClaims,
				scopes,
				client,
				tc.p.username,
				tc.p.password,
				msalClientOptions,
				popKey,
			)
			defer vcrRecorder.Stop()
			if tc.expectedError != nil {
				if !testutils.ErrorContains(err, tc.expectedError.Error()) {
					t.Errorf("expected error %s, but got %s", tc.expectedError.Error(), err)
				}
			} else if err != nil {
				t.Errorf("expected no error, but got: %s", err)
			} else {
				if token == "" {
					t.Error("expected valid token, but received empty token.")
				}
				claims := jwt.MapClaims{}
				parsed, _ := jwt.ParseWithClaims(token, &claims, nil)
				if claims["at"] != expectedToken {
					t.Errorf("unexpected token returned (expected %s, but got %s)", expectedToken, claims["at"])
				}
				if parsed.Header["typ"] != expectedTokenType {
					t.Errorf("unexpected token returned (expected %s, but got %s)", expectedTokenType, parsed.Header["typ"])
				}
			}
		})
	}
}

func TestGetPublicClient(t *testing.T) {
	httpClient := &http.Client{}
	authority := "https://login.microsoftonline.com/" + testutils.TenantID

	testCase := []struct {
		testName      string
		msalOptions   *MsalClientOptions
		expectedError error
	}{
		{
			// Test using custom HTTP transport
			testName: "TestGetPublicClientWithCustomTransport",
			msalOptions: &MsalClientOptions{
				Authority: authority,
				ClientID:  testutils.ClientID,
				Options: azcore.ClientOptions{
					Cloud:     cloud.AzurePublic,
					Transport: httpClient,
				},
				TenantID: testutils.TenantID,
			},
			expectedError: nil,
		},
		{
			// Test using default HTTP transport
			testName: "TestGetPublicClientWithDefaultTransport",
			msalOptions: &MsalClientOptions{
				Authority: authority,
				ClientID:  testutils.ClientID,
				Options: azcore.ClientOptions{
					Cloud: cloud.AzurePublic,
				},
				TenantID: testutils.TenantID,
			},
			expectedError: nil,
		},
		{
			// Test using incorrectly formatted authority
			testName: "TestGetPublicClientWithBadAuthority",
			msalOptions: &MsalClientOptions{
				Authority: "login.microsoft.com",
				ClientID:  testutils.ClientID,
				Options: azcore.ClientOptions{
					Cloud: cloud.AzurePublic,
				},
				TenantID: testutils.TenantID,
			},
			expectedError: fmt.Errorf("unable to create public client"),
		},
	}

	for _, tc := range testCase {
		t.Run(tc.testName, func(t *testing.T) {
			_, err := NewPublicClient(tc.msalOptions)
			if tc.expectedError != nil {
				if !testutils.ErrorContains(err, tc.expectedError.Error()) {
					t.Errorf("expected error %s, but got %s", tc.expectedError.Error(), err)
				}
			} else if err != nil {
				t.Errorf("expected no error creating client, but got: %s", err.Error())
			}
		})
	}
}

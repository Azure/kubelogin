package pop

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/kubelogin/pkg/internal/testutils"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

type resourceOwnerTokenVars struct {
	clientID    string
	username    string
	password    string
	resourceID  string
	tenantID    string
	popClaims   map[string]string
	oAuthConfig adal.OAuthConfig
}

func TestAcquirePoPTokenByUsernamePassword(t *testing.T) {
	pEnv := &resourceOwnerTokenVars{
		clientID:   os.Getenv(testutils.ClientID),
		username:   os.Getenv(testutils.Username),
		password:   os.Getenv(testutils.Password),
		resourceID: os.Getenv(testutils.ResourceID),
		tenantID:   os.Getenv(testutils.TenantID),
	}
	// Use defaults if environmental variables are empty
	if pEnv.clientID == "" {
		pEnv.clientID = testutils.ClientID
	}
	if pEnv.username == "" {
		pEnv.username = testutils.Username
	}
	if pEnv.password == "" {
		pEnv.password = testutils.Password
	}
	if pEnv.resourceID == "" {
		pEnv.resourceID = testutils.ResourceID
	}
	if pEnv.tenantID == "" {
		pEnv.tenantID = "00000000-0000-0000-0000-000000000000"
	}

	ctx := context.Background()
	scopes := []string{pEnv.resourceID + "/.default"}
	authority := "https://login.microsoftonline.com/" + pEnv.tenantID
	authorityEndpoint, err := url.Parse(authority)
	if err != nil {
		t.Errorf("error encountered when parsing active directory endpoint: %s", err)
	}
	var expectedToken string
	var token string
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
				resourceID: pEnv.resourceID,
				tenantID:   pEnv.tenantID,
				popClaims:  map[string]string{"u": "testhost"},
				oAuthConfig: adal.OAuthConfig{
					AuthorityEndpoint: *authorityEndpoint,
				},
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
				resourceID: pEnv.resourceID,
				tenantID:   pEnv.tenantID,
				popClaims:  map[string]string{"u": "testhost"},
				oAuthConfig: adal.OAuthConfig{
					AuthorityEndpoint: *authorityEndpoint,
				},
			},
			expectedError: nil,
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

			token, _, err = AcquirePoPTokenByUsernamePassword(
				ctx,
				tc.p.popClaims,
				scopes,
				tc.p.username,
				tc.p.password,
				&PublicClientOptions{
					Authority: authority,
					ClientID:  tc.p.clientID,
					Options:   &clientOpts,
				},
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

func TestGetPublicClient(t *testing.T) {
	httpClient := &http.Client{}
	authority := "https://login.microsoftonline.com/" + testutils.TenantID

	testCase := []struct {
		testName      string
		pcOptions     *PublicClientOptions
		expectedError error
	}{
		{
			// Test using custom HTTP transport
			testName: "TestGetPublicClientWithCustomTransport",
			pcOptions: &PublicClientOptions{
				Authority: authority,
				ClientID:  testutils.ClientID,
				Options: &azcore.ClientOptions{
					Cloud:     cloud.AzurePublic,
					Transport: httpClient,
				},
			},
			expectedError: nil,
		},
		{
			// Test using default HTTP transport
			testName: "TestGetPublicClientWithDefaultTransport",
			pcOptions: &PublicClientOptions{
				Authority: authority,
				ClientID:  testutils.ClientID,
				Options: &azcore.ClientOptions{
					Cloud: cloud.AzurePublic,
				},
			},
			expectedError: nil,
		},
		{
			// Test using incorrectly formatted authority
			testName: "TestGetPublicClientWithBadAuthority",
			pcOptions: &PublicClientOptions{
				Authority: "login.microsoft.com",
				ClientID:  testutils.ClientID,
				Options: &azcore.ClientOptions{
					Cloud: cloud.AzurePublic,
				},
			},
			expectedError: fmt.Errorf("unable to create public client"),
		},
	}

	var client *public.Client
	var err error

	for _, tc := range testCase {
		t.Run(tc.testName, func(t *testing.T) {
			client, err = getPublicClient(tc.pcOptions)

			if tc.expectedError != nil {
				if !testutils.ErrorContains(err, tc.expectedError.Error()) {
					t.Errorf("expected error %s, but got %s", tc.expectedError.Error(), err)
				}
			} else if err != nil {
				t.Errorf("expected no error, but got: %s", err)
			} else {
				if client == nil {
					t.Errorf("expected a client but got nil")
				}
			}
		})
	}
}

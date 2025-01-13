package token

import (
	"context"
	"fmt"
	"net/url"
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

func TestNewResourceOwnerTokenProvider(t *testing.T) {
	testCases := []struct {
		name          string
		clientID      string
		resourceID    string
		tenantID      string
		popClaims     map[string]string
		username      string
		password      string
		expectedError string
	}{
		{
			name:          "test new resource owner token provider with empty client ID should return error",
			expectedError: "clientID cannot be empty",
		},
		{
			name:          "test new resource owner token provider with empty username should return error",
			clientID:      "testclientid",
			expectedError: "username cannot be empty",
		},
		{
			name:          "test new resource owner token provider with empty password should return error",
			clientID:      "testclientid",
			username:      "testusername",
			expectedError: "password cannot be empty",
		},
		{
			name:          "test new resource owner token provider with empty resource ID should return error",
			clientID:      "testclientid",
			username:      "testusername",
			password:      "testpassword",
			expectedError: "resourceID cannot be empty",
		},
		{
			name:          "test new resource owner token provider with empty tenant ID should return error",
			clientID:      "testclientid",
			username:      "testusername",
			password:      "testpassword",
			resourceID:    "testresource",
			expectedError: "tenantID cannot be empty",
		},
		{
			name:          "test new resource owner token provider with no pop claims should not return error",
			clientID:      "testclientid",
			username:      "testusername",
			password:      "testpassword",
			resourceID:    "testresource",
			tenantID:      "testtenant",
			expectedError: "",
		},
		{
			name:          "test new resource owner token provider with pop claims should not return error",
			clientID:      "testclientid",
			username:      "testusername",
			password:      "testpassword",
			resourceID:    "testresource",
			tenantID:      "testtenant",
			popClaims:     map[string]string{"u": "testhost"},
			expectedError: "",
		},
	}
	var tokenProvider TokenProvider
	var err error
	oauthConfig := adal.OAuthConfig{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenProvider, err = newResourceOwnerTokenProvider(
				oauthConfig,
				tc.clientID,
				tc.username,
				tc.password,
				tc.resourceID,
				tc.tenantID,
				tc.popClaims,
				false,
			)

			if tc.expectedError != "" {
				if !testutils.ErrorContains(err, tc.expectedError) {
					t.Errorf("expected error %s, but got %s", tc.expectedError, err)
				}
			} else {
				if tokenProvider == nil {
					t.Errorf("expected token provider creation to succeed, but got error: %s", err)
				}
				rop := tokenProvider.(*resourceOwnerToken)
				if rop.clientID != tc.clientID {
					t.Errorf("expected client ID: %s but got: %s", tc.clientID, rop.clientID)
				}
				if rop.username != tc.username {
					t.Errorf("expected username: %s but got: %s", tc.username, rop.username)
				}
				if rop.password != tc.password {
					t.Errorf("expected password: %s but got: %s", tc.password, rop.password)
				}
				if rop.resourceID != tc.resourceID {
					t.Errorf("expected resource ID: %s but got: %s", tc.resourceID, rop.resourceID)
				}
				if rop.tenantID != tc.tenantID {
					t.Errorf("expected tenant ID: %s but got: %s", tc.tenantID, rop.tenantID)
				}
				if !cmp.Equal(rop.popClaims, tc.popClaims) {
					t.Errorf("expected PoP claims: %s but got: %s", tc.popClaims, rop.popClaims)
				}
			}
		})
	}
}

func TestROPCPoPTokenVCR(t *testing.T) {
	pEnv := &resourceOwnerToken{
		clientID:   os.Getenv(testutils.ClientID),
		username:   os.Getenv(testutils.Username),
		password:   os.Getenv(testutils.Password),
		resourceID: os.Getenv(testutils.ResourceID),
		tenantID:   os.Getenv(testutils.TenantID),
	}
	// Use defaults if environment variables are empty
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

	var expectedToken string
	var err error
	var token adal.Token
	expectedTokenType := "pop"
	authorityEndpoint, err := url.Parse("https://login.microsoftonline.com/" + pEnv.tenantID)
	if err != nil {
		t.Errorf("error encountered when parsing active directory endpoint: %s", err)
	}
	testCase := []struct {
		cassetteName  string
		p             *resourceOwnerToken
		expectedError error
	}{
		{
			// Test using bad password
			cassetteName: "ROPCPoPTokenFromBadPasswordVCR",
			p: &resourceOwnerToken{
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
			expectedError: fmt.Errorf("failed to create PoP token using resource owner flow"),
		},
		{
			// Test using ROPC username + password to get PoP token
			cassetteName: "ROPCPoPTokenFromUsernamePasswordVCR",
			p: &resourceOwnerToken{
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

			token, err = tc.p.tokenWithOptions(context.TODO(), &clientOpts)
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

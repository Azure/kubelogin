package token

import (
	"testing"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/kubelogin/pkg/internal/testutils"
	"github.com/google/go-cmp/cmp"
)

func TestNewInteractiveToken(t *testing.T) {
	testCases := []struct {
		name          string
		clientID      string
		resourceID    string
		tenantID      string
		popClaims     map[string]string
		expectedError string
	}{
		{
			name:          "test new interactive token provider with empty client ID should return error",
			expectedError: "clientID cannot be empty",
		},
		{
			name:          "test new interactive token provider with empty resource ID should return error",
			clientID:      "testclientid",
			expectedError: "resourceID cannot be empty",
		},
		{
			name:          "test new interactive token provider with empty tenant ID should return error",
			clientID:      "testclientid",
			resourceID:    "testresource",
			expectedError: "tenantID cannot be empty",
		},
		{
			name:          "test new interactive token provider with no pop claims should not return error",
			clientID:      "testclientid",
			resourceID:    "testresource",
			tenantID:      "testtenant",
			expectedError: "",
		},
		{
			name:          "test new interactive token provider with pop claims should not return error",
			clientID:      "testclientid",
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
			tokenProvider, err = newInteractiveTokenProvider(
				oauthConfig,
				tc.clientID,
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
				itp := tokenProvider.(*InteractiveToken)
				if itp.clientID != tc.clientID {
					t.Errorf("expected client ID: %s but got: %s", tc.clientID, itp.clientID)
				}
				if itp.resourceID != tc.resourceID {
					t.Errorf("expected resource ID: %s but got: %s", tc.resourceID, itp.resourceID)
				}
				if itp.tenantID != tc.tenantID {
					t.Errorf("expected tenant ID: %s but got: %s", tc.tenantID, itp.tenantID)
				}
				if !cmp.Equal(itp.popClaims, tc.popClaims) {
					t.Errorf("expected PoP claims: %s but got: %s", tc.popClaims, itp.popClaims)
				}
			}
		})
	}
}

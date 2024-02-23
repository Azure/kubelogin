package pop

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/kubelogin/pkg/internal/testutils"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

type confidentialTokenVars struct {
	clientID     string
	clientSecret string
	resourceID   string
	tenantID     string
	cloud        cloud.Configuration
	popClaims    map[string]string
}

func TestAcquirePoPTokenConfidential(t *testing.T) {
	pEnv := &confidentialTokenVars{
		clientID:     os.Getenv(testutils.ClientID),
		clientSecret: os.Getenv(testutils.ClientSecret),
		resourceID:   os.Getenv(testutils.ResourceID),
		tenantID:     os.Getenv(testutils.TenantID),
	}
	// Use defaults if environmental variables are empty
	if pEnv.clientID == "" {
		pEnv.clientID = testutils.ClientID
	}
	if pEnv.clientSecret == "" {
		pEnv.clientSecret = testutils.ClientSecret
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
	var expectedToken string
	var token string
	expectedTokenType := "pop"
	testCase := []struct {
		cassetteName  string
		p             *confidentialTokenVars
		expectedError error
		useSecret     bool
	}{
		{
			// Test using bad client secret
			cassetteName: "AcquirePoPTokenConfidentialFromBadSecretVCR",
			p: &confidentialTokenVars{
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
			cassetteName: "AcquirePoPTokenConfidentialWithSecretVCR",
			p: &confidentialTokenVars{
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

			cred, err := confidential.NewCredFromSecret(tc.p.clientSecret)
			if err != nil {
				t.Errorf("expected no error creating credential but got: %s", err)
			}

			token, _, err = AcquirePoPTokenConfidential(
				ctx,
				tc.p.popClaims,
				scopes,
				cred,
				authority,
				tc.p.clientID,
				tc.p.tenantID,
				&clientOpts,
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

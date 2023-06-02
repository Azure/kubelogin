package token

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

const (
	clientID          = "AZURE_CLIENT_ID"
	clientSecret      = "AAD_SERVICE_PRINCIPAL_CLIENT_SECRET"
	clientCert        = "AZURE_CLIENT_CER"
	clientCertPass    = "AZURE_CLIENT_CERTIFICATE_PASSWORD"
	resourceID        = "AZURE_RESOURCE_ID"
	tenantID          = "AZURE_TENANT_ID"
	vcrMode           = "VCR_MODE"
	vcrModeRecordOnly = "RecordOnly"
	badSecret         = "Bad_Secret"
	redactionToken    = "[REDACTED]"
)

func TestMissingLoginMethods(t *testing.T) {
	p := &servicePrincipalToken{}
	expectedErr := "service principal token requires either client secret or certificate"

	_, err := p.Token()
	if !ErrorContains(err, expectedErr) {
		t.Errorf("expected error %v, but got %v", expectedErr, err)
	}
}

func TestMissingCertFile(t *testing.T) {
	p := &servicePrincipalToken{
		clientCert: "testdata/noCertHere.pfx",
	}
	expectedErr := "failed to read the certificate file"

	_, err := p.Token()
	if !ErrorContains(err, expectedErr) {
		t.Errorf("expected error %v, but got %v", expectedErr, err)
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
		t.Errorf("expected error %v, but got %v", expectedErr, err)
	}
}

// getVCRHttpClient setup Go-vcr
func getVCRHttpClient(path string) (*recorder.Recorder, *http.Client) {
	if len(path) == 0 || path == "" {
		return nil, nil
	}

	opts := &recorder.Options{
		CassetteName: path,
		Mode:         getVCRMode(),
	}
	rec, _ := recorder.NewWithOptions(opts)
	// rec.SetRealTransport(&http.Transport{
	// 	TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	// 	// #nosec G402: PAV2 service only supports TLS12
	// 	TLSClientConfig: &tls.Config{
	// 		Renegotiation: tls.RenegotiateFreelyAsClient,
	// 		MinVersion:    tls.VersionTLS12,
	// 		MaxVersion:    tls.VersionTLS12,
	// 	},
	// })

	hook := func(i *cassette.Interaction) error {
		// Delete sensitive content
		delete(i.Response.Headers, "Set-Cookie")
		delete(i.Response.Headers, "X-Ms-Request-Id")
		if i.Request.Form["client_id"] != nil {
			i.Request.Form["client_id"] = []string{redactionToken}
		}
		if i.Request.Form["client_secret"] != nil && i.Request.Form["client_secret"][0] != badSecret {
			i.Request.Form["client_secret"] = []string{redactionToken}
		}
		if i.Request.Form["client_assertion"] != nil {
			i.Request.Form["client_assertion"] = []string{redactionToken}
		}
		if i.Request.Form["scope"] != nil {
			i.Request.Form["scope"] = []string{redactionToken + "/.default openid offline_access profile"}
		}
		i.Request.URL = strings.ReplaceAll(i.Request.URL, os.Getenv(tenantID), tenantID)
		i.Response.Body = strings.ReplaceAll(i.Response.Body, os.Getenv(tenantID), tenantID)

		if strings.Contains(i.Request.Body, "client_secret") {
			i.Request.Body = `client_id=[REDACTED]&client_secret=[REDACTED]&grant_type=client_credentials&scope=[REDACTED]%2F.default+openid+offline_access+profile`
		}

		if strings.Contains(i.Request.Body, "client_assertion") {
			i.Request.Body = `client_assertion=[REDACTED]&client_assertion_type=urn%3Aietf%3Aparams%3Aoauth%3Aclient-assertion-type%3Ajwt-bearer&client_id=[REDACTED]&client_info=1&grant_type=client_credentials&scope=[REDACTED]%2F.default+openid+offline_access+profile`
		}

		if strings.Contains(i.Response.Body, "access_token") {
			i.Response.Body = `{"token_type":"Bearer","expires_in":86399,"ext_expires_in":86399,"access_token":"[REDACTED]"}`
		}
		return nil
	}
	rec.AddHook(hook, recorder.BeforeSaveHook)

	rec.SetMatcher(customMatcher)
	rec.SetReplayableInteractions(true)

	return rec, rec.GetDefaultClient()
}

func customMatcher(r *http.Request, i cassette.Request) bool {
	id := os.Getenv(tenantID)
	if id == "" {
		id = "00000000-0000-0000-0000-000000000000"
	}
	switch os.Getenv(vcrMode) {
	case vcrModeRecordOnly:
	default:
		r.URL.Path = strings.ReplaceAll(r.URL.Path, id, tenantID)
	}
	return cassette.DefaultMatcher(r, i)
}

// Get go-vcr record mode from enviroment variable
func getVCRMode() recorder.Mode {
	switch os.Getenv(vcrMode) {
	case vcrModeRecordOnly:
		return recorder.ModeRecordOnly
	default:
		return recorder.ModeReplayOnly
	}
}

func TestServicePrincipalLoginVCR(t *testing.T) {
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

	testCase := []struct {
		cassetteName  string
		p             *servicePrincipalToken
		expectedError error
		useSecret     bool
	}{
		{
			// Test using incorrect secret value
			cassetteName: "BadSecretVCR",
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
			cassetteName: "SecretTokenVCR",
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
			cassetteName: "CertTokenVCR",
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
			vcrRecorder, httpClient := getVCRHttpClient(fmt.Sprintf("testdata/%s", tc.cassetteName))

			clientOpts := azcore.ClientOptions{
				Cloud:     cloud.AzurePublic,
				Transport: httpClient,
			}

			token, err := tc.p.TokenWithOptions(&clientOpts)
			defer vcrRecorder.Stop()
			if err != nil {
				if !ErrorContains(err, tc.expectedError.Error()) {
					t.Errorf("expected error %v, but got %v", tc.expectedError.Error(), err)
				}
			} else {
				if token.AccessToken == "" {
					t.Error("Expected valid token, but received empty token.")
				}
			}
		})
	}
}

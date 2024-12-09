package token

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testServerID   = "serverID"
	testAPIVersion = "2021-02-01"
	testTokenPath  = "/var/opt/azcmagent/tokens/fake.key"
)

func TestNewHIMDSToken(t *testing.T) {
	testCases := []struct {
		name             string
		identityEndpoint string
		expectedEndpoint string
		serverID         string
		apiVersion       string
	}{
		{
			name:             "default identity endpoint",
			identityEndpoint: "",
			expectedEndpoint: defaultIdentityEndpoint,
			serverID:         testServerID,
			apiVersion:       testAPIVersion,
		},
		{
			name:             "custom identity endpoint",
			identityEndpoint: "http://127.0.0.1:8080/metadata/identity/oauth2/token",
			expectedEndpoint: "http://127.0.0.1:8080/metadata/identity/oauth2/token",
			serverID:         testServerID,
			apiVersion:       testAPIVersion,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h, err := newHIMDSToken(tc.serverID, tc.apiVersion, tc.identityEndpoint)
			require.NoError(t, err)

			assert.Equal(t, tc.apiVersion, h.apiVersion)
			assert.Equal(t, tc.serverID, h.serverID)
			assert.Equal(t, tc.expectedEndpoint, h.identityEndpoint)
			assert.NotNil(t, h.httpClient)
		})
	}
}

func TestHIMDSToken(t *testing.T) {
	testCases := []struct {
		name          string
		challengePath string
		tokenResponse string
		svrURL        string
		expectedToken adal.Token
		expectErr     bool
	}{
		{
			name:          "valid response",
			svrURL:        "",
			challengePath: path.Join(t.TempDir(), "fake.key"),
			tokenResponse: `{"access_token":"accessToken","refresh_token":"refreshToken","expires_on":84245, "not_before":"1733696828", "resource": "serverID", "token_type":"Bearer"}`,
			expectedToken: adal.Token{
				AccessToken:  "accessToken",
				RefreshToken: "refreshToken",
				ExpiresOn:    json.Number("84245"),
				NotBefore:    json.Number("1733696828"),
				Resource:     testServerID,
				Type:         "Bearer",
			},
		},
		{
			name:          "invalid header",
			svrURL:        "",
			challengePath: "invalid",
			tokenResponse: "",
			expectedToken: adal.Token{},
			expectErr:     true,
		},
		{
			name:          "empty challenge token path",
			svrURL:        "",
			challengePath: "",
			tokenResponse: "",
			expectedToken: adal.Token{},
			expectErr:     true,
		},
		{
			name:          "client error",
			svrURL:        "http://invalid", // simulates client error
			challengePath: path.Join(t.TempDir(), "fake.key"),
			tokenResponse: `{"access_token":"accessToken","refresh_token":"refreshToken","expires_on":84245, "not_before":"1733696828", "resource": "serverID", "token_type":"Bearer"}`,
			expectedToken: adal.Token{},
			expectErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Www-Authenticate", "Basic realm="+tc.challengePath)

				if tc.tokenResponse != "" {
					w.Write([]byte(tc.tokenResponse))
				}
			}))
			defer srv.Close()

			url := srv.URL
			if tc.svrURL != "" {
				url = tc.svrURL
			}

			h := HIMDSToken{
				httpClient:       srv.Client(),
				apiVersion:       testAPIVersion,
				serverID:         testServerID,
				identityEndpoint: url,
			}

			// Write the test token in the challenge path
			if tc.challengePath != "" && tc.challengePath != "invalid" {
				err := os.WriteFile(tc.challengePath, []byte("accessToken"), 0644)
				require.NoError(t, err)
			}

			token, err := h.Token(context.Background())
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expectedToken, token)
		})
	}
}

func TestGetChallengePath(t *testing.T) {
	testCases := []struct {
		name         string
		handler      http.HandlerFunc
		servURL      string
		expectedPath string
		expectErr    bool
	}{
		{
			name: "valid response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Www-Authenticate", "Basic realm="+testTokenPath)
			},
			servURL:      "",
			expectedPath: testTokenPath,
			expectErr:    false,
		},
		{
			name: "invalid header",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Www-Authenticate", "invalid")
			},
			servURL:      "",
			expectedPath: "",
			expectErr:    true,
		},
		{
			name: "client error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Www-Authenticate", "Basic realm="+testTokenPath)
			},
			servURL:      "http://invalid", // simulates client error
			expectedPath: "",
			expectErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(tc.handler)
			defer srv.Close()

			url := srv.URL
			if tc.servURL != "" {
				url = tc.servURL
			}

			challengeTokenPath, err := getChallengeTokenPath(
				context.Background(),
				srv.Client(),
				url,
				testAPIVersion,
				testServerID,
			)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expectedPath, challengeTokenPath)
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	testCases := []struct {
		name          string
		challengePath string
		tokenResponse string
		servURL       string
		expectedToken adal.Token
		expectErr     bool
	}{
		{
			name:          "valid response",
			tokenResponse: `{"access_token":"accessToken","refresh_token":"refreshToken","expires_on":84245, "not_before":"1733696828", "resource": "serverID", "token_type":"Bearer"}`,
			challengePath: path.Join(t.TempDir(), "fake.key"),
			servURL:       "",
			expectedToken: adal.Token{
				AccessToken:  "accessToken",
				RefreshToken: "refreshToken",
				ExpiresOn:    json.Number("84245"),
				NotBefore:    json.Number("1733696828"),
				Resource:     testServerID,
				Type:         "Bearer",
			},
			expectErr: false,
		},
		{
			name:          "invalid header",
			challengePath: "invalid",
			tokenResponse: "",
			servURL:       "",
			expectedToken: adal.Token{},
			expectErr:     true,
		},
		{
			name:          "empty challenge token path",
			challengePath: "",
			tokenResponse: "",
			servURL:       "",
			expectedToken: adal.Token{},
			expectErr:     true,
		},
		{
			name:          "client error",
			challengePath: path.Join(t.TempDir(), "fake.key"),
			tokenResponse: "",
			servURL:       "http://invalid", // simulates client error
			expectedToken: adal.Token{},
			expectErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.tokenResponse != "" {
					w.Write([]byte(tc.tokenResponse))
				}

				w.Header().Set("Www-Authenticate", "Basic realm="+tc.challengePath)
			}))
			defer srv.Close()

			url := srv.URL
			if tc.servURL != "" {
				url = tc.servURL
			}

			if tc.challengePath != "" && tc.challengePath != "invalid" {
				// Write the test token in the challenge path
				err := os.WriteFile(tc.challengePath, []byte("accessToken"), 0644)
				require.NoError(t, err)
			}

			token, err := getBearerToken(
				context.Background(),
				srv.Client(),
				url,
				tc.challengePath,
				testAPIVersion,
				testServerID,
			)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expectedToken, token)
		})
	}
}

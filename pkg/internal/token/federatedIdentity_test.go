package token

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Azure/kubelogin/pkg/internal/testutils"
)

func TestNewWorkloadIdentityTokenProviderEmpty(t *testing.T) {
	testData := []struct {
		name string
	}{
		{
			name: "clientID cannot be empty",
		},
		{
			name: "tenantID cannot be empty",
		},
		{
			name: "either ACTIONS_ID_TOKEN_REQUEST_TOKEN and ACTIONS_ID_TOKEN_REQUEST_URL environment variables have to be set or federated token file has to be provided",
		},
		{
			name: "authorityHost cannot be empty",
		},
		{
			name: "serverID cannot be empty",
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {

			name := data.name
			var err error

			switch {
			case strings.Contains(name, "clientID"):
				_, err = newWorkloadIdentityToken("", "", "", "", "")
			case strings.Contains(name, "federated token file"):
				_, err = newWorkloadIdentityToken("test", "", "", "", "test")
			case strings.Contains(name, "authorityHost"):
				_, err = newWorkloadIdentityToken("test", "test", "", "", "test")
			case strings.Contains(name, "serverID"):
				_, err = newWorkloadIdentityToken("test", "test", "test", "", "test")
			case strings.Contains(name, "tenantID"):
				_, err = newWorkloadIdentityToken("test", "test", "test", "test", "")
			}

			if !testutils.ErrorContains(err, data.name) {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestReadJWTFromFSEmptyString(t *testing.T) {
	_, err := readJWTFromFS("")
	if !testutils.ErrorContains(err, "no such file or directory") {
		t.Errorf("unexpected error: %v", err)
	}
}

func invalidHttpRequest(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(msg))
}

func TestUseGitHubToken(t *testing.T) {
	var (
		ghToken   = "foo-bar"
		oidcToken = "oidc-token"
	)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			invalidHttpRequest(w, fmt.Sprintf("unexpected method: %s", r.Method))
			return
		}
		if r.URL.Query().Get("audience") != azureADAudience {
			invalidHttpRequest(w, fmt.Sprintf("unexpected audience: %s", r.URL.Query().Get("audience")))
			return
		}
		if r.Header.Get("Authorization") != fmt.Sprintf("bearer %s", ghToken) {
			invalidHttpRequest(w, fmt.Sprintf("unexpected Authorization header: %s", r.Header.Get("Authorization")))
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			invalidHttpRequest(w, fmt.Sprintf("unexpected Content-Type header: %s", r.Header.Get("Content-Type")))
			return
		}
		if r.Header.Get("Accept") != "application/json; api-version=2.0" {
			invalidHttpRequest(w, fmt.Sprintf("unexpected Accept header: %s", r.Header.Get("Accept")))
			return
		}
		tokenResponse := githubTokenResponse{
			Value: oidcToken,
		}

		json.NewEncoder(w).Encode(tokenResponse)
	}))
	defer ts.Close()

	t.Setenv(actionsIDTokenRequestURL, ts.URL)
	t.Setenv(actionsIDTokenRequestToken, ghToken)

	token, err := getGitHubToken(context.Background())
	if err != nil {
		t.Fatalf("getGitHubToken returned unexpected error: %s", err)
	}
	if token != oidcToken {
		t.Fatalf("got token: %s, expected: %s", token, oidcToken)
	}

}

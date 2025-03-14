package token

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGithubActionsCredential(t *testing.T) {
	t.Run("valid options", func(t *testing.T) {
		opts := &Options{
			ClientID: "test-client-id",
			TenantID: "test-tenant-id",
		}

		cred, err := newGithubActionsCredential(opts)
		assert.NoError(t, err)
		assert.NotNil(t, cred)
		assert.Equal(t, "GithubActionsCredential", cred.Name())
	})

	t.Run("missing client ID", func(t *testing.T) {
		opts := &Options{
			TenantID: "test-tenant-id",
		}

		cred, err := newGithubActionsCredential(opts)
		assert.Error(t, err)
		assert.Nil(t, cred)
		assert.Equal(t, "client ID cannot be empty", err.Error())
	})

	t.Run("missing tenant ID", func(t *testing.T) {
		opts := &Options{
			ClientID: "test-client-id",
		}

		cred, err := newGithubActionsCredential(opts)
		assert.Error(t, err)
		assert.Nil(t, cred)
		assert.Equal(t, "tenant ID cannot be empty", err.Error())
	})
}

func TestGetGitHubToken(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"value":"TEST_ACCESS_TOKEN"}`))
		}))
		defer ts.Close()

		os.Setenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN", "test-token")
		os.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", ts.URL)

		token, err := getGitHubToken(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "TEST_ACCESS_TOKEN", token)
	})

	t.Run("invalid token", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"value":""}`))
		}))
		defer ts.Close()

		os.Setenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN", "test-token")
		os.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", ts.URL)

		token, err := getGitHubToken(context.Background())
		assert.Error(t, err)
		assert.Equal(t, "", token)
		assert.Equal(t, "github actions ID token is empty", err.Error())
	})

	t.Run("http request failure", func(t *testing.T) {
		os.Setenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN", "test-token")
		os.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "http://invalid-url")

		token, err := getGitHubToken(context.Background())
		assert.Error(t, err)
		assert.Equal(t, "", token)
	})

	t.Run("invalid response from GitHub", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"invalid":"response"}`))
		}))
		defer ts.Close()

		os.Setenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN", "test-token")
		os.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", ts.URL)

		token, err := getGitHubToken(context.Background())
		assert.Error(t, err)
		assert.Equal(t, "", token)
	})
}

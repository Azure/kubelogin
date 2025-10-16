package token

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
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

func Test_newGithubActionsCredential(t *testing.T) {
	type args struct {
		opts *Options
	}
	tests := []struct {
		name    string
		args    args
		want    CredentialProvider
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newGithubActionsCredential(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("newGithubActionsCredential() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newGithubActionsCredential() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGithubActionsCredential_Name(t *testing.T) {
	tests := []struct {
		name string
		c    *GithubActionsCredential
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Name(); got != tt.want {
				t.Errorf("GithubActionsCredential.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGithubActionsCredential_Authenticate(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts *policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *GithubActionsCredential
		args    args
		want    azidentity.AuthenticationRecord
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Authenticate(tt.args.ctx, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("GithubActionsCredential.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GithubActionsCredential.Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGithubActionsCredential_GetToken(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *GithubActionsCredential
		args    args
		want    azcore.AccessToken
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.GetToken(tt.args.ctx, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("GithubActionsCredential.GetToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GithubActionsCredential.GetToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGithubActionsCredential_NeedAuthenticate(t *testing.T) {
	tests := []struct {
		name string
		c    *GithubActionsCredential
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.NeedAuthenticate(); got != tt.want {
				t.Errorf("GithubActionsCredential.NeedAuthenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getGitHubToken(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getGitHubToken(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("getGitHubToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getGitHubToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

package token

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

const (
	actionsIDTokenRequestToken = "ACTIONS_ID_TOKEN_REQUEST_TOKEN"
	actionsIDTokenRequestURL   = "ACTIONS_ID_TOKEN_REQUEST_URL"
	azureADAudience            = "api://AzureADTokenExchange"
	defaultScope               = "/.default"
)

type workloadIdentityToken struct {
	serverID string
	client   confidential.Client
}

type githubTokenResponse struct {
	Value string `json:"value"`
}

func newWorkloadIdentityToken(clientID, federatedTokenFile, authorityHost, serverID, tenantID string) (TokenProvider, error) {
	if clientID == "" {
		return nil, errors.New("clientID cannot be empty")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	hasActionsIDToken := os.Getenv(actionsIDTokenRequestToken) != "" && os.Getenv(actionsIDTokenRequestURL) != ""
	if federatedTokenFile == "" && !hasActionsIDToken {
		return nil, errors.New("either ACTIONS_ID_TOKEN_REQUEST_TOKEN and ACTIONS_ID_TOKEN_REQUEST_URL environment variables have to be set or federated token file has to be provided")
	}
	if authorityHost == "" {
		return nil, errors.New("authorityHost cannot be empty")
	}
	if serverID == "" {
		return nil, errors.New("serverID cannot be empty")
	}

	var cred confidential.Credential
	if federatedTokenFile != "" {
		cred = newCredentialFromTokenFile(federatedTokenFile)
	} else {
		cred = newCredentialFromGithub()
	}

	client, err := confidential.New(fmt.Sprintf("%s%s/oauth2/token", authorityHost, tenantID), clientID, cred)
	if err != nil {
		return nil, fmt.Errorf("failed to create confidential client for federated workload identity. %s", err)
	}

	return &workloadIdentityToken{
		serverID: serverID,
		client:   client,
	}, nil
}

func (p *workloadIdentityToken) Token() (adal.Token, error) {
	emptyToken := adal.Token{}

	resource := strings.TrimSuffix(p.serverID, "/")
	// .default needs to be added to the scope
	if !strings.HasSuffix(resource, ".default") {
		resource += defaultScope
	}

	result, err := p.client.AcquireTokenByCredential(context.Background(), []string{resource})
	if err != nil {
		return emptyToken, fmt.Errorf("failed to acquire token. %s", err)
	}

	return adal.Token{
		AccessToken: result.AccessToken,
		ExpiresOn:   json.Number(fmt.Sprintf("%d", result.ExpiresOn.UTC().Unix())),
		Resource:    p.serverID,
	}, nil
}

// newCredentialFromTokenFile creates a confidential.Credential from provided token file
func newCredentialFromTokenFile(federatedTokenFile string) confidential.Credential {
	cb := func(_ context.Context, _ confidential.AssertionRequestOptions) (string, error) {
		return readJWTFromFS(federatedTokenFile)
	}
	return confidential.NewCredFromAssertionCallback(cb)
}

// newCredentialFromGithub creates a confidential.Credential from github id token
func newCredentialFromGithub() confidential.Credential {
	cb := func(ctx context.Context, _ confidential.AssertionRequestOptions) (string, error) {
		return getGitHubToken(ctx)
	}
	return confidential.NewCredFromAssertionCallback(cb)
}

// readJWTFromFS reads the jwt from file system
func readJWTFromFS(tokenFilePath string) (string, error) {
	token, err := os.ReadFile(tokenFilePath)
	if err != nil {
		return "", err
	}
	return string(token), nil
}

func getGitHubToken(ctx context.Context) (string, error) {
	reqToken := os.Getenv(actionsIDTokenRequestToken)
	reqURL := os.Getenv(actionsIDTokenRequestURL)

	if reqToken == "" || reqURL == "" {
		return "", errors.New("ACTIONS_ID_TOKEN_REQUEST_TOKEN or ACTIONS_ID_TOKEN_REQUEST_URL is not set")
	}

	u, err := url.Parse(reqURL)
	if err != nil {
		return "", fmt.Errorf("unable to parse ACTIONS_ID_TOKEN_REQUEST_URL: %w", err)
	}
	q := u.Query()
	q.Set("audience", azureADAudience)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return "", err
	}

	// reference:
	// https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", reqToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json; api-version=2.0")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var body string
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			body = err.Error()
		} else {
			body = string(b)
		}

		return "", fmt.Errorf("github actions ID token request failed with status code: %d, response body: %s", resp.StatusCode, body)
	}

	var tokenResp githubTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	if tokenResp.Value == "" {
		return "", errors.New("github actions ID token is empty")
	}

	return tokenResp.Value, nil
}

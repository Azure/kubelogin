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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

type githubTokenResponse struct {
	Value string `json:"value"`
}

type GithubActionsCredential struct {
	client confidential.Client
}

var _ CredentialProvider = (*GithubActionsCredential)(nil)

func newGithubActionsCredential(opts *Options) (CredentialProvider, error) {
	if opts.ClientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}
	if opts.TenantID == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	cred := confidential.NewCredFromAssertionCallback(func(ctx context.Context, _ confidential.AssertionRequestOptions) (string, error) {
		return getGitHubToken(ctx)
	})

	o := []confidential.Option{
		confidential.WithInstanceDiscovery(!opts.DisableInstanceDiscovery),
	}
	if opts.httpClient != nil {
		o = append(o, confidential.WithHTTPClient(opts.httpClient))
	}
	client, err := confidential.New(
		fmt.Sprintf("%s%s/", opts.GetCloudConfiguration().ActiveDirectoryAuthorityHost, opts.TenantID),
		opts.ClientID, cred, o...)
	if err != nil {
		return nil, fmt.Errorf("failed to create github actions credential: %w", err)
	}

	return &GithubActionsCredential{client: client}, nil
}

func (c *GithubActionsCredential) Name() string {
	return "GithubActionsCredential"
}

func (c *GithubActionsCredential) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	return azidentity.AuthenticationRecord{}, errAuthenticateNotSupported
}

func (c *GithubActionsCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	result, err := c.client.AcquireTokenByCredential(ctx, opts.Scopes)
	if err != nil {
		return azcore.AccessToken{}, err
	}

	return azcore.AccessToken{Token: result.AccessToken, ExpiresOn: result.ExpiresOn}, nil
}

func (c *GithubActionsCredential) NeedAuthenticate() bool {
	return false
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

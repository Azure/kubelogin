package token

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Azure/go-autorest/autorest/adal"
)

const (
	defaultIdentityEndpoint = "http://127.0.0.1:40342/metadata/identity/oauth2/token"
)

// HIMDSToken is a struct that implements the TokenProvider interface to return a token using the HIMDS service.
type HIMDSToken struct {
	httpClient *http.Client
	apiVersion string
	serverID   string

	identityEndpoint string
}

// newHIMDSToken creates a new HIMDSToken instance which implements the TokenProvider interface.
func newHIMDSToken(serverID, apiVersion, identityEndpoint string) (HIMDSToken, error) {
	himdsEndpoint := identityEndpoint
	if himdsEndpoint == "" {
		himdsEndpoint = defaultIdentityEndpoint
	}

	return HIMDSToken{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		apiVersion:       apiVersion,
		serverID:         serverID,
		identityEndpoint: himdsEndpoint,
	}, nil
}

// Token implements the TokenProvider interface to return a token using the HIMDS service.
func (h HIMDSToken) Token(ctx context.Context) (adal.Token, error) {
	challengeTokenPath, err := getChallengeTokenPath(
		ctx,
		h.httpClient,
		h.identityEndpoint,
		h.apiVersion,
		h.serverID,
	)
	if err != nil {
		return adal.Token{}, fmt.Errorf("failed to get challenge token path: %w", err)
	}

	if challengeTokenPath == "" {
		return adal.Token{}, fmt.Errorf("challenge token path is empty")
	}

	return getBearerToken(ctx, h.httpClient, h.identityEndpoint, challengeTokenPath, h.apiVersion, h.serverID)
}

// getChallengeTokenPath  returns the challenge token path from the HIMDS service.
func getChallengeTokenPath(
	ctx context.Context,
	httpClient *http.Client,
	identityEndpoint, apiVersion, resource string,
) (string, error) {
	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, identityEndpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create himds request: %w", err)
	}

	// Add the required query parameters
	req.Header.Set("Metadata", "true")
	q := req.URL.Query()
	q.Add("api-version", apiVersion)
	q.Add("resource", resource)
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send himds token request: %w", err)
	}

	return extractTokenPath(resp.Header.Get("Www-Authenticate"))
}

// extractTokenPath extracts the token path from the WWW-Authenticate header.
func extractTokenPath(authHeader string) (string, error) {
	const prefix = "Basic realm="
	if len(authHeader) < len(prefix) {
		return "", fmt.Errorf("invalid auth header")
	}

	return strings.TrimPrefix(authHeader, prefix), nil
}

func getBearerToken(
	ctx context.Context,
	httpClient *http.Client,
	identityEndpoint, challengeTokenPath, apiVersion, resource string,
) (adal.Token, error) {
	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, identityEndpoint, nil)
	if err != nil {
		return adal.Token{}, fmt.Errorf("failed to create himds request: %w", err)
	}

	if challengeTokenPath == "" {
		return adal.Token{}, fmt.Errorf("challenge token path is empty")
	}

	challengeToken, err := os.ReadFile(challengeTokenPath)
	if err != nil {
		return adal.Token{}, fmt.Errorf("failed to read challenge token: %w", err)
	}

	// Add challenge token path to the request
	req.Header.Set("Metadata", "true")
	req.Header.Set("Authorization", "Basic "+string(challengeToken))

	// Add the required query parameters
	q := req.URL.Query()
	q.Add("api-version", apiVersion)
	q.Add("resource", resource)
	req.URL.RawQuery = q.Encode()

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return adal.Token{}, fmt.Errorf("failed to send himds token request: %w", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var t adal.Token
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return adal.Token{}, fmt.Errorf("failed to decode himds token response: %w", err)
	}

	return t, nil

}

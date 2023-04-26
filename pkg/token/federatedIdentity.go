package token

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

type workloadIdentityToken struct {
	clientID           string
	tenantID           string
	federatedTokenFile string
	authorityHost      string
	serverID           string
}

func newWorkloadIdentityToken(clientID, federatedTokenFile, authorityHost, serverID, tenantID string) (TokenProvider, error) {
	if clientID == "" {
		return nil, errors.New("clientID cannot be empty")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if federatedTokenFile == "" {
		return nil, errors.New("federatedTokenFile cannot be empty")
	}
	if authorityHost == "" {
		return nil, errors.New("authorityHost cannot be empty")
	}
	if serverID == "" {
		return nil, errors.New("serverID cannot be empty")
	}

	return &workloadIdentityToken{
		clientID:           clientID,
		tenantID:           tenantID,
		federatedTokenFile: federatedTokenFile,
		authorityHost:      authorityHost,
		serverID:           serverID,
	}, nil
}

func (p *workloadIdentityToken) Token() (adal.Token, error) {
	emptyToken := adal.Token{}
	cred, err := newCredential(p.federatedTokenFile)
	if err != nil {
		return emptyToken, err
	}

	// create the confidential client to request an AAD token
	confidentialClientApp, err := createClient(p.authorityHost, p.tenantID, p.clientID, cred)
	if err != nil {
		return emptyToken, err
	}

	resource := strings.TrimSuffix(p.serverID, "/")
	// .default needs to be added to the scope
	if !strings.HasSuffix(resource, ".default") {
		resource += "/.default"
	}

	result, err := confidentialClientApp.AcquireTokenByCredential(context.Background(), []string{resource})
	if err != nil {
		return emptyToken, fmt.Errorf("failed to acquire token. %s", err)
	}

	return adal.Token{
		AccessToken: result.AccessToken,
		ExpiresOn:   json.Number(fmt.Sprintf("%d", result.ExpiresOn.UTC().Unix())),
		Resource:    p.serverID,
	}, nil
}

func newCredential(federatedTokenFile string) (confidential.Credential, error) {
	signedAssertion, err := readJWTFromFS(federatedTokenFile)
	if err != nil {
		return confidential.Credential{}, fmt.Errorf("failed to read signed assertion from token file: %s", err)
	}
	signedAssertionCallback := func(_ context.Context, _ confidential.AssertionRequestOptions) (string, error) {
		return signedAssertion, nil
	}
	return confidential.NewCredFromAssertionCallback(signedAssertionCallback), nil
}

func createClient(authorityHost string, tenantID string, clientID string, cred confidential.Credential) (confidential.Client, error) {
	authority := fmt.Sprintf("%s%s/oauth2/token", authorityHost, tenantID)
	confidentialClientApp, err := confidential.New(
		authority,
		clientID,
		cred)

	if err != nil {
		return confidential.Client{}, fmt.Errorf("failed to create confidential client app. %s", err)
	}

	return confidentialClientApp, err
}

// readJWTFromFS reads the jwt from file system
func readJWTFromFS(tokenFilePath string) (string, error) {
	token, err := os.ReadFile(tokenFilePath)
	if err != nil {
		return "", err
	}
	return string(token), nil
}

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
	resourceID         string
}

func newWorkloadIdentityToken(clientID, federatedTokenFile, authorityHost, resourceID, tenantID string) (TokenProvider, error) {
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
	if resourceID == "" {
		return nil, errors.New("resourceID cannot be empty")
	}

	return &workloadIdentityToken{
		clientID:           clientID,
		tenantID:           tenantID,
		federatedTokenFile: federatedTokenFile,
		authorityHost:      authorityHost,
		resourceID:         resourceID,
	}, nil
}

func (p *workloadIdentityToken) Token() (adal.Token, error) {
	emptyToken := adal.Token{}

	signedAssertion, err := readJWTFromFS(p.federatedTokenFile)
	if err != nil {
		return emptyToken, fmt.Errorf("failed to read service account token: %s", err)
	}
	cred, err := confidential.NewCredFromAssertion(signedAssertion)
	if err != nil {
		return emptyToken, fmt.Errorf("failed to create confidential creds: %s", err)
	}

	// create the confidential client to request an AAD token
	confidentialClientApp, err := confidential.New(
		p.clientID,
		cred,
		confidential.WithAuthority(fmt.Sprintf("%s%s/oauth2/token", p.authorityHost, p.tenantID)))
	if err != nil {
		return emptyToken, fmt.Errorf("failed to create confidential client app. %s", err)
	}

	resource := strings.TrimSuffix(p.resourceID, "/")
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
		ExpiresOn:   json.Number(fmt.Sprintf("%v", result.ExpiresOn.UTC().Unix())),
	}, nil
}

// readJWTFromFS reads the jwt from file system
func readJWTFromFS(tokenFilePath string) (string, error) {
	token, err := os.ReadFile(tokenFilePath)
	if err != nil {
		return "", err
	}
	return string(token), nil
}

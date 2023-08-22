package token

import (
	"errors"
	"fmt"
	"os"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
)

type deviceCodeTokenProvider struct {
	clientID    string
	resourceID  string
	tenantID    string
	oAuthConfig adal.OAuthConfig
}

func newDeviceCodeTokenProvider(oAuthConfig adal.OAuthConfig, clientID, resourceID, tenantID string) (TokenProvider, error) {
	if clientID == "" {
		return nil, errors.New("clientID cannot be empty")
	}
	if resourceID == "" {
		return nil, errors.New("resourceID cannot be empty")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}

	return &deviceCodeTokenProvider{
		clientID:    clientID,
		resourceID:  resourceID,
		tenantID:    tenantID,
		oAuthConfig: oAuthConfig,
	}, nil
}

func (p *deviceCodeTokenProvider) Token() (adal.Token, error) {
	emptyToken := adal.Token{}
	client := &autorest.Client{}
	deviceCode, err := adal.InitiateDeviceAuth(client, p.oAuthConfig, p.clientID, p.resourceID)
	if err != nil {
		return emptyToken, fmt.Errorf("initialing the device code authentication: %s", err)
	}

	_, err = fmt.Fprintln(os.Stderr, *deviceCode.Message)
	if err != nil {
		return emptyToken, fmt.Errorf("prompting the device code message: %s", err)
	}

	token, err := adal.WaitForUserCompletion(client, deviceCode)
	if err != nil {
		return emptyToken, fmt.Errorf("waiting for device code authentication to complete: %s", err)
	}

	return *token, nil
}

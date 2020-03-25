package token

import (
	"errors"
	"fmt"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

type TokenProvider interface {
	Token() (adal.Token, error)
}

func newTokenProvider(o *Options, token *adal.Token) (TokenProvider, error) {
	oAuthConfig, err := getOAuthConfig(o.Environment, o.TenantID, o.IsLegacy)
	if err != nil {
		return nil, fmt.Errorf("failed to get oAuthConfig. isLegacy: %t, err: %s", o.IsLegacy, err)
	}
	switch o.LoginMethod {
	case deviceCodeLogin:
		return newDeviceCodeTokenProvider(*oAuthConfig, o.ClientID, o.ServerID, o.TenantID)
	case servicePrincipalLogin:
		return newServicePrincipalToken(*oAuthConfig, o.ClientID, o.ClientSecret, o.ServerID, o.TenantID)
	case ropcLogin:
		return newResourceOwnerToken(*oAuthConfig, o.ClientID, o.Username, o.Password, o.ServerID, o.TenantID)
	}

	return nil, errors.New("unsupported token provider")
}

func getOAuthConfig(envName, tenantID string, isLegacy bool) (*adal.OAuthConfig, error) {
	var (
		oAuthConfig *adal.OAuthConfig
		environment azure.Environment
		err         error
	)
	environment, err = getAzureEnvironment(envName)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment: %s", err)
	}
	if isLegacy {
		oAuthConfig, err = adal.NewOAuthConfig(environment.ActiveDirectoryEndpoint, tenantID)
	} else {
		oAuthConfig, err = adal.NewOAuthConfigWithAPIVersion(environment.ActiveDirectoryEndpoint, tenantID, nil)
	}
	return oAuthConfig, err
}

func getAzureEnvironment(environment string) (azure.Environment, error) {
	if environment == "" {
		environment = defaultEnvironmentName
	}
	return azure.EnvironmentFromName(environment)
}

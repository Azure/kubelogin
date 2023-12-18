package token

//go:generate sh -c "mockgen -destination mock_$GOPACKAGE/provider.go github.com/Azure/kubelogin/pkg/internal/token TokenProvider"

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

type TokenProvider interface {
	Token(ctx context.Context) (adal.Token, error)
}

// NewTokenProvider creates the TokenProvider instance with giving options.
func NewTokenProvider(o *Options) (TokenProvider, error) {
	oAuthConfig, err := getOAuthConfig(o.Environment, o.TenantID, o.IsLegacy)
	if err != nil {
		return nil, fmt.Errorf("failed to get oAuthConfig. isLegacy: %t, err: %s", o.IsLegacy, err)
	}
	cloudConfiguration, err := getCloudConfig(o.Environment)
	if err != nil {
		return nil, fmt.Errorf("failed to get cloud.Configuration. err: %s", err)
	}
	popClaimsMap, err := parsePoPClaims(o.PoPTokenClaims)
	if o.IsPoPTokenEnabled && err != nil {
		return nil, err
	}
	switch o.LoginMethod {
	case DeviceCodeLogin:
		return newDeviceCodeTokenProvider(*oAuthConfig, o.ClientID, o.ServerID, o.TenantID)
	case InteractiveLogin:
		return newInteractiveTokenProvider(*oAuthConfig, o.ClientID, o.ServerID, o.TenantID, popClaimsMap)
	case ServicePrincipalLogin:
		if o.IsLegacy {
			return newLegacyServicePrincipalToken(*oAuthConfig, o.ClientID, o.ClientSecret, o.ClientCert, o.ClientCertPassword, o.ServerID, o.TenantID)
		}
		return newServicePrincipalTokenProvider(cloudConfiguration, o.ClientID, o.ClientSecret, o.ClientCert, o.ClientCertPassword, o.ServerID, o.TenantID, popClaimsMap)
	case ROPCLogin:
		return newResourceOwnerToken(*oAuthConfig, o.ClientID, o.Username, o.Password, o.ServerID, o.TenantID)
	case MSILogin:
		return newManagedIdentityToken(o.ClientID, o.IdentityResourceID, o.ServerID)
	case AzureCLILogin:
		return newAzureCLIToken(o.ServerID, o.TenantID, o.Timeout)
	case WorkloadIdentityLogin:
		return newWorkloadIdentityToken(o.ClientID, o.FederatedTokenFile, o.AuthorityHost, o.ServerID, o.TenantID)
	}

	return nil, errors.New("unsupported token provider")
}

func getCloudConfig(envName string) (cloud.Configuration, error) {
	env, err := getAzureEnvironment(envName)
	c := cloud.Configuration{
		ActiveDirectoryAuthorityHost: env.ActiveDirectoryEndpoint,
	}
	return c, err
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

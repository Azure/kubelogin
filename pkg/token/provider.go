package token

//go:generate sh -c "mockgen -destination mock_$GOPACKAGE/provider.go github.com/Azure/kubelogin/pkg/token TokenProvider"

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

type TokenProvider interface {
	Token() (adal.Token, error)
}

func newTokenProvider(o *Options) (TokenProvider, error) {
	oAuthConfig, err := getOAuthConfig(o.Environment, o.TenantID, o.IsLegacy)
	if err != nil {
		return nil, fmt.Errorf("failed to get oAuthConfig. isLegacy: %t, err: %s", o.IsLegacy, err)
	}
	cloudConfiguration, err := getCloudConfig(o.Environment)
	if err != nil {
		return nil, fmt.Errorf("failed to get cloud.Configuration. err: %s", err)
	}
	popClaimsMap, err := parsePopClaims(o.PoPTokenClaims)
	if o.IsPoPTokenEnabled && err != nil {
		return nil, fmt.Errorf("failed to parse pop token claims. err: %w", err)
	}
	switch o.LoginMethod {
	case DeviceCodeLogin:
		return newDeviceCodeTokenProvider(*oAuthConfig, o.ClientID, o.ServerID, o.TenantID)
	case InteractiveLogin:
		return newInteractiveTokenProvider(*oAuthConfig, o.ClientID, o.ServerID, o.TenantID, popClaimsMap)
	case ServicePrincipalLogin:
		return newServicePrincipalToken(cloudConfiguration, o.ClientID, o.ClientSecret, o.ClientCert, o.ClientCertPassword, o.ServerID, o.TenantID, popClaimsMap)
	case ROPCLogin:
		return newResourceOwnerToken(*oAuthConfig, o.ClientID, o.Username, o.Password, o.ServerID, o.TenantID)
	case MSILogin:
		return newManagedIdentityToken(o.ClientID, o.IdentityResourceID, o.ServerID)
	case AzureCLILogin:
		return newAzureCLIToken(o.ServerID, o.TenantID)
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

// Parses the pop token claims. Pop token claims are passed in as a comma-separated string
// in the format "key1=val1,key2=val2".
func parsePopClaims(popClaims string) (map[string]string, error) {
	claimsArray := strings.Split(popClaims, ",")
	claimsMap := make(map[string]string)
	for _, claim := range claimsArray {
		claimPair := strings.Split(claim, "=")
		key := strings.TrimSpace(claimPair[0])
		val := strings.TrimSpace(claimPair[1])
		if key == "" || val == "" {
			return nil, fmt.Errorf("error parsing PoP token claims. Ensure the claims are formatted as `key=value` with no extra whitespace")
		}
		claimsMap[key] = val
	}
	if claimsMap["u"] == "" {
		return nil, fmt.Errorf("required u-claim not provided for PoP token flow. Please provide the ARM ID of the connected cluster in the format `u=<ARM_ID>`")
	}
	return claimsMap, nil
}

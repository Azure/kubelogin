package token

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"k8s.io/client-go/util/homedir"
)

type Options struct {
	LoginMethod            string
	ClientID               string
	ClientSecret           string
	ClientCert             string
	Username               string
	Password               string
	ServerID               string
	TenantID               string
	Environment            string
	IsLegacy               bool
	TokenCacheDir          string
	tokenCacheFile         string
	IdentityResourceId     string
	FederatedTokenFile     string
	AuthorityHost          string
	UseAzureRMTerraformEnv bool
}

const (
	defaultEnvironmentName = "AzurePublicCloud"

	DeviceCodeLogin       = "devicecode"
	ServicePrincipalLogin = "spn"
	ROPCLogin             = "ropc"
	MSILogin              = "msi"
	AzureCLILogin         = "azurecli"
	WorkloadIdentityLogin = "workloadidentity"
	manualTokenLogin      = "manual_token"

	envWorkloadIdentityClientID           = "AZURE_CLIENT_ID"
	envWorkloadIdentityFederatedTokenFile = "AZURE_FEDERATED_TOKEN_FILE"
	envWorkloadIdentityAuthorityHost      = "AZURE_AUTHORITY_HOST"
	envROPCUsername                       = "AAD_USER_PRINCIPAL_NAME"
	envROPCPassword                       = "AAD_USER_PRINCIPAL_PASSWORD"
	envLoginMethod                        = "AAD_LOGIN_METHOD"
)

var (
	supportedLogin                  []string
	DefaultTokenCacheDir            = homedir.HomeDir() + "/.kube/cache/kubelogin/"
	envServicePrincipalClientID     = "AAD_SERVICE_PRINCIPAL_CLIENT_ID"
	envServicePrincipalClientSecret = "AAD_SERVICE_PRINCIPAL_CLIENT_SECRET"
	envServicePrincipalClientCert   = "AAD_SERVICE_PRINCIPAL_CLIENT_CERTIFICATE"
	envTenantID                     = "AZURE_TENANT_ID"
)

func init() {
	supportedLogin = []string{DeviceCodeLogin, ServicePrincipalLogin, ROPCLogin, MSILogin, AzureCLILogin, WorkloadIdentityLogin}
}

func GetSupportedLogins() string {
	return strings.Join(supportedLogin, ", ")
}

func NewOptions() Options {
	return Options{
		LoginMethod:   DeviceCodeLogin,
		Environment:   defaultEnvironmentName,
		TokenCacheDir: DefaultTokenCacheDir,
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.LoginMethod, "login", "l", o.LoginMethod, fmt.Sprintf("Login method. Supported methods: %s. It may be specified in %s environment variable", GetSupportedLogins(), envLoginMethod))
	fs.StringVar(&o.ClientID, "client-id", o.ClientID, fmt.Sprintf("AAD client application ID. It may be specified in %s environment variable", envServicePrincipalClientID))
	fs.StringVar(&o.ClientSecret, "client-secret", o.ClientSecret, fmt.Sprintf("AAD client application secret. Used in spn login. It may be specified in %s environment variable", envServicePrincipalClientSecret))
	fs.StringVar(&o.ClientCert, "client-certificate", o.ClientCert, fmt.Sprintf("AAD client cert in pfx. Used in spn login. It may be specified in %s environment variable", envServicePrincipalClientCert))
	fs.StringVar(&o.Username, "username", o.Username, fmt.Sprintf("user name for ropc login flow. It may be specified in %s environment variable", envROPCUsername))
	fs.StringVar(&o.Password, "password", o.Password, fmt.Sprintf("password for ropc login flow. It may be specified in %s environment variable", envROPCPassword))
	fs.StringVar(&o.IdentityResourceId, "identity-resource-id", o.IdentityResourceId, "Managed Identity resource id.")
	fs.StringVar(&o.ServerID, "server-id", o.ServerID, "AAD server application ID")
	fs.StringVar(&o.FederatedTokenFile, "federated-token-file", o.FederatedTokenFile, "Workload Identity federated token file")
	fs.StringVar(&o.AuthorityHost, "authority-host", o.AuthorityHost, "Workload Identity authority host")
	fs.StringVar(&o.TokenCacheDir, "token-cache-dir", o.TokenCacheDir, "directory to cache token")
	fs.StringVarP(&o.TenantID, "tenant-id", "t", o.TenantID, fmt.Sprintf("AAD tenant ID. It may be specified in %s environment variable", envTenantID))
	fs.StringVarP(&o.Environment, "environment", "e", o.Environment, "Azure environment name")
	fs.BoolVar(&o.IsLegacy, "legacy", o.IsLegacy, "set to true to get token with 'spn:' prefix in audience claim")
	fs.BoolVar(&o.UseAzureRMTerraformEnv, "use-azurerm-env-vars", o.UseAzureRMTerraformEnv, "Use environment variable names of Terraform Azure Provider (ARM_CLIENT_ID, ARM_CLIENT_SECRET, ARM_CLIENT_CERTIFICATE_PATH, ARM_TENANT_ID)")
}

func (o *Options) Validate() error {
	foundValidLoginMethod := false
	for _, v := range supportedLogin {
		if o.LoginMethod == v {
			foundValidLoginMethod = true
		}
	}

	if !foundValidLoginMethod {
		return fmt.Errorf("'%s' is not a supported login method. Supported method is one of %s", o.LoginMethod, GetSupportedLogins())
	}
	return nil
}

func (o *Options) UpdateFromEnv() {
	o.tokenCacheFile = getCacheFileName(o)

	if o.UseAzureRMTerraformEnv {
		envServicePrincipalClientID = "ARM_CLIENT_ID"
		envServicePrincipalClientSecret = "ARM_CLIENT_SECRET"
		envServicePrincipalClientCert = "ARM_CLIENT_CERTIFICATE_PATH"
		envTenantID = "ARM_TENANT_ID"
	}

	if v, ok := os.LookupEnv(envServicePrincipalClientID); ok {
		o.ClientID = v
	}
	if v, ok := os.LookupEnv(envServicePrincipalClientSecret); ok {
		o.ClientSecret = v
	}
	if v, ok := os.LookupEnv(envServicePrincipalClientCert); ok {
		o.ClientCert = v
	}
	if v, ok := os.LookupEnv(envROPCUsername); ok {
		o.Username = v
	}
	if v, ok := os.LookupEnv(envROPCPassword); ok {
		o.Password = v
	}
	if v, ok := os.LookupEnv(envLoginMethod); ok {
		o.LoginMethod = v
	}

	if v, ok := os.LookupEnv(envTenantID); ok {
		o.TenantID = v
	}

	if o.LoginMethod == WorkloadIdentityLogin {
		if v, ok := os.LookupEnv(envWorkloadIdentityClientID); ok {
			o.ClientID = v
		}
		if v, ok := os.LookupEnv(envWorkloadIdentityFederatedTokenFile); ok {
			o.FederatedTokenFile = v
		}
		if v, ok := os.LookupEnv(envWorkloadIdentityAuthorityHost); ok {
			o.AuthorityHost = v
		}
	}
}

func (o *Options) String() string {
	return fmt.Sprintf("Login Method: %s, Environment: %s, TenantID: %s, ServerID: %s, ClientID: %s, IsLegacy: %t, msiResourceID: %s, tokenCacheDir: %s, tokenCacheFile: %s",
		o.LoginMethod,
		o.Environment,
		o.TenantID,
		o.ServerID,
		o.ClientID,
		o.IsLegacy,
		o.IdentityResourceId,
		o.TokenCacheDir,
		o.tokenCacheFile)
}

func getCacheFileName(o *Options) string {
	// format: ${environment}-${server-id}-${client-id}-${tenant-id}[_legacy].json
	cacheFileNameFormat := "%s-%s-%s-%s.json"
	if o.IsLegacy {
		cacheFileNameFormat = "%s-%s-%s-%s_legacy.json"
	}
	return filepath.Join(o.TokenCacheDir, fmt.Sprintf(cacheFileNameFormat, o.Environment, o.ServerID, o.ClientID, o.TenantID))
}

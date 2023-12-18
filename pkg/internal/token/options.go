package token

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/client-go/util/homedir"
)

type Options struct {
	LoginMethod            string
	ClientID               string
	ClientSecret           string
	ClientCert             string
	ClientCertPassword     string
	Username               string
	Password               string
	ServerID               string
	TenantID               string
	Environment            string
	IsLegacy               bool
	Timeout                time.Duration
	TokenCacheDir          string
	tokenCacheFile         string
	IdentityResourceID     string
	FederatedTokenFile     string
	AuthorityHost          string
	UseAzureRMTerraformEnv bool
	IsPoPTokenEnabled      bool
	PoPTokenClaims         string
}

const (
	defaultEnvironmentName = "AzurePublicCloud"

	DeviceCodeLogin       = "devicecode"
	InteractiveLogin      = "interactive"
	ServicePrincipalLogin = "spn"
	ROPCLogin             = "ropc"
	MSILogin              = "msi"
	AzureCLILogin         = "azurecli"
	WorkloadIdentityLogin = "workloadidentity"
	manualTokenLogin      = "manual_token"

	// env vars
	loginMethod                        = "AAD_LOGIN_METHOD"
	kubeloginROPCUsername              = "AAD_USER_PRINCIPAL_NAME"
	kubeloginROPCPassword              = "AAD_USER_PRINCIPAL_PASSWORD"
	kubeloginClientID                  = "AAD_SERVICE_PRINCIPAL_CLIENT_ID"
	kubeloginClientSecret              = "AAD_SERVICE_PRINCIPAL_CLIENT_SECRET"
	kubeloginClientCertificatePath     = "AAD_SERVICE_PRINCIPAL_CLIENT_CERTIFICATE"
	kubeloginClientCertificatePassword = "AAD_SERVICE_PRINCIPAL_CLIENT_CERTIFICATE_PASSWORD"

	// env vars used by Terraform
	terraformClientID                  = "ARM_CLIENT_ID"
	terraformClientSecret              = "ARM_CLIENT_SECRET"
	terraformClientCertificatePath     = "ARM_CLIENT_CERTIFICATE_PATH"
	terraformClientCertificatePassword = "ARM_CLIENT_CERTIFICATE_PASSWORD"
	terraformTenantID                  = "ARM_TENANT_ID"

	// env vars following azure sdk naming convention
	azureAuthorityHost             = "AZURE_AUTHORITY_HOST"
	azureClientCertificatePassword = "AZURE_CLIENT_CERTIFICATE_PASSWORD"
	azureClientCertificatePath     = "AZURE_CLIENT_CERTIFICATE_PATH"
	azureClientID                  = "AZURE_CLIENT_ID"
	azureClientSecret              = "AZURE_CLIENT_SECRET"
	azureFederatedTokenFile        = "AZURE_FEDERATED_TOKEN_FILE"
	azureTenantID                  = "AZURE_TENANT_ID"
	azureUsername                  = "AZURE_USERNAME"
	azurePassword                  = "AZURE_PASSWORD"
)

var (
	supportedLogin       []string
	DefaultTokenCacheDir = homedir.HomeDir() + "/.kube/cache/kubelogin/"
)

func init() {
	supportedLogin = []string{DeviceCodeLogin, InteractiveLogin, ServicePrincipalLogin, ROPCLogin, MSILogin, AzureCLILogin, WorkloadIdentityLogin}
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
	fs.StringVarP(&o.LoginMethod, "login", "l", o.LoginMethod,
		fmt.Sprintf("Login method. Supported methods: %s. It may be specified in %s environment variable", GetSupportedLogins(), loginMethod))
	fs.StringVar(&o.ClientID, "client-id", o.ClientID,
		fmt.Sprintf("AAD client application ID. It may be specified in %s or %s environment variable", kubeloginClientID, azureClientID))
	fs.StringVar(&o.ClientSecret, "client-secret", o.ClientSecret,
		fmt.Sprintf("AAD client application secret. Used in spn login. It may be specified in %s or %s environment variable", kubeloginClientSecret, azureClientSecret))
	fs.StringVar(&o.ClientCert, "client-certificate", o.ClientCert,
		fmt.Sprintf("AAD client cert in pfx. Used in spn login. It may be specified in %s or %s environment variable", kubeloginClientCertificatePath, azureClientCertificatePath))
	fs.StringVar(&o.ClientCertPassword, "client-certificate-password", o.ClientCertPassword,
		fmt.Sprintf("Password for AAD client cert. Used in spn login. It may be specified in %s or %s environment variable", kubeloginClientCertificatePassword, azureClientCertificatePassword))
	fs.StringVar(&o.Username, "username", o.Username,
		fmt.Sprintf("user name for ropc login flow. It may be specified in %s or %s environment variable", kubeloginROPCUsername, azureUsername))
	fs.StringVar(&o.Password, "password", o.Password,
		fmt.Sprintf("password for ropc login flow. It may be specified in %s or %s environment variable", kubeloginROPCPassword, azurePassword))
	fs.StringVar(&o.IdentityResourceID, "identity-resource-id", o.IdentityResourceID, "Managed Identity resource id.")
	fs.StringVar(&o.ServerID, "server-id", o.ServerID, "AAD server application ID")
	fs.StringVar(&o.FederatedTokenFile, "federated-token-file", o.FederatedTokenFile,
		fmt.Sprintf("Workload Identity federated token file. It may be specified in %s environment variable", azureFederatedTokenFile))
	fs.StringVar(&o.AuthorityHost, "authority-host", o.AuthorityHost,
		fmt.Sprintf("Workload Identity authority host. It may be specified in %s environment variable", azureAuthorityHost))
	fs.StringVar(&o.TokenCacheDir, "token-cache-dir", o.TokenCacheDir, "directory to cache token")
	fs.StringVarP(&o.TenantID, "tenant-id", "t", o.TenantID, fmt.Sprintf("AAD tenant ID. It may be specified in %s environment variable", azureTenantID))
	fs.StringVarP(&o.Environment, "environment", "e", o.Environment, "Azure environment name")
	fs.BoolVar(&o.IsLegacy, "legacy", o.IsLegacy, "set to true to get token with 'spn:' prefix in audience claim")
	fs.BoolVar(&o.UseAzureRMTerraformEnv, "use-azurerm-env-vars", o.UseAzureRMTerraformEnv,
		"Use environment variable names of Terraform Azure Provider (ARM_CLIENT_ID, ARM_CLIENT_SECRET, ARM_CLIENT_CERTIFICATE_PATH, ARM_CLIENT_CERTIFICATE_PASSWORD, ARM_TENANT_ID)")
	fs.BoolVar(&o.IsPoPTokenEnabled, "pop-enabled", o.IsPoPTokenEnabled, "set to true to use a PoP token for authentication or false to use a regular bearer token")
	fs.DurationVar(&o.Timeout, "timeout", 30*time.Second,
		fmt.Sprintf("Timeout duration for Azure CLI token requests. It may be specified in %s environment variable", "AZURE_CLI_TIMEOUT"))
	fs.StringVar(&o.PoPTokenClaims, "pop-claims", o.PoPTokenClaims, "contains a comma-separated list of claims to attach to the pop token in the format `key=val,key2=val2`. At minimum, specify the ARM ID of the cluster as `u=ARM_ID`")
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

	// both of the following checks ensure that --pop-enabled and --pop-claims flags are provided together
	if o.IsPoPTokenEnabled && o.PoPTokenClaims == "" {
		return fmt.Errorf("if enabling pop token mode, please provide the pop-claims flag containing the PoP token claims as a comma-separated string: `u=popClaimHost,key1=val1`")
	}

	if o.PoPTokenClaims != "" && !o.IsPoPTokenEnabled {
		return fmt.Errorf("pop-enabled flag is required to use the PoP token feature. Please provide both pop-enabled and pop-claims flags")
	}

	if o.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}

	return nil
}

func (o *Options) UpdateFromEnv() {
	o.tokenCacheFile = getCacheFileName(o)

	if o.UseAzureRMTerraformEnv {
		if v, ok := os.LookupEnv(terraformClientID); ok {
			o.ClientID = v
		}
		if v, ok := os.LookupEnv(terraformClientSecret); ok {
			o.ClientSecret = v
		}
		if v, ok := os.LookupEnv(terraformClientCertificatePath); ok {
			o.ClientCert = v
		}
		if v, ok := os.LookupEnv(terraformClientCertificatePassword); ok {
			o.ClientCertPassword = v
		}
		if v, ok := os.LookupEnv(terraformTenantID); ok {
			o.TenantID = v
		}
	} else {
		if v, ok := os.LookupEnv(kubeloginClientID); ok {
			o.ClientID = v
		}
		if v, ok := os.LookupEnv(azureClientID); ok {
			o.ClientID = v
		}
		if v, ok := os.LookupEnv(kubeloginClientSecret); ok {
			o.ClientSecret = v
		}
		if v, ok := os.LookupEnv(azureClientSecret); ok {
			o.ClientSecret = v
		}
		if v, ok := os.LookupEnv(kubeloginClientCertificatePath); ok {
			o.ClientCert = v
		}
		if v, ok := os.LookupEnv(azureClientCertificatePath); ok {
			o.ClientCert = v
		}
		if v, ok := os.LookupEnv(kubeloginClientCertificatePassword); ok {
			o.ClientCertPassword = v
		}
		if v, ok := os.LookupEnv(azureClientCertificatePassword); ok {
			o.ClientCertPassword = v
		}
		if v, ok := os.LookupEnv(azureTenantID); ok {
			o.TenantID = v
		}
	}

	if v, ok := os.LookupEnv(kubeloginROPCUsername); ok {
		o.Username = v
	}
	if v, ok := os.LookupEnv(azureUsername); ok {
		o.Username = v
	}
	if v, ok := os.LookupEnv(kubeloginROPCPassword); ok {
		o.Password = v
	}
	if v, ok := os.LookupEnv(azurePassword); ok {
		o.Password = v
	}
	if v, ok := os.LookupEnv(loginMethod); ok {
		o.LoginMethod = v
	}

	if o.LoginMethod == WorkloadIdentityLogin {
		if v, ok := os.LookupEnv(azureClientID); ok {
			o.ClientID = v
		}
		if v, ok := os.LookupEnv(azureFederatedTokenFile); ok {
			o.FederatedTokenFile = v
		}
		if v, ok := os.LookupEnv(azureAuthorityHost); ok {
			o.AuthorityHost = v
		}
	}
	if v, ok := os.LookupEnv("AZURE_CLI_TIMEOUT"); ok {
		if timeout, err := time.ParseDuration(v); err == nil {
			o.Timeout = timeout
		}
	}
}

func (o *Options) ToString() string {
	azureConfigDir := os.Getenv("AZURE_CONFIG_DIR")
	return fmt.Sprintf("Login Method: %s, Environment: %s, TenantID: %s, ServerID: %s, ClientID: %s, IsLegacy: %t, msiResourceID: %s, Timeout: %v, tokenCacheDir: %s, tokenCacheFile: %s, AZURE_CONFIG_DIR: %s",
		o.LoginMethod,
		o.Environment,
		o.TenantID,
		o.ServerID,
		o.ClientID,
		o.IsLegacy,
		o.IdentityResourceID,
		o.Timeout,
		o.TokenCacheDir,
		o.tokenCacheFile,
		azureConfigDir,
	)
}

func getCacheFileName(o *Options) string {
	// format: ${environment}-${server-id}-${client-id}-${tenant-id}[_legacy].json
	cacheFileNameFormat := "%s-%s-%s-%s.json"
	if o.IsLegacy {
		cacheFileNameFormat = "%s-%s-%s-%s_legacy.json"
	}
	return filepath.Join(o.TokenCacheDir, fmt.Sprintf(cacheFileNameFormat, o.Environment, o.ServerID, o.ClientID, o.TenantID))
}

// parsePoPClaims parses the pop token claims. Pop token claims are passed in as a
// comma-separated string in the format "key1=val1,key2=val2"
func parsePoPClaims(popClaims string) (map[string]string, error) {
	if strings.TrimSpace(popClaims) == "" {
		return nil, fmt.Errorf("failed to parse PoP token claims: no claims provided")
	}
	claimsArray := strings.Split(popClaims, ",")
	claimsMap := make(map[string]string)
	for _, claim := range claimsArray {
		claimPair := strings.Split(claim, "=")
		if len(claimPair) < 2 {
			return nil, fmt.Errorf("failed to parse PoP token claims. Ensure the claims are formatted as `key=value` with no extra whitespace")
		}
		key := strings.TrimSpace(claimPair[0])
		val := strings.TrimSpace(claimPair[1])
		if key == "" || val == "" {
			return nil, fmt.Errorf("failed to parse PoP token claims. Ensure the claims are formatted as `key=value` with no extra whitespace")
		}
		claimsMap[key] = val
	}
	if claimsMap["u"] == "" {
		return nil, fmt.Errorf("required u-claim not provided for PoP token flow. Please provide the ARM ID of the cluster in the format `u=<ARM_ID>`")
	}
	return claimsMap, nil
}

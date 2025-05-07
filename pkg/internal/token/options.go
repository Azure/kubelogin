package token

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/kubelogin/pkg/internal/env"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/util/homedir"
)

type Options struct {
	LoginMethod                string
	ClientID                   string
	ClientSecret               string
	ClientCert                 string
	ClientCertPassword         string
	Username                   string
	Password                   string
	ServerID                   string
	TenantID                   string
	Environment                string
	IsLegacy                   bool
	Timeout                    time.Duration
	AuthRecordCacheDir         string
	authRecordCacheFile        string
	IdentityResourceID         string
	FederatedTokenFile         string
	AuthorityHost              string
	UseAzureRMTerraformEnv     bool
	IsPoPTokenEnabled          bool
	PoPTokenClaims             string
	DisableEnvironmentOverride bool
	UsePersistentCache         bool
	DisableInstanceDiscovery   bool
	httpClient                 *http.Client
	RedirectURL                string
	LoginHint                  string
}

const (
	defaultEnvironmentName = "AzurePublicCloud"

	DeviceCodeLogin        = "devicecode"
	InteractiveLogin       = "interactive"
	ServicePrincipalLogin  = "spn"
	ROPCLogin              = "ropc"
	MSILogin               = "msi"
	AzureCLILogin          = "azurecli"
	AzureDeveloperCLILogin = "azd"
	WorkloadIdentityLogin  = "workloadidentity"
)

var (
	supportedLogin            []string
	DefaultAuthRecordCacheDir = homedir.HomeDir() + "/.kube/cache/kubelogin/"
)

func init() {
	supportedLogin = []string{DeviceCodeLogin, InteractiveLogin, ServicePrincipalLogin, ROPCLogin, MSILogin, AzureCLILogin, AzureDeveloperCLILogin, WorkloadIdentityLogin}
}

func GetSupportedLogins() string {
	return strings.Join(supportedLogin, ", ")
}

func NewOptions(usePersistentCache bool) Options {
	envAuthRecordCacheDir := os.Getenv("KUBECACHEDIR")
	return Options{
		LoginMethod: DeviceCodeLogin,
		Environment: defaultEnvironmentName,
		AuthRecordCacheDir: func() string {
			if envAuthRecordCacheDir != "" {
				return envAuthRecordCacheDir
			}
			return DefaultAuthRecordCacheDir
		}(),
		UsePersistentCache: usePersistentCache,
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.LoginMethod, "login", "l", o.LoginMethod,
		fmt.Sprintf("Login method. Supported methods: %s. It may be specified in %s environment variable", GetSupportedLogins(), env.LoginMethod))
	fs.StringVar(&o.ClientID, "client-id", o.ClientID,
		fmt.Sprintf("AAD client application ID. It may be specified in %s or %s environment variable", env.KubeloginClientID, env.AzureClientID))
	fs.StringVar(&o.ClientSecret, "client-secret", o.ClientSecret,
		fmt.Sprintf("AAD client application secret. Used in spn login. It may be specified in %s or %s environment variable", env.KubeloginClientSecret, env.AzureClientSecret))
	fs.StringVar(&o.ClientCert, "client-certificate", o.ClientCert,
		fmt.Sprintf("AAD client cert in pfx. Used in spn login. It may be specified in %s or %s environment variable", env.KubeloginClientCertificatePath, env.AzureClientCertificatePath))
	fs.StringVar(&o.ClientCertPassword, "client-certificate-password", o.ClientCertPassword,
		fmt.Sprintf("Password for AAD client cert. Used in spn login. It may be specified in %s or %s environment variable", env.KubeloginClientCertificatePassword, env.AzureClientCertificatePassword))
	fs.StringVar(&o.Username, "username", o.Username,
		fmt.Sprintf("user name for ropc login flow. It may be specified in %s or %s environment variable", env.KubeloginROPCUsername, env.AzureUsername))
	fs.StringVar(&o.Password, "password", o.Password,
		fmt.Sprintf("password for ropc login flow. It may be specified in %s or %s environment variable", env.KubeloginROPCPassword, env.AzurePassword))
	fs.StringVar(&o.IdentityResourceID, "identity-resource-id", o.IdentityResourceID, "Managed Identity resource id.")
	fs.StringVar(&o.ServerID, "server-id", o.ServerID, "AAD server application ID")
	fs.StringVar(&o.FederatedTokenFile, "federated-token-file", o.FederatedTokenFile,
		fmt.Sprintf("Workload Identity federated token file. It may be specified in %s environment variable", env.AzureFederatedTokenFile))
	fs.StringVar(&o.AuthorityHost, "authority-host", o.AuthorityHost,
		fmt.Sprintf("Workload Identity authority host. It may be specified in %s environment variable", env.AzureAuthorityHost))
	fs.StringVar(&o.AuthRecordCacheDir, "token-cache-dir", o.AuthRecordCacheDir, "directory to cache authentication record")
	_ = fs.MarkDeprecated("token-cache-dir", "use --cache-dir instead")
	fs.StringVar(&o.AuthRecordCacheDir, "cache-dir", o.AuthRecordCacheDir, "directory to cache authentication record")
	fs.StringVarP(&o.TenantID, "tenant-id", "t", o.TenantID, fmt.Sprintf("AAD tenant ID. It may be specified in %s environment variable", env.AzureTenantID))
	fs.StringVarP(&o.Environment, "environment", "e", o.Environment, "Azure environment name")
	fs.BoolVar(&o.IsLegacy, "legacy", o.IsLegacy, "set to true to get token with 'spn:' prefix in audience claim")
	fs.BoolVar(&o.UseAzureRMTerraformEnv, "use-azurerm-env-vars", o.UseAzureRMTerraformEnv,
		"Use environment variable names of Terraform Azure Provider (ARM_CLIENT_ID, ARM_CLIENT_SECRET, ARM_CLIENT_CERTIFICATE_PATH, ARM_CLIENT_CERTIFICATE_PASSWORD, ARM_TENANT_ID)")
	fs.BoolVar(&o.IsPoPTokenEnabled, "pop-enabled", o.IsPoPTokenEnabled, "set to true to use a PoP token for authentication or false to use a regular bearer token")
	fs.DurationVar(&o.Timeout, "timeout", 60*time.Second,
		fmt.Sprintf("Timeout duration for Azure CLI token requests. It may be specified in %s environment variable", "AZURE_CLI_TIMEOUT"))
	fs.StringVar(&o.PoPTokenClaims, "pop-claims", o.PoPTokenClaims, "contains a comma-separated list of claims to attach to the pop token in the format `key=val,key2=val2`. At minimum, specify the ARM ID of the cluster as `u=ARM_ID`")
	fs.BoolVar(&o.DisableEnvironmentOverride, "disable-environment-override", o.DisableEnvironmentOverride, "Enable or disable the use of env-variables. Default false")
	fs.BoolVar(&o.DisableInstanceDiscovery, "disable-instance-discovery", o.DisableInstanceDiscovery, "set to true to disable instance discovery in environments with their own simple Identity Provider (not AAD) that do not have instance metadata discovery endpoint. Default false")
	fs.StringVar(&o.RedirectURL, "redirect-url", o.RedirectURL, "The URL Microsoft Entra ID will redirect to with the access token. This is only used for interactive login. This is an optional parameter.")
	fs.StringVar(&o.LoginHint, "login-hint", o.LoginHint, "The login hint to pre-fill the username in the interactive login flow.")
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

	if o.AuthorityHost != "" {
		u, err := url.ParseRequestURI(o.AuthorityHost)
		if err != nil {
			return fmt.Errorf("authority host %q is not valid: %s", o.AuthorityHost, err)
		}
		if u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("authority host %q is not valid", o.AuthorityHost)
		}
		if !strings.HasSuffix(o.AuthorityHost, "/") {
			return fmt.Errorf("authority host %q should have a trailing slash", o.AuthorityHost)
		}
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
	o.authRecordCacheFile = getAuthenticationRecordFileName(o)

	if o.DisableEnvironmentOverride {
		return
	}

	if o.UseAzureRMTerraformEnv {
		if v, ok := os.LookupEnv(env.TerraformClientID); ok {
			o.ClientID = v
		}
		if v, ok := os.LookupEnv(env.TerraformClientSecret); ok {
			o.ClientSecret = v
		}
		if v, ok := os.LookupEnv(env.TerraformClientCertificatePath); ok {
			o.ClientCert = v
		}
		if v, ok := os.LookupEnv(env.TerraformClientCertificatePassword); ok {
			o.ClientCertPassword = v
		}
		if v, ok := os.LookupEnv(env.TerraformTenantID); ok {
			o.TenantID = v
		}
	} else {
		if v, ok := os.LookupEnv(env.KubeloginClientID); ok {
			o.ClientID = v
		}
		if v, ok := os.LookupEnv(env.AzureClientID); ok {
			o.ClientID = v
		}
		if v, ok := os.LookupEnv(env.KubeloginClientSecret); ok {
			o.ClientSecret = v
		}
		if v, ok := os.LookupEnv(env.AzureClientSecret); ok {
			o.ClientSecret = v
		}
		if v, ok := os.LookupEnv(env.KubeloginClientCertificatePath); ok {
			o.ClientCert = v
		}
		if v, ok := os.LookupEnv(env.AzureClientCertificatePath); ok {
			o.ClientCert = v
		}
		if v, ok := os.LookupEnv(env.KubeloginClientCertificatePassword); ok {
			o.ClientCertPassword = v
		}
		if v, ok := os.LookupEnv(env.AzureClientCertificatePassword); ok {
			o.ClientCertPassword = v
		}
		if v, ok := os.LookupEnv(env.AzureTenantID); ok {
			o.TenantID = v
		}
	}

	if v, ok := os.LookupEnv(env.KubeloginROPCUsername); ok {
		o.Username = v
	}
	if v, ok := os.LookupEnv(env.AzureUsername); ok {
		o.Username = v
	}
	if v, ok := os.LookupEnv(env.KubeloginROPCPassword); ok {
		o.Password = v
	}
	if v, ok := os.LookupEnv(env.AzurePassword); ok {
		o.Password = v
	}
	if v, ok := os.LookupEnv(env.LoginMethod); ok {
		o.LoginMethod = v
	}

	if o.LoginMethod == WorkloadIdentityLogin {
		if v, ok := os.LookupEnv(env.AzureClientID); ok {
			o.ClientID = v
		}
		if v, ok := os.LookupEnv(env.AzureFederatedTokenFile); ok {
			o.FederatedTokenFile = v
		}
		if v, ok := os.LookupEnv(env.AzureAuthorityHost); ok {
			o.AuthorityHost = v
		}
	}
	if v, ok := os.LookupEnv("AZURE_CLI_TIMEOUT"); ok {
		if timeout, err := time.ParseDuration(v); err == nil {
			o.Timeout = timeout
		}
	}
}

func (o *Options) GetCloudConfiguration() cloud.Configuration {
	if o.AuthorityHost != "" {
		return cloud.Configuration{
			ActiveDirectoryAuthorityHost: o.AuthorityHost,
		}
	}

	switch strings.ToUpper(o.Environment) {
	case "AZURECLOUD":
		fallthrough
	case "AZUREPUBLIC":
		fallthrough
	case "AZUREPUBLICCLOUD":
		return cloud.AzurePublic
	case "AZUREUSGOVERNMENT":
		fallthrough
	case "AZUREUSGOVERNMENTCLOUD":
		return cloud.AzureGovernment
	case "AZURECHINACLOUD":
		return cloud.AzureChina
	}
	return cloud.AzurePublic
}

func (o *Options) ToString() string {
	azureConfigDir := os.Getenv("AZURE_CONFIG_DIR")

	parts := []string{
		fmt.Sprintf("Login Method: %s", o.LoginMethod),
		fmt.Sprintf("Environment: %s", o.Environment),
		fmt.Sprintf("TenantID: %s", o.TenantID),
		fmt.Sprintf("ServerID: %s", o.ServerID),
		fmt.Sprintf("ClientID: %s", o.ClientID),
		fmt.Sprintf("IsLegacy: %t", o.IsLegacy),
		fmt.Sprintf("msiResourceID: %s", o.IdentityResourceID),
		fmt.Sprintf("Timeout: %v", o.Timeout),
		fmt.Sprintf("authRecordCacheDir: %s", o.AuthRecordCacheDir),
		fmt.Sprintf("tokenauthRecordFile: %s", o.authRecordCacheFile),
		fmt.Sprintf("AZURE_CONFIG_DIR: %s", azureConfigDir),
		fmt.Sprintf("RedirectURL: %s", o.RedirectURL),
		fmt.Sprintf("LoginHint: %s", o.LoginHint),
	}

	return strings.Join(parts, ", ")
}

func getAuthenticationRecordFileName(o *Options) string {
	return filepath.Join(o.AuthRecordCacheDir, "auth.json")
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

func (o *Options) AddCompletions(cmd *cobra.Command) {
	_ = cmd.RegisterFlagCompletionFunc("login", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return supportedLogin, cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.MarkFlagFilename("client-certificate", "pfx", "cert")
	_ = cmd.MarkFlagFilename("federated-token-file", "")
	_ = cmd.MarkFlagDirname("token-cache-dir")

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		// Set a default completion function if none was set. We don't look
		// up if it does already have one set, because Cobra does this for
		// us, and returns an error (which we ignore for this reason).
		_ = cmd.RegisterFlagCompletionFunc(flag.Name, cobra.NoFileCompletions)
	})
}

package options

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/Azure/kubelogin/pkg/internal/env"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/util/homedir"
)

// CommandType represents the type of command being executed
type CommandType int

const (
	ConvertCommand CommandType = iota
	TokenCommand
)

// UnifiedOptions represents the centralized options struct for all kubelogin commands
type UnifiedOptions struct {
	// Core authentication options
	LoginMethod  string `flag:"login,l" env:"AAD_LOGIN_METHOD" validate:"required,oneof=devicecode interactive spn ropc msi azurecli azd workloadidentity" description:"Login method. Supported methods: devicecode, interactive, spn, ropc, msi, azurecli, azd, workloadidentity. It may be specified in AAD_LOGIN_METHOD environment variable"`
	ClientID     string `flag:"client-id" env:"AZURE_CLIENT_ID,AAD_SERVICE_PRINCIPAL_CLIENT_ID" description:"AAD client application ID. It may be specified in AAD_SERVICE_PRINCIPAL_CLIENT_ID or AZURE_CLIENT_ID environment variable"`
	ClientSecret string `flag:"client-secret" env:"AZURE_CLIENT_SECRET,AAD_SERVICE_PRINCIPAL_CLIENT_SECRET" sensitive:"true" description:"AAD client application secret. Used in spn login. It may be specified in AAD_SERVICE_PRINCIPAL_CLIENT_SECRET or AZURE_CLIENT_SECRET environment variable"`
	TenantID     string `flag:"tenant-id,t" env:"AZURE_TENANT_ID" validate:"required" description:"AAD tenant ID. It may be specified in AZURE_TENANT_ID environment variable"`
	ServerID     string `flag:"server-id" description:"AAD server application ID"`

	// Certificate options
	ClientCert         string `flag:"client-certificate" env:"AZURE_CLIENT_CERTIFICATE_PATH,AAD_SERVICE_PRINCIPAL_CLIENT_CERTIFICATE" description:"AAD client cert in pfx. Used in spn login. It may be specified in AAD_SERVICE_PRINCIPAL_CLIENT_CERTIFICATE or AZURE_CLIENT_CERTIFICATE_PATH environment variable"`
	ClientCertPassword string `flag:"client-certificate-password" env:"AZURE_CLIENT_CERTIFICATE_PASSWORD,AAD_SERVICE_PRINCIPAL_CLIENT_CERTIFICATE_PASSWORD" sensitive:"true" description:"Password for AAD client cert. Used in spn login. It may be specified in AAD_SERVICE_PRINCIPAL_CLIENT_CERTIFICATE_PASSWORD or AZURE_CLIENT_CERTIFICATE_PASSWORD environment variable"`

	// ROPC options
	Username string `flag:"username" env:"AZURE_USERNAME,AAD_USER_PRINCIPAL_NAME" description:"user name for ropc login flow. It may be specified in AAD_USER_PRINCIPAL_NAME or AZURE_USERNAME environment variable"`
	Password string `flag:"password" env:"AZURE_PASSWORD,AAD_USER_PRINCIPAL_PASSWORD" sensitive:"true" description:"password for ropc login flow. It may be specified in AAD_USER_PRINCIPAL_PASSWORD or AZURE_PASSWORD environment variable"`

	// MSI options
	IdentityResourceID string `flag:"identity-resource-id" description:"Managed Identity resource id."`

	// Workload Identity options
	FederatedTokenFile string `flag:"federated-token-file" env:"AZURE_FEDERATED_TOKEN_FILE" description:"Workload Identity federated token file. It may be specified in AZURE_FEDERATED_TOKEN_FILE environment variable"`
	AuthorityHost      string `flag:"authority-host" env:"AZURE_AUTHORITY_HOST" validate:"omitempty,url" description:"Workload Identity authority host. It may be specified in AZURE_AUTHORITY_HOST environment variable"`

	// Advanced options
	Environment                string        `flag:"environment,e" env:"AZURE_ENVIRONMENT" default:"AzurePublicCloud" description:"Azure environment name"`
	IsLegacy                   bool          `flag:"legacy" description:"set to true to get token with 'spn:' prefix in audience claim"`
	Timeout                    time.Duration `flag:"timeout" env:"AZURE_CLI_TIMEOUT" default:"60s" description:"Timeout duration for Azure CLI token requests. It may be specified in AZURE_CLI_TIMEOUT environment variable"`
	UseAzureRMTerraformEnv     bool          `flag:"use-azurerm-env-vars" description:"Use environment variable names of Terraform Azure Provider (ARM_CLIENT_ID, ARM_CLIENT_SECRET, ARM_CLIENT_CERTIFICATE_PATH, ARM_CLIENT_CERTIFICATE_PASSWORD, ARM_TENANT_ID)"`
	IsPoPTokenEnabled          bool          `flag:"pop-enabled" description:"set to true to use a PoP token for authentication or false to use a regular bearer token"`
	PoPTokenClaims             string        `flag:"pop-claims" description:"contains a comma-separated list of claims to attach to the pop token in the format \u0060key=val,key2=val2\u0060. At minimum, specify the ARM ID of the cluster as \u0060u=ARM_ID\u0060"`
	DisableEnvironmentOverride bool          `flag:"disable-environment-override" description:"Enable or disable the use of env-variables. Default false"`
	DisableInstanceDiscovery   bool          `flag:"disable-instance-discovery" description:"set to true to disable instance discovery in environments with their own simple Identity Provider (not AAD) that do not have instance metadata discovery endpoint. Default false"`
	RedirectURL                string        `flag:"redirect-url" description:"The URL Microsoft Entra ID will redirect to with the access token. This is only used for interactive login. This is an optional parameter."`
	LoginHint                  string        `flag:"login-hint" description:"The login hint to pre-fill the username in the interactive login flow."`

	// Cache options
	AuthRecordCacheDir string `flag:"cache-dir" env:"KUBECACHEDIR" description:"directory to cache authentication record"`
	UsePersistentCache bool   `description:"Use persistent cache"`

	// Converter-specific options
	Context        string `flag:"context" commands:"convert" description:"The name of the kubeconfig context to use"`
	AzureConfigDir string `flag:"azure-config-dir" commands:"convert" description:"Azure CLI config path"`
	Kubeconfig     string `flag:"kubeconfig" commands:"convert" description:"Path to the kubeconfig file to use for CLI requests."`

	// Internal fields
	flags   *pflag.FlagSet
	command CommandType
}

// NewUnifiedOptions creates a new UnifiedOptions instance for the specified command type
func NewUnifiedOptions(cmdType CommandType) *UnifiedOptions {
	envAuthRecordCacheDir := os.Getenv("KUBECACHEDIR")
	defaultCacheDir := homedir.HomeDir() + "/.kube/cache/kubelogin/"
	if envAuthRecordCacheDir != "" {
		defaultCacheDir = envAuthRecordCacheDir
	}

	opts := &UnifiedOptions{
		LoginMethod:        "devicecode",
		Environment:        "AzurePublicCloud",
		Timeout:            60 * time.Second,
		AuthRecordCacheDir: defaultCacheDir,
		UsePersistentCache: cmdType == TokenCommand,
		command:            cmdType,
	}

	return opts
}

// isSet checks if a flag was explicitly set by the user (vs. having a default value)
func (o *UnifiedOptions) isSet(flagName string) bool {
	if o.flags == nil {
		return false
	}
	found := false
	o.flags.Visit(func(f *pflag.Flag) {
		if f.Name == flagName {
			found = true
		}
	})
	return found
}

// RegisterFlags automatically registers CLI flags based on struct tags
func (o *UnifiedOptions) RegisterFlags(fs *pflag.FlagSet) error {
	o.flags = fs

	val := reflect.ValueOf(o).Elem()
	typ := reflect.TypeOf(o).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields and internal fields
		if !field.CanSet() || strings.HasPrefix(fieldType.Name, "flags") || strings.HasPrefix(fieldType.Name, "command") {
			continue
		}

		flagTag := fieldType.Tag.Get("flag")
		if flagTag == "" {
			continue
		}

		// Check if this field should be included for the current command
		commandsTag := fieldType.Tag.Get("commands")
		if commandsTag != "" {
			allowedCommands := strings.Split(commandsTag, ",")
			currentCmd := "convert"
			if o.command == TokenCommand {
				currentCmd = "token"
			}

			found := false
			for _, cmd := range allowedCommands {
				if strings.TrimSpace(cmd) == currentCmd {
					found = true
					break
				}
			}
			if !found {
				continue // Skip this flag for this command
			}
		}

		description := fieldType.Tag.Get("description")
		defaultTag := fieldType.Tag.Get("default")

		// Parse flag name and shorthand
		flagParts := strings.Split(flagTag, ",")
		flagName := flagParts[0]
		var shorthand string
		if len(flagParts) > 1 {
			shorthand = flagParts[1]
		}

		// Register flag based on field type
		switch field.Kind() {
		case reflect.String:
			defaultVal := defaultTag
			if defaultVal == "" {
				defaultVal = field.String()
			}
			if shorthand != "" {
				fs.StringVarP(field.Addr().Interface().(*string), flagName, shorthand, defaultVal, description)
			} else {
				fs.StringVar(field.Addr().Interface().(*string), flagName, defaultVal, description)
			}

		case reflect.Bool:
			defaultVal := false
			if defaultTag == "true" {
				defaultVal = true
			} else if field.Bool() {
				defaultVal = true
			}
			fs.BoolVar(field.Addr().Interface().(*bool), flagName, defaultVal, description)

		case reflect.Int64:
			if field.Type() == reflect.TypeOf(time.Duration(0)) {
				defaultVal := time.Duration(0)
				if defaultTag != "" {
					if parsed, err := time.ParseDuration(defaultTag); err == nil {
						defaultVal = parsed
					}
				} else {
					defaultVal = field.Interface().(time.Duration)
				}
				fs.DurationVar(field.Addr().Interface().(*time.Duration), flagName, defaultVal, description)
			}
		}
	}

	// Add deprecated flags for backward compatibility
	// token-cache-dir is deprecated in favor of cache-dir
	fs.StringVar(&o.AuthRecordCacheDir, "token-cache-dir", o.AuthRecordCacheDir, "directory to cache authentication record")
	_ = fs.MarkDeprecated("token-cache-dir", "use --cache-dir instead")

	return nil
}

// LoadFromEnv loads values from environment variables based on struct tags
func (o *UnifiedOptions) LoadFromEnv() {
	if o.DisableEnvironmentOverride {
		return
	}

	val := reflect.ValueOf(o).Elem()
	typ := reflect.TypeOf(o).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			continue
		}

		// Handle multiple environment variable names
		envVars := strings.Split(envTag, ",")
		for _, envVar := range envVars {
			envVar = strings.TrimSpace(envVar)
			if value, ok := os.LookupEnv(envVar); ok {
				switch field.Kind() {
				case reflect.String:
					// Always override with environment variable if it exists
					field.SetString(value)
				case reflect.Bool:
					if value == "true" {
						field.SetBool(true)
					}
				case reflect.Int64:
					if field.Type() == reflect.TypeOf(time.Duration(0)) {
						if duration, err := time.ParseDuration(value); err == nil {
							field.Set(reflect.ValueOf(duration))
						}
					}
				}
				break // Use first found environment variable
			}
		}
	}

	// Handle Terraform environment variables if enabled
	if o.UseAzureRMTerraformEnv {
		o.loadTerraformEnvVars()
	}
}

// loadTerraformEnvVars loads Terraform Azure Provider environment variables
func (o *UnifiedOptions) loadTerraformEnvVars() {
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
}

// ExecuteCommand executes the appropriate command based on the command type
func (o *UnifiedOptions) ExecuteCommand(ctx context.Context, flags *pflag.FlagSet) error {
	o.flags = flags
	o.LoadFromEnv()

	switch o.command {
	case ConvertCommand:
		// Convert command handles its own validation after field extraction
		return o.executeConvert()
	case TokenCommand:
		// Token command handles its own validation and execution
		return o.executeToken(ctx)
	default:
		return fmt.Errorf("unknown command type: %v", o.command)
	}
}

// RegisterCompletions registers shell completion functions for flags
func (o *UnifiedOptions) RegisterCompletions(cmd *cobra.Command) error {
	// Register completion for login method
	_ = cmd.RegisterFlagCompletionFunc("login", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"devicecode", "interactive", "spn", "ropc", "msi", "azurecli", "azd", "workloadidentity"}, cobra.ShellCompDirectiveNoFileComp
	})

	// Mark file/directory completions
	_ = cmd.MarkFlagFilename("client-certificate", "pfx", "cert")
	_ = cmd.MarkFlagFilename("federated-token-file", "")
	_ = cmd.MarkFlagDirname("cache-dir")
	_ = cmd.MarkFlagDirname("azure-config-dir")

	// Set default completion for all other flags
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		_ = cmd.RegisterFlagCompletionFunc(flag.Name, cobra.NoFileCompletions)
	})

	return nil
}

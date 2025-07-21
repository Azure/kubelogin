package options

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	klog "k8s.io/klog/v2"

	"github.com/Azure/kubelogin/pkg/internal/converter"
	"github.com/Azure/kubelogin/pkg/internal/token"
)

// executeConvert executes the convert-kubeconfig command
func (o *UnifiedOptions) executeConvert() error {
	// Use direct conversion when supported, fallback to legacy for backward compatibility
	pathOptions := clientcmd.NewDefaultPathOptions()
	if o.flags != nil {
		pathOptions.LoadingRules.ExplicitPath, _ = o.flags.GetString("kubeconfig")
	}

	// Direct conversion approach - eliminate legacy bridge
	return o.ConvertKubeconfig(pathOptions)
}

// ConvertKubeconfig performs direct kubeconfig conversion using unified options
func (o *UnifiedOptions) ConvertKubeconfig(pathOptions *clientcmd.PathOptions) error {
	klog.V(5).Info(o.ToString())

	// Load kubeconfig using clientcmd
	config, kubeconfigs, err := o.loadKubeConfig(pathOptions)
	if err != nil {
		return fmt.Errorf("unable to load kubeconfig: %w", err)
	}

	klog.V(5).Infof("Loading kubeconfig from %s", strings.Join(kubeconfigs, ":"))

	// Determine target context
	targetAuthInfo := ""
	if o.Context != "" {
		if config.Contexts[o.Context] == nil {
			return fmt.Errorf("no context exists with the name: %q", o.Context)
		}
		targetAuthInfo = config.Contexts[o.Context].AuthInfo
	}

	// Process each auth info in the kubeconfig
	for name, authInfo := range config.AuthInfos {
		if targetAuthInfo != "" && name != targetAuthInfo {
			continue
		}

		klog.V(5).Infof("context: %q", name)

		// Skip if not using kubelogin or legacy azure auth
		if !o.isExecUsingKubelogin(authInfo) && !o.isLegacyAzureAuth(authInfo) {
			continue
		}

		klog.V(5).Info("converting...")

		// Build exec config using reflection and struct tags
		execConfig, err := o.buildExecConfig(authInfo)
		if err != nil {
			return fmt.Errorf("failed to build exec config for %s: %w", name, err)
		}

		// Create a temporary copy of options with extracted values for final validation
		tempOptions := *o
		tempOptions.populateFromExecConfig(execConfig)

		// Validate the complete configuration after extraction
		// This ensures we have all required fields either from user input or kubeconfig
		if err := tempOptions.ValidateForConversion(); err != nil {
			return fmt.Errorf("conversion failed for %s: %w", name, err)
		}

		// Update auth info with new exec config
		authInfo.Exec = execConfig
		authInfo.AuthProvider = nil
	}

	// Save the updated kubeconfig
	return clientcmd.ModifyConfig(pathOptions, *config, true)
}

// executeToken executes the get-token command
func (o *UnifiedOptions) executeToken(ctx context.Context) error {
	// Validate options first - strict validation for immediate token execution
	if err := o.ValidateForTokenExecution(); err != nil {
		return err
	}

	// Create credential directly using existing token infrastructure
	credential, err := token.NewAzIdentityCredential(azidentity.AuthenticationRecord{}, o.ToTokenOptions())
	if err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}

	// Create and execute the token plugin
	plugin := &tokenPlugin{
		credential: credential,
		options:    o.ToTokenOptions(),
	}

	return plugin.Do(ctx)
}

// tokenPlugin implements the token retrieval logic using credential builders
type tokenPlugin struct {
	credential token.CredentialProvider
	options    *token.Options
}

// Do executes the token retrieval and outputs the result
func (p *tokenPlugin) Do(ctx context.Context) error {
	// Convert to legacy token options and use existing token.New logic for now
	// This maintains compatibility while using the new credential builder system
	legacyPlugin, err := token.New(p.options)
	if err != nil {
		return err
	}

	return legacyPlugin.Do(ctx)
}

// ToTokenOptions converts UnifiedOptions to legacy token.Options for backward compatibility
func (o *UnifiedOptions) ToTokenOptions() *token.Options {
	return &token.Options{
		LoginMethod:                o.LoginMethod,
		ClientID:                   o.ClientID,
		ClientSecret:               o.ClientSecret,
		ClientCert:                 o.ClientCert,
		ClientCertPassword:         o.ClientCertPassword,
		Username:                   o.Username,
		Password:                   o.Password,
		ServerID:                   o.ServerID,
		TenantID:                   o.TenantID,
		Environment:                o.Environment,
		IsLegacy:                   o.IsLegacy,
		Timeout:                    o.Timeout,
		AuthRecordCacheDir:         o.AuthRecordCacheDir,
		IdentityResourceID:         o.IdentityResourceID,
		FederatedTokenFile:         o.FederatedTokenFile,
		AuthorityHost:              o.AuthorityHost,
		UseAzureRMTerraformEnv:     o.UseAzureRMTerraformEnv,
		IsPoPTokenEnabled:          o.IsPoPTokenEnabled,
		PoPTokenClaims:             o.PoPTokenClaims,
		DisableEnvironmentOverride: o.DisableEnvironmentOverride,
		UsePersistentCache:         o.UsePersistentCache,
		DisableInstanceDiscovery:   o.DisableInstanceDiscovery,
		RedirectURL:                o.RedirectURL,
		LoginHint:                  o.LoginHint,
	}
}

// ToConverterOptions converts UnifiedOptions to legacy converter.Options for backward compatibility
func (o *UnifiedOptions) ToConverterOptions() *converter.Options {
	// Create converter options with embedded token options
	converterOpts := converter.New()
	converterOpts.TokenOptions = *o.ToTokenOptions()
	converterOpts.Flags = o.flags

	// Use reflection to set private fields since they don't have setters
	val := reflect.ValueOf(&converterOpts).Elem()
	if contextField := val.FieldByName("context"); contextField.IsValid() && contextField.CanSet() {
		contextField.SetString(o.Context)
	}
	if azureConfigDirField := val.FieldByName("azureConfigDir"); azureConfigDirField.IsValid() && azureConfigDirField.CanSet() {
		azureConfigDirField.SetString(o.AzureConfigDir)
	}

	return &converterOpts
}

// Constants for exec configuration
const (
	execName        = "kubelogin"
	getTokenCommand = "get-token"
	execAPIVersion  = "client.authentication.k8s.io/v1beta1"
	execInstallHint = `
kubelogin is not installed which is required to connect to AAD enabled cluster.

To learn more, please go to https://azure.github.io/kubelogin/
`
	azureAuthProvider = "azure"
	azureConfigDir    = "AZURE_CONFIG_DIR"

	// Field name constants for shouldIncludeArg
	fieldNameClientID    = "ClientID"
	fieldNameTenantID    = "TenantID"
	fieldNameEnvironment = "Environment"

	// Login method constants - use token package constants
	loginMethodSPN              = token.ServicePrincipalLogin
	loginMethodDeviceCode       = token.DeviceCodeLogin
	loginMethodInteractive      = token.InteractiveLogin
	loginMethodROPC             = token.ROPCLogin
	loginMethodMSI              = token.MSILogin
	loginMethodWorkloadIdentity = token.WorkloadIdentityLogin
	loginMethodAzureCLI         = token.AzureCLILogin
	loginMethodAzd              = token.AzureDeveloperCLILogin
)

// loadKubeConfig loads the kubeconfig using clientcmd and returns config and file paths
func (o *UnifiedOptions) loadKubeConfig(pathOptions *clientcmd.PathOptions) (*api.Config, []string, error) {
	// Create config flags for loading kubeconfig
	configFlags := &genericclioptions.ConfigFlags{
		KubeConfig: func() *string { s := ""; return &s }(),
	}

	// Set kubeconfig path if specified
	if o.Kubeconfig != "" {
		configFlags.KubeConfig = &o.Kubeconfig
	}

	clientConfig := configFlags.ToRawKubeConfigLoader()
	var kubeconfigs []string

	if clientConfig.ConfigAccess() != nil {
		if clientConfig.ConfigAccess().GetExplicitFile() != "" {
			kubeconfigs = append(kubeconfigs, clientConfig.ConfigAccess().GetExplicitFile())
		} else {
			kubeconfigs = append(kubeconfigs, clientConfig.ConfigAccess().GetLoadingPrecedence()...)
		}
	}

	config, err := clientConfig.RawConfig()
	return &config, kubeconfigs, err
}

// isLegacyAzureAuth checks if auth info uses legacy azure auth provider
func (o *UnifiedOptions) isLegacyAzureAuth(authInfo *api.AuthInfo) bool {
	if authInfo == nil || authInfo.AuthProvider == nil {
		return false
	}
	return authInfo.AuthProvider.Name == azureAuthProvider
}

// isExecUsingKubelogin checks if auth info uses kubelogin exec
func (o *UnifiedOptions) isExecUsingKubelogin(authInfo *api.AuthInfo) bool {
	if authInfo == nil || authInfo.Exec == nil {
		return false
	}
	return authInfo.Exec.Command == execName || strings.Contains(authInfo.Exec.Command, execName)
}

// buildExecConfig creates exec config using reflection and struct tags
func (o *UnifiedOptions) buildExecConfig(authInfo *api.AuthInfo) (*api.ExecConfig, error) {
	exec := &api.ExecConfig{
		Command:     execName,
		Args:        []string{getTokenCommand},
		APIVersion:  execAPIVersion,
		InstallHint: execInstallHint,
	}

	// Preserve any existing install hint
	if authInfo.Exec != nil && authInfo.Exec.InstallHint != "" {
		exec.InstallHint = authInfo.Exec.InstallHint
	}

	// Add login method (always required)
	exec.Args = append(exec.Args, "--login", o.LoginMethod)

	// Get existing values from authInfo for fields not explicitly set
	existingValues := o.extractExistingValues(authInfo)

	// Use reflection to automatically process all fields
	val := reflect.ValueOf(o).Elem()
	typ := reflect.TypeOf(o).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		flagTag := field.Tag.Get("flag")
		if flagTag == "" || o.isConverterOnlyField(field.Name) {
			continue
		}

		// Extract flag name (before comma if exists)
		flagName := strings.Split(flagTag, ",")[0]
		argName := "--" + flagName

		// Get field value, with fallback to existing values
		fieldVal := val.Field(i)
		stringVal := o.fieldToString(fieldVal)

		// Use existing value if current value is empty
		if stringVal == "" {
			if existingVal, exists := existingValues[argName]; exists {
				stringVal = existingVal
			}
		}

		// Add to args if not empty and should be included
		if stringVal != "" && o.shouldIncludeArg(field.Name, stringVal) {
			exec.Args = append(exec.Args, argName, stringVal)
		}
	}

	// Handle boolean flags (no value, just presence)
	if o.IsLegacy || o.shouldPreserveLegacyFlag(authInfo) {
		exec.Args = append(exec.Args, "--legacy")
	}
	if o.IsPoPTokenEnabled {
		exec.Args = append(exec.Args, "--pop-enabled")
	}
	if o.DisableEnvironmentOverride {
		exec.Args = append(exec.Args, "--disable-environment-override")
	}
	if o.DisableInstanceDiscovery {
		exec.Args = append(exec.Args, "--disable-instance-discovery")
	}

	// Add environment variables for specific cases
	exec.Env = o.buildEnvVars()

	return exec, nil
}

// extractExistingValues extracts existing argument values from authInfo
func (o *UnifiedOptions) extractExistingValues(authInfo *api.AuthInfo) map[string]string {
	existing := make(map[string]string)

	// Extract from legacy auth provider
	if o.isLegacyAzureAuth(authInfo) {
		if config := authInfo.AuthProvider.Config; config != nil {
			if val, ok := config["client-id"]; ok {
				existing["--client-id"] = val
			}
			if val, ok := config["apiserver-id"]; ok {
				existing["--server-id"] = val
			}
			if val, ok := config["tenant-id"]; ok {
				existing["--tenant-id"] = val
			}
			if val, ok := config["environment"]; ok {
				existing["--environment"] = val
			}
		}
	}

	// Extract from existing exec args
	if authInfo.Exec != nil && authInfo.Exec.Args != nil {
		for i := 0; i < len(authInfo.Exec.Args)-1; i++ {
			arg := authInfo.Exec.Args[i]
			if strings.HasPrefix(arg, "--") && i+1 < len(authInfo.Exec.Args) {
				next := authInfo.Exec.Args[i+1]
				if !strings.HasPrefix(next, "--") {
					existing[arg] = next
				}
			}
		}
	}

	return existing
}

// populateFromExecConfig updates the options with values from the exec config
// This is used for validation after extraction to ensure all required fields are present
func (o *UnifiedOptions) populateFromExecConfig(exec *api.ExecConfig) {
	if exec == nil || exec.Args == nil {
		return
	}

	// Parse the exec args to populate options
	for i := 0; i < len(exec.Args)-1; i++ {
		arg := exec.Args[i]
		if !strings.HasPrefix(arg, "--") || i+1 >= len(exec.Args) {
			continue
		}

		value := exec.Args[i+1]
		if strings.HasPrefix(value, "--") {
			continue // Next arg is another flag, skip
		}

		switch arg {
		case "--client-id":
			if o.ClientID == "" {
				o.ClientID = value
			}
		case "--tenant-id":
			if o.TenantID == "" {
				o.TenantID = value
			}
		case "--server-id":
			if o.ServerID == "" {
				o.ServerID = value
			}
		case "--environment":
			if o.Environment == "" {
				o.Environment = value
			}
		case "--client-secret":
			if o.ClientSecret == "" {
				o.ClientSecret = value
			}
		case "--client-certificate":
			if o.ClientCert == "" {
				o.ClientCert = value
			}
		case "--client-certificate-password":
			if o.ClientCertPassword == "" {
				o.ClientCertPassword = value
			}
		case "--username":
			if o.Username == "" {
				o.Username = value
			}
		case "--password":
			if o.Password == "" {
				o.Password = value
			}
		case "--identity-resource-id":
			if o.IdentityResourceID == "" {
				o.IdentityResourceID = value
			}
		case "--authority-host":
			if o.AuthorityHost == "" {
				o.AuthorityHost = value
			}
		case "--federated-token-file":
			if o.FederatedTokenFile == "" {
				o.FederatedTokenFile = value
			}
		case "--cache-dir":
			if o.AuthRecordCacheDir == "" {
				o.AuthRecordCacheDir = value
			}
		case "--pop-claims":
			if o.PoPTokenClaims == "" {
				o.PoPTokenClaims = value
			}
		case "--redirect-url":
			if o.RedirectURL == "" {
				o.RedirectURL = value
			}
		case "--login-hint":
			if o.LoginHint == "" {
				o.LoginHint = value
			}
		case "--login":
			if o.LoginMethod == "" {
				o.LoginMethod = value
			}
		}
	}

	// Handle boolean flags
	for _, arg := range exec.Args {
		switch arg {
		case "--legacy":
			if !o.IsLegacy {
				o.IsLegacy = true
			}
		case "--pop-enabled":
			if !o.IsPoPTokenEnabled {
				o.IsPoPTokenEnabled = true
			}
		case "--disable-environment-override":
			if !o.DisableEnvironmentOverride {
				o.DisableEnvironmentOverride = true
			}
		case "--disable-instance-discovery":
			if !o.DisableInstanceDiscovery {
				o.DisableInstanceDiscovery = true
			}
		}
	}
}

// fieldToString converts a field value to string representation
func (o *UnifiedOptions) fieldToString(fieldVal reflect.Value) string {
	switch fieldVal.Kind() {
	case reflect.String:
		return fieldVal.String()
	case reflect.Bool:
		if fieldVal.Bool() {
			return trueValue
		}
		return ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if fieldVal.Int() != 0 {
			return fmt.Sprintf("%d", fieldVal.Int())
		}
		return ""
	default:
		if !fieldVal.IsZero() {
			return fmt.Sprintf("%v", fieldVal.Interface())
		}
		return ""
	}
}

// isConverterOnlyField checks if a field should be excluded from exec args
func (o *UnifiedOptions) isConverterOnlyField(fieldName string) bool {
	converterOnlyFields := map[string]bool{
		"Context":        true,
		"AzureConfigDir": true,
		"flags":          true,
		"command":        true,
	}
	return converterOnlyFields[fieldName]
}

// shouldIncludeArg determines if an argument should be included based on login method
func (o *UnifiedOptions) shouldIncludeArg(fieldName, value string) bool {
	// Always include server ID for all login methods
	if fieldName == "ServerID" {
		return true
	}

	// Include based on login method requirements
	switch o.LoginMethod {
	case loginMethodSPN:
		// For SPN, include client secret OR client certificate, but not both
		if fieldName == "ClientSecret" {
			// Only include client secret if no client certificate is provided
			return o.ClientCert == ""
		}
		if fieldName == "ClientCert" || fieldName == "ClientCertPassword" {
			// Include client certificate fields if certificate is provided
			return o.ClientCert != ""
		}
		// SPN requires ClientID and TenantID
		return fieldName == fieldNameClientID || fieldName == fieldNameTenantID ||
			fieldName == fieldNameEnvironment || fieldName == "PoPTokenClaims" || fieldName == "RedirectURL"
	case loginMethodDeviceCode, loginMethodInteractive:
		// These methods require ClientID and TenantID
		return fieldName == fieldNameClientID || fieldName == fieldNameTenantID ||
			fieldName == fieldNameEnvironment || fieldName == "PoPTokenClaims" ||
			fieldName == "RedirectURL" || fieldName == "LoginHint"
	case loginMethodROPC:
		// ROPC requires ClientID and TenantID plus username/password
		return fieldName == fieldNameClientID || fieldName == fieldNameTenantID ||
			fieldName == "Username" || fieldName == "Password" || fieldName == fieldNameEnvironment
	case loginMethodMSI:
		// MSI can optionally include ClientID for specific identity, but only if explicitly set
		if fieldName == fieldNameClientID {
			return o.isSet("client-id")
		}
		if fieldName == "IdentityResourceID" {
			return o.isSet("identity-resource-id")
		}
		return false
	case loginMethodWorkloadIdentity:
		// Workload Identity requires ClientID and TenantID
		return fieldName == fieldNameClientID || fieldName == fieldNameTenantID ||
			fieldName == "FederatedTokenFile" || fieldName == "AuthorityHost"
	case loginMethodAzureCLI, loginMethodAzd:
		// Azure CLI and AZD authentication don't use ClientID/TenantID args
		// This is intentional - see issue #123 about Azure CLI with MSI
		// For cache directory, only include if explicitly set by user (not default value)
		if fieldName == "AuthRecordCacheDir" {
			return o.isSet("cache-dir")
		}
		return false
	}

	// Include cache and other optional fields for all methods
	return fieldName == "AuthRecordCacheDir"
}

// shouldPreserveLegacyFlag determines if legacy flag should be preserved
func (o *UnifiedOptions) shouldPreserveLegacyFlag(authInfo *api.AuthInfo) bool {
	if o.isLegacyAzureAuth(authInfo) && authInfo.AuthProvider.Config != nil {
		if configMode, ok := authInfo.AuthProvider.Config["config-mode"]; ok {
			return configMode == "" || configMode == "0"
		}
		return true // Default to legacy for old azure auth provider
	}
	return false
}

// buildEnvVars creates environment variables for the exec config
func (o *UnifiedOptions) buildEnvVars() []api.ExecEnvVar {
	var envVars []api.ExecEnvVar
	if o.AzureConfigDir != "" && (o.LoginMethod == loginMethodAzureCLI || o.LoginMethod == loginMethodAzd) {
		envVars = append(envVars, api.ExecEnvVar{
			Name:  azureConfigDir,
			Value: o.AzureConfigDir,
		})
	}
	return envVars
}

package converter

import (
	"fmt"
	"strings"

	"github.com/Azure/kubelogin/pkg/internal/token"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	klog "k8s.io/klog/v2"
)

const (
	azureAuthProvider = "azure"
	cfgClientID       = "client-id"
	cfgApiserverID    = "apiserver-id"
	cfgTenantID       = "tenant-id"
	cfgEnvironment    = "environment"
	cfgConfigMode     = "config-mode"

	argClientID                   = "--client-id"
	argServerID                   = "--server-id"
	argTenantID                   = "--tenant-id"
	argEnvironment                = "--environment"
	argClientSecret               = "--client-secret"
	argClientCert                 = "--client-certificate"
	argClientCertPassword         = "--client-certificate-password"
	argIsLegacy                   = "--legacy"
	argUsername                   = "--username"
	argPassword                   = "--password"
	argLoginMethod                = "--login"
	argIdentityResourceID         = "--identity-resource-id"
	argAuthorityHost              = "--authority-host"
	argFederatedTokenFile         = "--federated-token-file"
	argTokenCacheDir              = "--token-cache-dir"
	argAuthRecordCacheDir         = "--cache-dir"
	argIsPoPTokenEnabled          = "--pop-enabled"
	argPoPTokenClaims             = "--pop-claims"
	argDisableEnvironmentOverride = "--disable-environment-override"
	argRedirectURL                = "--redirect-url"
	argLoginHint                  = "--login-hint"

	flagAzureConfigDir             = "azure-config-dir"
	flagClientID                   = "client-id"
	flagContext                    = "context"
	flagServerID                   = "server-id"
	flagTenantID                   = "tenant-id"
	flagEnvironment                = "environment"
	flagClientSecret               = "client-secret"
	flagClientCert                 = "client-certificate"
	flagClientCertPassword         = "client-certificate-password"
	flagIsLegacy                   = "legacy"
	flagUsername                   = "username"
	flagPassword                   = "password"
	flagLoginMethod                = "login"
	flagIdentityResourceID         = "identity-resource-id"
	flagAuthorityHost              = "authority-host"
	flagFederatedTokenFile         = "federated-token-file"
	flagTokenCacheDir              = "token-cache-dir"
	flagAuthRecordCacheDir         = "cache-dir"
	flagIsPoPTokenEnabled          = "pop-enabled"
	flagPoPTokenClaims             = "pop-claims"
	flagDisableEnvironmentOverride = "disable-environment-override"
	flagRedirectURL                = "redirect-url"
	flagLoginHint                  = "login-hint"

	execName        = "kubelogin"
	getTokenCommand = "get-token"
	execAPIVersion  = "client.authentication.k8s.io/v1beta1"
	execInstallHint = `
kubelogin is not installed which is required to connect to AAD enabled cluster.

To learn more, please go to https://azure.github.io/kubelogin/
`

	azureConfigDir = "AZURE_CONFIG_DIR"
)

func getArgValues(o Options, authInfo *api.AuthInfo) (
	argServerIDVal,
	argClientIDVal,
	argEnvironmentVal,
	argTenantIDVal,
	argAuthRecordCacheDirVal,
	argPoPTokenClaimsVal,
	argRedirectURLVal,
	argLoginHintVal string,
	argIsLegacyConfigModeVal,
	argIsPoPTokenEnabledVal bool,
) {
	if authInfo == nil {
		return
	}

	isLegacyAuthProvider := isLegacyAzureAuth(authInfo)

	if o.isSet(flagEnvironment) {
		argEnvironmentVal = o.TokenOptions.Environment
	} else if isLegacyAuthProvider {
		if x, ok := authInfo.AuthProvider.Config[cfgEnvironment]; ok {
			argEnvironmentVal = x
		}
	} else {
		argEnvironmentVal = getExecArg(authInfo, argEnvironment)
	}

	if o.isSet(flagTenantID) {
		argTenantIDVal = o.TokenOptions.TenantID
	} else if isLegacyAuthProvider {
		if x, ok := authInfo.AuthProvider.Config[cfgTenantID]; ok {
			argTenantIDVal = x
		}
	} else {
		argTenantIDVal = getExecArg(authInfo, argTenantID)
	}

	if o.isSet(flagClientID) {
		argClientIDVal = o.TokenOptions.ClientID
	} else if isLegacyAuthProvider {
		if x, ok := authInfo.AuthProvider.Config[cfgClientID]; ok {
			argClientIDVal = x
		}
	} else {
		argClientIDVal = getExecArg(authInfo, argClientID)
	}

	if o.isSet(flagServerID) {
		argServerIDVal = o.TokenOptions.ServerID
	} else if isLegacyAuthProvider {
		if x, ok := authInfo.AuthProvider.Config[cfgApiserverID]; ok {
			argServerIDVal = x
		}
	} else {
		argServerIDVal = getExecArg(authInfo, argServerID)
	}

	if o.isSet(flagIsLegacy) && o.TokenOptions.IsLegacy {
		argIsLegacyConfigModeVal = true
	} else if isLegacyAuthProvider {
		if x := authInfo.AuthProvider.Config[cfgConfigMode]; x == "" || x == "0" {
			argIsLegacyConfigModeVal = true
		}
	} else {
		if found := getExecBoolArg(authInfo, argIsLegacy); found {
			argIsLegacyConfigModeVal = true
		}
	}

	if o.isSet(flagAuthRecordCacheDir) || o.isSet(flagTokenCacheDir) {
		argAuthRecordCacheDirVal = o.TokenOptions.AuthRecordCacheDir
	} else {
		if val := getExecArg(authInfo, argAuthRecordCacheDir); val != "" {
			argAuthRecordCacheDirVal = val
		} else {
			argAuthRecordCacheDirVal = getExecArg(authInfo, argTokenCacheDir)
		}
	}

	if o.isSet(flagIsPoPTokenEnabled) {
		argIsPoPTokenEnabledVal = o.TokenOptions.IsPoPTokenEnabled
	} else {
		if found := getExecBoolArg(authInfo, argIsPoPTokenEnabled); found {
			argIsPoPTokenEnabledVal = true
		}
	}

	if o.isSet(flagPoPTokenClaims) {
		argPoPTokenClaimsVal = o.TokenOptions.PoPTokenClaims
	} else {
		argPoPTokenClaimsVal = getExecArg(authInfo, argPoPTokenClaims)
	}

	if o.isSet(flagRedirectURL) {
		argRedirectURLVal = o.TokenOptions.RedirectURL
	} else {
		argRedirectURLVal = getExecArg(authInfo, argRedirectURL)
	}

	if o.isSet(flagLoginHint) {
		argLoginHintVal = o.TokenOptions.LoginHint
	} else {
		argLoginHintVal = getExecArg(authInfo, argLoginHint)
	}

	return
}

func isLegacyAzureAuth(authInfoPtr *api.AuthInfo) (ok bool) {
	if authInfoPtr == nil {
		return
	}
	if authInfoPtr.AuthProvider == nil {
		return
	}
	return authInfoPtr.AuthProvider.Name == azureAuthProvider
}

func isExecUsingkubelogin(authInfoPtr *api.AuthInfo) (ok bool) {
	if authInfoPtr == nil {
		return
	}
	if authInfoPtr.Exec == nil {
		return
	}
	lowerc := strings.ToLower(authInfoPtr.Exec.Command)
	return strings.Contains(lowerc, "kubelogin")
}

func Convert(o Options, pathOptions *clientcmd.PathOptions) error {
	clientConfig := o.configFlags.ToRawKubeConfigLoader()
	var kubeconfigs []string

	klog.V(5).Info(o.ToString())

	if clientConfig.ConfigAccess() != nil {
		if clientConfig.ConfigAccess().GetExplicitFile() != "" {
			kubeconfigs = append(kubeconfigs, clientConfig.ConfigAccess().GetExplicitFile())
		} else {
			kubeconfigs = append(kubeconfigs, clientConfig.ConfigAccess().GetLoadingPrecedence()...)
		}
	}

	klog.V(5).Infof("Loading kubeconfig from %s", strings.Join(kubeconfigs, ":"))

	config, err := clientConfig.RawConfig()
	if err != nil {
		return fmt.Errorf("unable to load kubeconfig: %s", err)
	}

	targetAuthInfo := ""

	if o.context != "" {
		if config.Contexts[o.context] == nil {
			return fmt.Errorf("no context exists with the name: %q", o.context)
		}
		targetAuthInfo = config.Contexts[o.context].AuthInfo
	}

	for name, authInfo := range config.AuthInfos {

		if targetAuthInfo != "" && name != targetAuthInfo {
			continue
		}

		klog.V(5).Infof("context: %q", name)

		//  is it legacy aad auth or is it exec using kubelogin?
		if !isExecUsingkubelogin(authInfo) && !isLegacyAzureAuth(authInfo) {
			continue
		}

		klog.V(5).Info("converting...")

		argServerIDVal,
			argClientIDVal,
			argEnvironmentVal,
			argTenantIDVal,
			argAuthRecordCacheDirVal,
			argPoPTokenClaimsVal,
			argRedirectURLVal,
			argLoginHintVal,
			isLegacyConfigMode,
			isPoPTokenEnabled := getArgValues(o, authInfo)
		exec := &api.ExecConfig{
			Command: execName,
			Args: []string{
				getTokenCommand,
			},
			APIVersion:  execAPIVersion,
			InstallHint: execInstallHint,
		}

		// Preserve any existing install hint
		if authInfo.Exec != nil && authInfo.Exec.InstallHint != "" {
			exec.InstallHint = authInfo.Exec.InstallHint
		}

		exec.Args = append(exec.Args, argLoginMethod, o.TokenOptions.LoginMethod)

		// all login methods require --server-id specified
		if argServerIDVal == "" {
			return fmt.Errorf("%s is required", argServerID)
		}
		exec.Args = append(exec.Args, argServerID, argServerIDVal)

		if argAuthRecordCacheDirVal != "" {
			exec.Args = append(exec.Args, argAuthRecordCacheDir, argAuthRecordCacheDirVal)
		}

		switch o.TokenOptions.LoginMethod {
		case token.AzureDeveloperCLILogin:
			if o.isSet(flagTenantID) {
				exec.Args = append(exec.Args, argTenantID, o.TokenOptions.TenantID)
			}

		case token.AzureCLILogin:

			if o.azureConfigDir != "" {
				exec.Env = append(exec.Env, api.ExecEnvVar{Name: azureConfigDir, Value: o.azureConfigDir})
			}

			// when convert to azurecli login, tenantID from the input kubeconfig will be disregarded and
			// will have to come from explicit flag `--tenant-id`.
			// this is because azure cli logged in using MSI does not allow specifying tenant ID
			// see https://github.com/Azure/kubelogin/issues/123#issuecomment-1209652342
			if o.isSet(flagTenantID) {
				exec.Args = append(exec.Args, argTenantID, o.TokenOptions.TenantID)
			}

		case token.DeviceCodeLogin:

			if argClientIDVal == "" {
				return fmt.Errorf("%s is required", argClientID)
			}

			exec.Args = append(exec.Args, argClientID, argClientIDVal)

			if argTenantIDVal == "" {
				return fmt.Errorf("%s is required", argTenantID)
			}

			exec.Args = append(exec.Args, argTenantID, argTenantIDVal)

			if argEnvironmentVal != "" {
				// environment is optional
				exec.Args = append(exec.Args, argEnvironment, argEnvironmentVal)
			}

			if isLegacyConfigMode {
				exec.Args = append(exec.Args, argIsLegacy)
			}

		case token.InteractiveLogin:

			if argClientIDVal == "" {
				return fmt.Errorf("%s is required", argClientID)
			}

			exec.Args = append(exec.Args, argClientID, argClientIDVal)

			if argTenantIDVal == "" {
				return fmt.Errorf("%s is required", argTenantID)
			}

			exec.Args = append(exec.Args, argTenantID, argTenantIDVal)

			if argEnvironmentVal != "" {
				// environment is optional
				exec.Args = append(exec.Args, argEnvironment, argEnvironmentVal)
			}

			// PoP token flags are optional but must be provided together
			exec.Args, err = validatePoPClaims(exec.Args, isPoPTokenEnabled, argPoPTokenClaims, argPoPTokenClaimsVal)
			if err != nil {
				return err
			}

			if argRedirectURLVal != "" {
				exec.Args = append(exec.Args, argRedirectURL, argRedirectURLVal)
			}

			if argLoginHintVal != "" {
				exec.Args = append(exec.Args, argLoginHint, argLoginHintVal)
			}

		case token.ServicePrincipalLogin:

			if argClientIDVal == "" {
				return fmt.Errorf("%s is required", argClientID)
			}

			exec.Args = append(exec.Args, argClientID, argClientIDVal)

			if argTenantIDVal == "" {
				return fmt.Errorf("%s is required", argTenantID)
			}

			exec.Args = append(exec.Args, argTenantID, argTenantIDVal)

			if argEnvironmentVal != "" {
				// environment is optional
				exec.Args = append(exec.Args, argEnvironment, argEnvironmentVal)
			}

			if o.isSet(flagClientSecret) {
				exec.Args = append(exec.Args, argClientSecret, o.TokenOptions.ClientSecret)
			}

			if o.isSet(flagClientCert) {
				exec.Args = append(exec.Args, argClientCert, o.TokenOptions.ClientCert)
			}

			if o.isSet(flagClientCertPassword) {
				exec.Args = append(exec.Args, argClientCertPassword, o.TokenOptions.ClientCertPassword)
			}

			if isLegacyConfigMode {
				exec.Args = append(exec.Args, argIsLegacy)
			}

			// PoP token flags are optional but must be provided together
			exec.Args, err = validatePoPClaims(exec.Args, isPoPTokenEnabled, argPoPTokenClaims, argPoPTokenClaimsVal)
			if err != nil {
				return err
			}

			if o.isSet(flagDisableEnvironmentOverride) {
				exec.Args = append(exec.Args, argDisableEnvironmentOverride)
			}

		case token.MSILogin:

			if o.isSet(flagClientID) {
				exec.Args = append(exec.Args, argClientID, o.TokenOptions.ClientID)
			} else if o.isSet(flagIdentityResourceID) {
				exec.Args = append(exec.Args, argIdentityResourceID, o.TokenOptions.IdentityResourceID)
			}

		case token.ROPCLogin:

			if argClientIDVal == "" {
				return fmt.Errorf("%s is required", argClientID)
			}

			exec.Args = append(exec.Args, argClientID, argClientIDVal)

			if argTenantIDVal == "" {
				return fmt.Errorf("%s is required", argTenantID)
			}

			exec.Args = append(exec.Args, argTenantID, argTenantIDVal)

			if argEnvironmentVal != "" {
				// environment is optional
				exec.Args = append(exec.Args, argEnvironment, argEnvironmentVal)
			}

			if o.isSet(flagUsername) {
				exec.Args = append(exec.Args, argUsername, o.TokenOptions.Username)
			}

			if o.isSet(flagPassword) {
				exec.Args = append(exec.Args, argPassword, o.TokenOptions.Password)
			}

			if isLegacyConfigMode {
				exec.Args = append(exec.Args, argIsLegacy)
			}

		case token.WorkloadIdentityLogin:

			if o.isSet(flagClientID) {
				exec.Args = append(exec.Args, argClientID, o.TokenOptions.ClientID)
			}

			if o.isSet(flagTenantID) {
				exec.Args = append(exec.Args, argTenantID, o.TokenOptions.TenantID)
			}

			if o.isSet(flagAuthorityHost) {
				exec.Args = append(exec.Args, argAuthorityHost, o.TokenOptions.AuthorityHost)
			}

			if o.isSet(flagFederatedTokenFile) {
				exec.Args = append(exec.Args, argFederatedTokenFile, o.TokenOptions.FederatedTokenFile)
			}
		}

		authInfo.Exec = exec
		authInfo.AuthProvider = nil
	}
	err = clientcmd.ModifyConfig(pathOptions, config, true)
	return err
}

// get the item in Exec.Args[] right after someArg
func getExecArg(authInfoPtr *api.AuthInfo, someArg string) (resultStr string) {
	if someArg == "" {
		return
	}
	if authInfoPtr == nil || authInfoPtr.Exec == nil || authInfoPtr.Exec.Args == nil {
		return
	}
	if len(authInfoPtr.Exec.Args) < 1 {
		return
	}
	for i := range authInfoPtr.Exec.Args {
		if authInfoPtr.Exec.Args[i] == someArg {
			if len(authInfoPtr.Exec.Args) > i+1 {
				return authInfoPtr.Exec.Args[i+1]
			}
		}
	}
	return
}

func getExecBoolArg(authInfoPtr *api.AuthInfo, someArg string) bool {
	if someArg == "" {
		return false
	}
	if authInfoPtr == nil || authInfoPtr.Exec == nil || authInfoPtr.Exec.Args == nil {
		return false
	}
	if len(authInfoPtr.Exec.Args) < 1 {
		return false
	}
	for i := range authInfoPtr.Exec.Args {
		if authInfoPtr.Exec.Args[i] == someArg {
			return true
		}
	}
	return false
}

// If enabling PoP token support, users must provide both "--pop-enabled" and "--pop-claims" flags together.
// If either is provided without the other, validation should throw an error, otherwise the get-token command
// will fail under the hood.
func validatePoPClaims(args []string, isPopTokenEnabled bool, popTokenClaimsFlag, popTokenClaimsVal string) ([]string, error) {
	if isPopTokenEnabled && popTokenClaimsVal == "" {
		// pop-enabled and pop-claims must be provided together
		return args, fmt.Errorf("%s is required when specifying %s", argPoPTokenClaims, argIsPoPTokenEnabled)
	}

	if popTokenClaimsVal != "" && !isPopTokenEnabled {
		// pop-enabled and pop-claims must be provided together
		return args, fmt.Errorf("%s is required when specifying %s", argIsPoPTokenEnabled, argPoPTokenClaims)
	}

	if isPopTokenEnabled && popTokenClaimsVal != "" {
		args = append(args, argIsPoPTokenEnabled)
		args = append(args, popTokenClaimsFlag, popTokenClaimsVal)
	}

	return args, nil
}

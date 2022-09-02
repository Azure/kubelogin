package converter

import (
	"fmt"
	"strings"

	"github.com/Azure/kubelogin/pkg/token"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	azureAuthProvider = "azure"
	cfgClientID       = "client-id"
	cfgApiserverID    = "apiserver-id"
	cfgTenantID       = "tenant-id"
	cfgEnvironment    = "environment"
	cfgConfigMode     = "config-mode"

	argClientID           = "--client-id"
	argServerID           = "--server-id"
	argTenantID           = "--tenant-id"
	argEnvironment        = "--environment"
	argClientSecret       = "--client-secret"
	argClientCert         = "--client-certificate"
	argClientCertPassword = "--client-certificate-password"
	argIsLegacy           = "--legacy"
	argUsername           = "--username"
	argPassword           = "--password"
	argLoginMethod        = "--login"
	argIdentityResourceID = "--identity-resource-id"
	argAuthorityHost      = "--authority-host"
	argFederatedTokenFile = "--federated-token-file"
	argTokenCacheDir      = "--token-cache-dir"

	flagClientID           = "client-id"
	flagServerID           = "server-id"
	flagTenantID           = "tenant-id"
	flagEnvironment        = "environment"
	flagClientSecret       = "client-secret"
	flagClientCert         = "client-certificate"
	flagClientCertPassword = "client-certificate-password"
	flagIsLegacy           = "legacy"
	flagUsername           = "username"
	flagPassword           = "password"
	flagLoginMethod        = "login"
	flagIdentityResourceID = "identity-resource-id"
	flagAuthorityHost      = "authority-host"
	flagFederatedTokenFile = "federated-token-file"
	flagTokenCacheDir      = "token-cache-dir"

	execName        = "kubelogin"
	getTokenCommand = "get-token"
	execAPIVersion  = "client.authentication.k8s.io/v1beta1"
)

func getArgValues(o Options, authInfo *api.AuthInfo) (argServerIDVal, argClientIDVal, argEnvironmentVal, argTenantIDVal, argTokenCacheDirVal string, argIsLegacyConfigModeVal bool) {
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

	if o.isSet(flagTokenCacheDir) {
		argTokenCacheDirVal = o.TokenOptions.TokenCacheDir
	} else {
		argTokenCacheDirVal = getExecArg(authInfo, argTokenCacheDir)
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
	config, err := o.configFlags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return fmt.Errorf("unable to load kubeconfig: %s", err)
	}

	for _, authInfo := range config.AuthInfos {

		//  is it legacy aad auth or is it exec using kubelogin?
		if !isExecUsingkubelogin(authInfo) && !isLegacyAzureAuth(authInfo) {
			continue
		}
		argServerIDVal, argClientIDVal, argEnvironmentVal, argTenantIDVal, argTokenCacheDirVal, isLegacyConfigMode := getArgValues(o, authInfo)
		exec := &api.ExecConfig{
			Command: execName,
			Args: []string{
				getTokenCommand,
			},
			APIVersion: execAPIVersion,
		}

		exec.Args = append(exec.Args, argLoginMethod)
		exec.Args = append(exec.Args, o.TokenOptions.LoginMethod)

		// all login methods require --server-id specified
		if argServerIDVal == "" {
			return fmt.Errorf("%s is required", argServerID)
		}
		exec.Args = append(exec.Args, argServerID)
		exec.Args = append(exec.Args, argServerIDVal)

		if argTokenCacheDirVal != "" {
			exec.Args = append(exec.Args, argTokenCacheDir)
			exec.Args = append(exec.Args, argTokenCacheDirVal)
		}

		switch o.TokenOptions.LoginMethod {
		case token.AzureCLILogin:

			// when convert to azurecli login, tenantID from the input kubeconfig will be disregarded and
			// will have to come from explicit flag `--tenant-id`.
			// this is because azure cli logged in using MSI does not allow specifying tenant ID
			// see https://github.com/Azure/kubelogin/issues/123#issuecomment-1209652342
			if o.isSet(flagTenantID) {
				exec.Args = append(exec.Args, argTenantID)
				exec.Args = append(exec.Args, o.TokenOptions.TenantID)
			}

		case token.DeviceCodeLogin:

			if argClientIDVal == "" {
				return fmt.Errorf("%s is required", argClientID)
			}

			exec.Args = append(exec.Args, argClientID)
			exec.Args = append(exec.Args, argClientIDVal)

			if argTenantIDVal == "" {
				return fmt.Errorf("%s is required", argTenantID)
			}

			exec.Args = append(exec.Args, argTenantID)
			exec.Args = append(exec.Args, argTenantIDVal)

			if argEnvironmentVal != "" {
				// environment is optional
				exec.Args = append(exec.Args, argEnvironment)
				exec.Args = append(exec.Args, argEnvironmentVal)
			}

			if isLegacyConfigMode {
				exec.Args = append(exec.Args, argIsLegacy)
			}

		case token.ServicePrincipalLogin:

			if argClientIDVal == "" {
				return fmt.Errorf("%s is required", argClientID)
			}

			exec.Args = append(exec.Args, argClientID)
			exec.Args = append(exec.Args, argClientIDVal)

			if argTenantIDVal == "" {
				return fmt.Errorf("%s is required", argTenantID)
			}

			exec.Args = append(exec.Args, argTenantID)
			exec.Args = append(exec.Args, argTenantIDVal)

			if argEnvironmentVal != "" {
				// environment is optional
				exec.Args = append(exec.Args, argEnvironment)
				exec.Args = append(exec.Args, argEnvironmentVal)
			}

			if o.isSet(flagClientSecret) {
				exec.Args = append(exec.Args, argClientSecret)
				exec.Args = append(exec.Args, o.TokenOptions.ClientSecret)
			}

			if o.isSet(flagClientCert) {
				exec.Args = append(exec.Args, argClientCert)
				exec.Args = append(exec.Args, o.TokenOptions.ClientCert)
			}

			if o.isSet(flagClientCertPassword) {
				exec.Args = append(exec.Args, argClientCertPassword)
				exec.Args = append(exec.Args, o.TokenOptions.ClientCertPassword)
			}

			if isLegacyConfigMode {
				exec.Args = append(exec.Args, argIsLegacy)
			}

		case token.MSILogin:

			if o.isSet(flagClientID) {
				exec.Args = append(exec.Args, argClientID)
				exec.Args = append(exec.Args, o.TokenOptions.ClientID)
			} else if o.isSet(flagIdentityResourceID) {
				exec.Args = append(exec.Args, argIdentityResourceID)
				exec.Args = append(exec.Args, o.TokenOptions.IdentityResourceId)
			}

		case token.ROPCLogin:

			if argClientIDVal == "" {
				return fmt.Errorf("%s is required", argClientID)
			}

			exec.Args = append(exec.Args, argClientID)
			exec.Args = append(exec.Args, argClientIDVal)

			if argTenantIDVal == "" {
				return fmt.Errorf("%s is required", argTenantID)
			}

			exec.Args = append(exec.Args, argTenantID)
			exec.Args = append(exec.Args, argTenantIDVal)

			if argEnvironmentVal != "" {
				// environment is optional
				exec.Args = append(exec.Args, argEnvironment)
				exec.Args = append(exec.Args, argEnvironmentVal)
			}

			if o.isSet(flagUsername) {
				exec.Args = append(exec.Args, argUsername)
				exec.Args = append(exec.Args, o.TokenOptions.Username)
			}

			if o.isSet(flagPassword) {
				exec.Args = append(exec.Args, argPassword)
				exec.Args = append(exec.Args, o.TokenOptions.Password)
			}

			if isLegacyConfigMode {
				exec.Args = append(exec.Args, argIsLegacy)
			}

		case token.WorkloadIdentityLogin:

			if o.isSet(flagClientID) {
				exec.Args = append(exec.Args, argClientID)
				exec.Args = append(exec.Args, o.TokenOptions.ClientID)
			}

			if o.isSet(flagTenantID) {
				exec.Args = append(exec.Args, argTenantID)
				exec.Args = append(exec.Args, o.TokenOptions.TenantID)
			}

			if o.isSet(flagAuthorityHost) {
				exec.Args = append(exec.Args, argAuthorityHost)
				exec.Args = append(exec.Args, o.TokenOptions.AuthorityHost)
			}

			if o.isSet(flagFederatedTokenFile) {
				exec.Args = append(exec.Args, argFederatedTokenFile)
				exec.Args = append(exec.Args, o.TokenOptions.FederatedTokenFile)
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

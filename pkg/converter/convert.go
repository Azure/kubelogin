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
	argIsLegacy           = "--legacy"
	argUsername           = "--username"
	argPassword           = "--password"
	argLoginMethod        = "--login"
	argIdentityResourceID = "--identity-resource-id"
	argAuthorityHost      = "--authority-host"

	flagClientID           = "client-id"
	flagServerID           = "server-id"
	flagTenantID           = "tenant-id"
	flagEnvironment        = "environment"
	flagClientSecret       = "client-secret"
	flagClientCert         = "client-certificate"
	flagIsLegacy           = "legacy"
	flagUsername           = "username"
	flagPassword           = "password"
	flagLoginMethod        = "login"
	flagIdentityResourceID = "identity-resource-id"
	flagAuthorityHost      = "authority-host"

	execName        = "kubelogin"
	getTokenCommand = "get-token"
	execAPIVersion  = "client.authentication.k8s.io/v1beta1"
)

func getArgValues(o Options, authInfo *api.AuthInfo) (argServerIDVal, argClientIDVal, argEnvironmentVal, argTenantIDVal string, argIsLegacyConfigModeVal bool) {
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

func Convert(o Options) error {
	config, err := o.configFlags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return fmt.Errorf("unable to load kubeconfig: %s", err)
	}

	for _, authInfo := range config.AuthInfos {

		//  is it legacy aad auth or is it exec using kubelogin?
		if !isExecUsingkubelogin(authInfo) && !isLegacyAzureAuth(authInfo) {
			continue
		}
		argServerIDVal, argClientIDVal, argEnvironmentVal, argTenantIDVal, isLegacyConfigMode := getArgValues(o, authInfo)
		exec := &api.ExecConfig{
			Command: execName,
			Args: []string{
				getTokenCommand,
			},
			APIVersion: execAPIVersion,
		}

		switch o.TokenOptions.LoginMethod {
		case token.AzureCLILogin:
			if argServerIDVal == "" {
				return fmt.Errorf("%s is required", argServerID)
			}
			exec.Args = append(exec.Args, argServerID)
			exec.Args = append(exec.Args, argServerIDVal)

			exec.Args = append(exec.Args, argLoginMethod)
			exec.Args = append(exec.Args, o.TokenOptions.LoginMethod)

		case token.DeviceCodeLogin:
			if argServerIDVal == "" {
				return fmt.Errorf("%s is required", argServerID)
			}

			exec.Args = append(exec.Args, argServerID)
			exec.Args = append(exec.Args, argServerIDVal)

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

			exec.Args = append(exec.Args, argLoginMethod)
			exec.Args = append(exec.Args, o.TokenOptions.LoginMethod)

			if isLegacyConfigMode {
				exec.Args = append(exec.Args, argIsLegacy)
			}

		case token.ServicePrincipalLogin:
			if argServerIDVal == "" {
				return fmt.Errorf("%s is required", argServerID)
			}

			exec.Args = append(exec.Args, argServerID)
			exec.Args = append(exec.Args, argServerIDVal)

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

			exec.Args = append(exec.Args, argEnvironment)
			exec.Args = append(exec.Args, argEnvironmentVal)
			exec.Args = append(exec.Args, argLoginMethod)
			exec.Args = append(exec.Args, o.TokenOptions.LoginMethod)

			if o.isSet(flagClientSecret) {
				exec.Args = append(exec.Args, argClientSecret)
				exec.Args = append(exec.Args, o.TokenOptions.ClientSecret)
			}

			if o.isSet(flagClientCert) {
				exec.Args = append(exec.Args, argClientCert)
				exec.Args = append(exec.Args, o.TokenOptions.ClientCert)
			}

			if isLegacyConfigMode {
				exec.Args = append(exec.Args, argIsLegacy)
			}

		case token.MSILogin:
			if argServerIDVal == "" {
				return fmt.Errorf("%s is required", argServerID)
			}

			exec.Args = append(exec.Args, argServerID)
			exec.Args = append(exec.Args, argServerIDVal)

			exec.Args = append(exec.Args, argLoginMethod)
			exec.Args = append(exec.Args, o.TokenOptions.LoginMethod)

			if o.isSet(flagClientID) {
				exec.Args = append(exec.Args, argClientID)
				exec.Args = append(exec.Args, o.TokenOptions.ClientID)
			} else if o.isSet(flagIdentityResourceID) {
				exec.Args = append(exec.Args, argIdentityResourceID)
				exec.Args = append(exec.Args, o.TokenOptions.IdentityResourceId)
			}

		case token.ROPCLogin:
			if argServerIDVal == "" {
				return fmt.Errorf("%s is required", argServerID)
			}

			exec.Args = append(exec.Args, argServerID)
			exec.Args = append(exec.Args, argServerIDVal)

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

			exec.Args = append(exec.Args, argLoginMethod)
			exec.Args = append(exec.Args, o.TokenOptions.LoginMethod)

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
			if argServerIDVal == "" {
				return fmt.Errorf("%s is required", argServerID)
			}

			exec.Args = append(exec.Args, argServerID)
			exec.Args = append(exec.Args, argServerIDVal)

			exec.Args = append(exec.Args, argLoginMethod)
			exec.Args = append(exec.Args, o.TokenOptions.LoginMethod)

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

		default:
			return fmt.Errorf("unsupported login mehod: %s", o.TokenOptions.LoginMethod)
		}
		authInfo.Exec = exec
		authInfo.AuthProvider = nil
	}
	err = clientcmd.ModifyConfig(clientcmd.NewDefaultPathOptions(), config, true)
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

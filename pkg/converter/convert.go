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

	argClientID     = "--client-id"
	argServerID     = "--server-id"
	argTenantID     = "--tenant-id"
	argEnvironment  = "--environment"
	argClientSecret = "--client-secret"
	argClientCert   = "--client-certificate"
	argIsLegacy     = "--legacy"
	argUsername     = "--username"
	argPassword     = "--password"
	argLoginMethod  = "--login"

	flagClientID     = "client-id"
	flagServerID     = "server-id"
	flagTenantID     = "tenant-id"
	flagEnvironment  = "environment"
	flagClientSecret = "client-secret"
	flagClientCert   = "client-certificate"
	flagIsLegacy     = "legacy"
	flagUsername     = "username"
	flagPassword     = "password"
	flagLoginMethod  = "login"

	execName        = "kubelogin"
	getTokenCommand = "get-token"
	execAPIVersion  = "client.authentication.k8s.io/v1beta1"
)

func getArgValues(o Options, authInfo *api.AuthInfo) (argServerIDVal, argClientIDVal, argEnvironmentVal, argTenantIDVal, cfgConfigModeVal string) {
	if authInfo == nil {
		return
	}
	authProviderBool := authInfo.AuthProvider != nil

	if o.isSet(flagEnvironment) {
		argEnvironmentVal = o.TokenOptions.Environment
	} else if authProviderBool {
		x, ok := authInfo.AuthProvider.Config[cfgEnvironment]
		if ok {
			argEnvironmentVal = x
		}
	} else {
		argEnvironmentVal = getExecArg(authInfo, argEnvironment)
	}

	if o.isSet(flagTenantID) {
		argTenantIDVal = o.TokenOptions.TenantID
	} else if authProviderBool {
		x, ok := authInfo.AuthProvider.Config[cfgTenantID]
		if ok {
			argTenantIDVal = x
		}
	} else {
		argTenantIDVal = getExecArg(authInfo, argTenantID)
	}

	if o.isSet(flagClientID) {
		argClientIDVal = o.TokenOptions.ClientID
	} else if authProviderBool {
		x, ok := authInfo.AuthProvider.Config[cfgClientID]
		if ok {
			argClientIDVal = x
		}
	} else {
		argClientIDVal = getExecArg(authInfo, argClientID)
	}

	if o.isSet(flagServerID) {
		argServerIDVal = o.TokenOptions.ServerID
	} else if authProviderBool {
		// .. is special, we look for cfgApiserverID
		x, ok := authInfo.AuthProvider.Config[cfgApiserverID]
		if ok {
			argServerIDVal = x
		}
	} else {
		argServerIDVal = getExecArg(authInfo, argServerID)
	}

	// cfgConfigMode available only in authInfo.AuthProvider.Config,
	// although the same precedence would work here as well.
	if authProviderBool {
		x, ok := authInfo.AuthProvider.Config[cfgConfigMode]
		if ok {
			cfgConfigModeVal = x
		}
	}

	return
}

func isLegacyAADAuth(authInfoPtr *api.AuthInfo) (ok bool) {
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

	// MSI, AzureCLI, and WorkloadIdentity login bypass most login fields, so we'll check for them and exclude them
	isMSI := o.TokenOptions.LoginMethod == token.MSILogin
	isAzureCLI := o.TokenOptions.LoginMethod == token.AzureCLILogin
	isWorkloadIdentity := o.TokenOptions.LoginMethod == token.WorkloadIdentityLogin
	isAlternativeLogin := isMSI || isAzureCLI || isWorkloadIdentity
	for _, authInfo := range config.AuthInfos {

		//  is it legacy aad auth or is it exec using kubelogin?
		if !isExecUsingkubelogin(authInfo) && !isLegacyAADAuth(authInfo) {
			continue
		}
		argServerIDVal, argClientIDVal, argEnvironmentVal, argTenantIDVal, cfgConfigModeVal := getArgValues(o, authInfo)
		exec := &api.ExecConfig{
			Command: execName,
			Args: []string{
				getTokenCommand,
			},
			APIVersion: execAPIVersion,
		}

		if isExecUsingkubelogin(authInfo) {
			if argServerIDVal == "" {
				return fmt.Errorf("Err: Invalid arg %v", argServerID)
			}
			switch o.TokenOptions.LoginMethod {
			case token.AzureCLILogin:
				exec.Args = append(exec.Args, argServerID)
				exec.Args = append(exec.Args, argServerIDVal)
				exec.Args = append(exec.Args, argLoginMethod)
				exec.Args = append(exec.Args, o.TokenOptions.LoginMethod)
			case token.DeviceCodeLogin:
				exec.Args = append(exec.Args, argServerID)
				exec.Args = append(exec.Args, argServerIDVal)
				exec.Args = append(exec.Args, argClientID)
				exec.Args = append(exec.Args, argClientIDVal)
				exec.Args = append(exec.Args, argTenantID)
				exec.Args = append(exec.Args, argTenantIDVal)
				exec.Args = append(exec.Args, argLoginMethod)
				exec.Args = append(exec.Args, o.TokenOptions.LoginMethod)
			default:
				return fmt.Errorf("%q is not supported yet", o.TokenOptions.LoginMethod)
			}
		} else {
			if !isAlternativeLogin && o.isSet(flagEnvironment) {
				exec.Args = append(exec.Args, argEnvironment)
				exec.Args = append(exec.Args, o.TokenOptions.Environment)
			} else if !isAlternativeLogin && argEnvironmentVal != "" {
				exec.Args = append(exec.Args, argEnvironment)
				exec.Args = append(exec.Args, argEnvironmentVal)
			}
			if o.isSet(flagServerID) {
				exec.Args = append(exec.Args, argServerID)
				exec.Args = append(exec.Args, o.TokenOptions.ServerID)
			} else if argServerIDVal != "" {
				exec.Args = append(exec.Args, argServerID)
				exec.Args = append(exec.Args, argServerIDVal)
			}
			if o.isSet(flagClientID) {
				exec.Args = append(exec.Args, argClientID)
				exec.Args = append(exec.Args, o.TokenOptions.ClientID)
			} else if !isAlternativeLogin && argClientIDVal != "" {
				// when MSI is enabled, the clientID in azure authInfo will be disregarded
				exec.Args = append(exec.Args, argClientID)
				exec.Args = append(exec.Args, argClientIDVal)
			}
			if !isAlternativeLogin && o.isSet(flagTenantID) {
				exec.Args = append(exec.Args, argTenantID)
				exec.Args = append(exec.Args, o.TokenOptions.TenantID)
			} else if !isAlternativeLogin && argTenantIDVal != "" {
				exec.Args = append(exec.Args, argTenantID)
				exec.Args = append(exec.Args, argTenantIDVal)
			}
			if !isAlternativeLogin && o.isSet(flagIsLegacy) && o.TokenOptions.IsLegacy {
				exec.Args = append(exec.Args, argIsLegacy)
			} else if !isAlternativeLogin && (cfgConfigModeVal == "" || cfgConfigModeVal == "0") {
				exec.Args = append(exec.Args, argIsLegacy)
			}
			if !isAlternativeLogin && o.isSet(flagClientSecret) {
				exec.Args = append(exec.Args, argClientSecret)
				exec.Args = append(exec.Args, o.TokenOptions.ClientSecret)
			}
			if !isAlternativeLogin && o.isSet(flagClientCert) {
				exec.Args = append(exec.Args, argClientCert)
				exec.Args = append(exec.Args, o.TokenOptions.ClientCert)
			}
			if !isAlternativeLogin && o.isSet(flagUsername) {
				exec.Args = append(exec.Args, argUsername)
				exec.Args = append(exec.Args, o.TokenOptions.Username)
			}
			if !isAlternativeLogin && o.isSet(flagPassword) {
				exec.Args = append(exec.Args, argPassword)
				exec.Args = append(exec.Args, o.TokenOptions.Password)
			}
			if o.isSet(flagLoginMethod) {
				exec.Args = append(exec.Args, argLoginMethod)
				exec.Args = append(exec.Args, o.TokenOptions.LoginMethod)
			}
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

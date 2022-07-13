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

type goody struct {
	// 1. cfgApiserverID - argServerIDVal
	// 2. cfgClientID - argClientIDVal
	// 3. cfgEnvironment - argEnvironmentVal
	// 4. cfgTenantID - argTenantIDVal
	// 5. cfgConfigMode - cfgConfigModeVal

	// exec formatted kubeconfig doesn't use --config-mode (1) nor apiserver-id (5)

	argEnvironmentVal, argClientIDVal, argServerIDVal, argTenantIDVal, cfgConfigModeVal string
}

func getGoody(o Options, authInfo *api.AuthInfo) (goodyPtr *goody) {
	if authInfo == nil {
		return
	}
	authProviderBool := authInfo.AuthProvider != nil
	goodyPtr = &goody{}

	/* precedence and modus operandi:
	Looking for Bleh:
	1. o.TokenOptions.Bleh
	2. authInfo.AuthProvider.Config[Bleh]
	3. getExecArg(authInfo, Bleh)
	*/

	if o.isSet(flagEnvironment) {
		goodyPtr.argEnvironmentVal = o.TokenOptions.Environment
	} else if authProviderBool {
		x, ok := authInfo.AuthProvider.Config[cfgEnvironment]
		if ok {
			goodyPtr.argEnvironmentVal = x
		}
	} else {
		goodyPtr.argEnvironmentVal = getExecArg(authInfo, argEnvironment)
	}

	if o.isSet(flagTenantID) {
		goodyPtr.argTenantIDVal = o.TokenOptions.TenantID
	} else if authProviderBool {
		x, ok := authInfo.AuthProvider.Config[cfgTenantID]
		if ok {
			goodyPtr.argTenantIDVal = x
		}
	} else {
		goodyPtr.argTenantIDVal = getExecArg(authInfo, argTenantID)
	}

	if o.isSet(flagClientID) {
		goodyPtr.argClientIDVal = o.TokenOptions.ClientID
	} else if authProviderBool {
		x, ok := authInfo.AuthProvider.Config[cfgClientID]
		if ok {
			goodyPtr.argClientIDVal = x
		}
	} else {
		goodyPtr.argClientIDVal = getExecArg(authInfo, argClientID)
	}

	if o.isSet(flagServerID) {
		goodyPtr.argServerIDVal = o.TokenOptions.ServerID
	} else if authProviderBool {
		// .. is special, we look for cfgApiserverID
		x, ok := authInfo.AuthProvider.Config[cfgApiserverID]
		if ok {
			goodyPtr.argServerIDVal = x
		}
	} else {
		goodyPtr.argServerIDVal = getExecArg(authInfo, argServerID)
	}

	// cfgConfigMode available only in authInfo.AuthProvider.Config,
	// although the same precedence would work here as well.
	if authProviderBool {
		x, ok := authInfo.AuthProvider.Config[cfgConfigMode]
		if ok {
			goodyPtr.cfgConfigModeVal = x
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
		goody := getGoody(o, authInfo)

		exec := &api.ExecConfig{
			Command: execName,
			Args: []string{
				getTokenCommand,
			},
			APIVersion: execAPIVersion,
		}

		if isExecUsingkubelogin(authInfo) {
			if goody.argServerIDVal == "" {
				return fmt.Errorf("Err: Invalid arg %v", argServerID)
			}
			switch o.TokenOptions.LoginMethod {
			case token.AzureCLILogin:
				exec.Args = append(exec.Args, argServerID)
				exec.Args = append(exec.Args, goody.argServerIDVal)
				exec.Args = append(exec.Args, argLoginMethod)
				exec.Args = append(exec.Args, o.TokenOptions.LoginMethod)
			case token.DeviceCodeLogin:
				exec.Args = append(exec.Args, argServerID)
				exec.Args = append(exec.Args, goody.argServerIDVal)
				exec.Args = append(exec.Args, argClientID)
				exec.Args = append(exec.Args, goody.argClientIDVal)
				exec.Args = append(exec.Args, argTenantID)
				exec.Args = append(exec.Args, goody.argTenantIDVal)
				exec.Args = append(exec.Args, argLoginMethod)
				exec.Args = append(exec.Args, o.TokenOptions.LoginMethod)
			}
		} else {
			if !isAlternativeLogin && o.isSet(flagEnvironment) {
				exec.Args = append(exec.Args, argEnvironment)
				exec.Args = append(exec.Args, o.TokenOptions.Environment)
			} else if !isAlternativeLogin && goody.argEnvironmentVal != "" {
				exec.Args = append(exec.Args, argEnvironment)
				exec.Args = append(exec.Args, goody.argEnvironmentVal)
			}
			if o.isSet(flagServerID) {
				exec.Args = append(exec.Args, argServerID)
				exec.Args = append(exec.Args, o.TokenOptions.ServerID)
			} else if goody.argServerIDVal != "" {
				exec.Args = append(exec.Args, argServerID)
				exec.Args = append(exec.Args, goody.argServerIDVal)
			}
			if o.isSet(flagClientID) {
				exec.Args = append(exec.Args, argClientID)
				exec.Args = append(exec.Args, o.TokenOptions.ClientID)
			} else if !isAlternativeLogin && goody.argClientIDVal != "" {
				// when MSI is enabled, the clientID in azure authInfo will be disregarded
				exec.Args = append(exec.Args, argClientID)
				exec.Args = append(exec.Args, goody.argClientIDVal)
			}
			if !isAlternativeLogin && o.isSet(flagTenantID) {
				exec.Args = append(exec.Args, argTenantID)
				exec.Args = append(exec.Args, o.TokenOptions.TenantID)
			} else if !isAlternativeLogin && goody.argTenantIDVal != "" {
				exec.Args = append(exec.Args, argTenantID)
				exec.Args = append(exec.Args, goody.argTenantIDVal)
			}
			if !isAlternativeLogin && o.isSet(flagIsLegacy) && o.TokenOptions.IsLegacy {
				exec.Args = append(exec.Args, argIsLegacy)
			} else if !isAlternativeLogin && (goody.cfgConfigModeVal == "" || goody.cfgConfigModeVal == "0") {
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

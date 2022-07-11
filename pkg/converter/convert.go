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

func getServerId(authInfoPtr *api.AuthInfo) (serverId string) {
	if authInfoPtr == nil || authInfoPtr.Exec == nil || authInfoPtr.Exec.Args == nil {
		return
	}
	if len(authInfoPtr.Exec.Args) < 1 {
		return
	}
	for i := range authInfoPtr.Exec.Args {
		if authInfoPtr.Exec.Args[i] == argServerID {
			if len(authInfoPtr.Exec.Args) > i+1 {
				return authInfoPtr.Exec.Args[i+1]
			}
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

		exec := &api.ExecConfig{
			Command: execName,
			Args: []string{
				getTokenCommand,
			},
			APIVersion: execAPIVersion,
		}

		if !o.TokenOptions.IsLegacy && isExecUsingkubelogin(authInfo) {

			switch o.TokenOptions.LoginMethod {
			case token.AzureCLILogin, token.ServicePrincipalLogin, token.DeviceCodeLogin, token.WorkloadIdentityLogin, token.ROPCLogin, token.MSILogin: //azurecli, spn, devicecode, workloadidentity, ropc, msi
				exec.Args = append(exec.Args, argServerID)
				serveridArg := getServerId(authInfo)
				if serveridArg == "" {
					return fmt.Errorf("Err: Invalid serveridArg")
				}
				exec.Args = append(exec.Args, serveridArg)
				exec.Args = append(exec.Args, argLoginMethod)
				exec.Args = append(exec.Args, o.TokenOptions.LoginMethod)
			default:
				return fmt.Errorf("%q is not supported yet", o.TokenOptions.LoginMethod)
			}

		} else {

			if !isAlternativeLogin && o.isSet(flagEnvironment) {
				exec.Args = append(exec.Args, argEnvironment)
				exec.Args = append(exec.Args, o.TokenOptions.Environment)
			} else if !isAlternativeLogin && authInfo.AuthProvider.Config[cfgEnvironment] != "" {
				exec.Args = append(exec.Args, argEnvironment)
				exec.Args = append(exec.Args, authInfo.AuthProvider.Config[cfgEnvironment])
			}
			if o.isSet(flagServerID) {
				exec.Args = append(exec.Args, argServerID)
				exec.Args = append(exec.Args, o.TokenOptions.ServerID)
			} else if authInfo.AuthProvider.Config[cfgApiserverID] != "" {
				exec.Args = append(exec.Args, argServerID)
				exec.Args = append(exec.Args, authInfo.AuthProvider.Config[cfgApiserverID])
			}
			if o.isSet(flagClientID) {
				exec.Args = append(exec.Args, argClientID)
				exec.Args = append(exec.Args, o.TokenOptions.ClientID)
			} else if !isAlternativeLogin && authInfo.AuthProvider.Config[cfgClientID] != "" {
				// when MSI is enabled, the clientID in azure authInfo will be disregarded
				exec.Args = append(exec.Args, argClientID)
				exec.Args = append(exec.Args, authInfo.AuthProvider.Config[cfgClientID])
			}
			if !isAlternativeLogin && o.isSet(flagTenantID) {
				exec.Args = append(exec.Args, argTenantID)
				exec.Args = append(exec.Args, o.TokenOptions.TenantID)
			} else if !isAlternativeLogin && authInfo.AuthProvider.Config[cfgTenantID] != "" {
				exec.Args = append(exec.Args, argTenantID)
				exec.Args = append(exec.Args, authInfo.AuthProvider.Config[cfgTenantID])
			}
			if !isAlternativeLogin && o.isSet(flagIsLegacy) && o.TokenOptions.IsLegacy {
				exec.Args = append(exec.Args, argIsLegacy)
			} else if !isAlternativeLogin && (authInfo.AuthProvider.Config[cfgConfigMode] == "" || authInfo.AuthProvider.Config[cfgConfigMode] == "0") {
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

package converter

import (
	"fmt"

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
	argClientCert   = "--client-cert"
	argIsLegacy     = "--legacy"
	argUsername     = "--username"
	argPassword     = "--password"
	argLoginMethod  = "--login"

	flagClientID     = "client-id"
	flagServerID     = "server-id"
	flagTenantID     = "tenant-id"
	flagEnvironment  = "environment"
	flagClientSecret = "client-secret"
	flagClientCert   = "client-cert"
	flagIsLegacy     = "legacy"
	flagUsername     = "username"
	flagPassword     = "password"
	flagLoginMethod  = "login"

	execName        = "kubelogin"
	getTokenCommand = "get-token"
	execAPIVersion  = "client.authentication.k8s.io/v1beta1"
)

func Convert(o Options) error {
	config, err := o.configFlags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return fmt.Errorf("unable to load kubeconfig: %s", err)
	}

	isMSI := o.TokenOptions.LoginMethod == token.MSILogin
	for _, authInfo := range config.AuthInfos {
		if authInfo != nil {
			if authInfo.AuthProvider == nil || authInfo.AuthProvider.Name != azureAuthProvider {
				continue
			}
			exec := &api.ExecConfig{
				Command: execName,
				Args: []string{
					getTokenCommand,
				},
				APIVersion: execAPIVersion,
			}
			if !isMSI && o.isSet(flagEnvironment) {
				exec.Args = append(exec.Args, argEnvironment)
				exec.Args = append(exec.Args, o.TokenOptions.Environment)
			} else if !isMSI && authInfo.AuthProvider.Config[cfgEnvironment] != "" {
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
			} else if !isMSI && authInfo.AuthProvider.Config[cfgClientID] != "" {
				// when MSI is enabled, the clientID in azure authInfo will be disregarded
				exec.Args = append(exec.Args, argClientID)
				exec.Args = append(exec.Args, authInfo.AuthProvider.Config[cfgClientID])
			}
			if !isMSI && o.isSet(flagTenantID) {
				exec.Args = append(exec.Args, argTenantID)
				exec.Args = append(exec.Args, o.TokenOptions.TenantID)
			} else if !isMSI && authInfo.AuthProvider.Config[cfgTenantID] != "" {
				exec.Args = append(exec.Args, argTenantID)
				exec.Args = append(exec.Args, authInfo.AuthProvider.Config[cfgTenantID])
			}
			if !isMSI && o.isSet(flagIsLegacy) && o.TokenOptions.IsLegacy {
				exec.Args = append(exec.Args, argIsLegacy)
			} else if !isMSI && (authInfo.AuthProvider.Config[cfgConfigMode] == "" || authInfo.AuthProvider.Config[cfgConfigMode] == "0") {
				exec.Args = append(exec.Args, argIsLegacy)
			}
			if !isMSI && o.isSet(flagClientSecret) {
				exec.Args = append(exec.Args, argClientSecret)
				exec.Args = append(exec.Args, o.TokenOptions.ClientSecret)
			}
			if !isMSI && o.isSet(flagClientCert) {
				exec.Args = append(exec.Args, argClientCert)
				exec.Args = append(exec.Args, o.TokenOptions.ClientCert)
			}
			if !isMSI && o.isSet(flagUsername) {
				exec.Args = append(exec.Args, argUsername)
				exec.Args = append(exec.Args, o.TokenOptions.Username)
			}
			if !isMSI && o.isSet(flagPassword) {
				exec.Args = append(exec.Args, argPassword)
				exec.Args = append(exec.Args, o.TokenOptions.Password)
			}
			if o.isSet(flagLoginMethod) {
				exec.Args = append(exec.Args, argLoginMethod)
				exec.Args = append(exec.Args, o.TokenOptions.LoginMethod)
			}
			authInfo.Exec = exec
			authInfo.AuthProvider = nil
		}
	}

	clientcmd.ModifyConfig(clientcmd.NewDefaultPathOptions(), config, true)

	return nil
}

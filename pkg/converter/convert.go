package converter

import (
	"fmt"

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
	argIsLegacy     = "--legacy"
	argUsername     = "--username"
	argPassword     = "--password"
	argLoginMethod  = "--login"
)

func Convert(o Options) error {
	config, err := o.configFlags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return fmt.Errorf("unable to load kubeconfig: %s", err)
	}

	for _, authInfo := range config.AuthInfos {
		if authInfo != nil {
			if authInfo.AuthProvider == nil || authInfo.AuthProvider.Name != azureAuthProvider {
				continue
			}
			exec := &api.ExecConfig{
				Command: "kubelogin",
				Args: []string{
					"get-token",
				},
				APIVersion: "client.authentication.k8s.io/v1beta1",
			}
			if o.isSet("environment") {
				exec.Args = append(exec.Args, argEnvironment)
				exec.Args = append(exec.Args, o.TokenOptions.Environment)
			} else if authInfo.AuthProvider.Config[cfgEnvironment] != "" {
				exec.Args = append(exec.Args, argEnvironment)
				exec.Args = append(exec.Args, authInfo.AuthProvider.Config[cfgEnvironment])
			}
			if o.isSet("server-id") {
				exec.Args = append(exec.Args, argServerID)
				exec.Args = append(exec.Args, o.TokenOptions.ServerID)
			} else if authInfo.AuthProvider.Config[cfgApiserverID] != "" {
				exec.Args = append(exec.Args, argServerID)
				exec.Args = append(exec.Args, authInfo.AuthProvider.Config[cfgApiserverID])
			}
			if o.isSet("client-id") {
				exec.Args = append(exec.Args, argClientID)
				exec.Args = append(exec.Args, o.TokenOptions.ClientID)
			} else if authInfo.AuthProvider.Config[cfgClientID] != "" {
				exec.Args = append(exec.Args, argClientID)
				exec.Args = append(exec.Args, authInfo.AuthProvider.Config[cfgClientID])
			}
			if o.isSet("tenant-id") {
				exec.Args = append(exec.Args, argTenantID)
				exec.Args = append(exec.Args, o.TokenOptions.TenantID)
			} else if authInfo.AuthProvider.Config[cfgClientID] != "" {
				exec.Args = append(exec.Args, argTenantID)
				exec.Args = append(exec.Args, authInfo.AuthProvider.Config[cfgTenantID])
			}
			if o.isSet("legacy") && o.TokenOptions.IsLegacy {
				exec.Args = append(exec.Args, argIsLegacy)
			} else if authInfo.AuthProvider.Config[cfgConfigMode] == "" {
				exec.Args = append(exec.Args, argIsLegacy)
			}
			if o.isSet("client-secret") {
				exec.Args = append(exec.Args, argClientSecret)
				exec.Args = append(exec.Args, o.TokenOptions.ClientSecret)
			}
			if o.isSet("username") {
				exec.Args = append(exec.Args, argUsername)
				exec.Args = append(exec.Args, o.TokenOptions.Username)
			}
			if o.isSet("password") {
				exec.Args = append(exec.Args, argPassword)
				exec.Args = append(exec.Args, o.TokenOptions.Password)
			}
			if o.isSet("login") {
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

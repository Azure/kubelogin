package converter

import (
	"testing"

	"github.com/Azure/kubelogin/pkg/token"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestConvert(t *testing.T) {
	const (
		clusterName  = "aks"
		envName      = "foo"
		serverID     = "serverID"
		clientID     = "clientID"
		tenantID     = "tenantID"
		clientSecret = "foosecret"
		clientCert   = "/tmp/clientcert"
		username     = "foo123"
		password     = "foobar"
		loginMethod  = "device"
	)
	testData := []struct {
		name               string
		authProviderConfig map[string]string
		overrideFlags      map[string]string
		expectedArgs       []string
		execArgItems       []string
		command            string
	}{
		{
			name: "non azure kubeconfig",
		},
		{
			name: "using legacy azure auth, when convert token with msi login, client id should be empty",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "0",
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.MSILogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.MSILogin,
			},
		},
		{
			name: "using legacy azure auth, convert token with msi login and override",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "0",
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.MSILogin,
				flagClientID:    "msi-client-id",
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, "msi-client-id",
				argLoginMethod, token.MSILogin,
			},
		},
		{
			name: "using legacy azure auth, convert token with workload identity",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "0",
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.WorkloadIdentityLogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.WorkloadIdentityLogin,
			},
		},
		{
			name: "using legacy azure auth, convert token with override flags in default legacy mode",
			overrideFlags: map[string]string{
				flagEnvironment:  envName,
				flagServerID:     serverID,
				flagClientID:     clientID,
				flagTenantID:     tenantID,
				flagClientSecret: clientSecret,
				flagClientCert:   clientCert,
				flagUsername:     username,
				flagPassword:     password,
				flagLoginMethod:  loginMethod,
			},
			expectedArgs: []string{
				getTokenCommand,
				argEnvironment, envName,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argIsLegacy,
				argClientSecret, clientSecret,
				argClientCert, clientCert,
				argUsername, username,
				argPassword, password,
				argLoginMethod, loginMethod,
			},
		},
		{
			name: "using legacy azure auth, convert token with override flags overriding legacy mode",
			authProviderConfig: map[string]string{
				cfgConfigMode: "1",
			},
			overrideFlags: map[string]string{
				flagEnvironment:  envName,
				flagServerID:     serverID,
				flagClientID:     clientID,
				flagTenantID:     tenantID,
				flagClientSecret: clientSecret,
				flagClientCert:   clientCert,
				flagUsername:     username,
				flagPassword:     password,
				flagLoginMethod:  loginMethod,
				flagIsLegacy:     "true",
			},
			expectedArgs: []string{
				getTokenCommand,
				argEnvironment, envName,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argIsLegacy,
				argClientSecret, clientSecret,
				argClientCert, clientCert,
				argUsername, username,
				argPassword, password,
				argLoginMethod, loginMethod,
			},
		},
		{
			name: "using legacy azure auth, convert token in legacy mode",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
			},
			expectedArgs: []string{
				getTokenCommand,
				argEnvironment,
				envName,
				argServerID,
				serverID,
				argClientID,
				clientID,
				argTenantID,
				tenantID,
				argIsLegacy,
			},
		},
		{
			name: "using legacy azure auth, convert token in legacy mode 0",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "0",
			},
			expectedArgs: []string{
				getTokenCommand,
				argEnvironment,
				envName,
				argServerID,
				serverID,
				argClientID,
				clientID,
				argTenantID,
				tenantID,
				argIsLegacy,
			},
		},
		{
			name: "using legacy azure auth, convert token legacy azure auth",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "1",
			},
			expectedArgs: []string{
				getTokenCommand,
				argEnvironment,
				envName,
				argServerID,
				serverID,
				argClientID,
				clientID,
				argTenantID,
				tenantID,
			},
		},
		{
			name:         "exec format kubeconfig, convert from azurecli to azurecli",
			execArgItems: []string{getTokenCommand, argEnvironment, envName, argServerID, serverID, argClientID, clientID, argTenantID, tenantID, argLoginMethod, "azurecli"},
			expectedArgs: []string{getTokenCommand, argServerID, serverID, argLoginMethod, token.AzureCLILogin},
			overrideFlags: map[string]string{
				flagLoginMethod: token.AzureCLILogin,
			},
			command: execName,
		},
		{
			name:          "exec format kubeconfig, convert from azurecli to azurecli, with args as overrides",
			execArgItems:  []string{getTokenCommand},
			expectedArgs:  []string{getTokenCommand, argServerID, serverID, argLoginMethod, token.AzureCLILogin},
			overrideFlags: map[string]string{flagLoginMethod: token.AzureCLILogin, flagServerID: serverID, flagClientID: clientID, flagTenantID: tenantID, flagEnvironment: envName},
			command:       execName,
		},
		{
			name:         "exec format kubeconfig, convert from azurecli to DeviceCodeLogin",
			execArgItems: []string{getTokenCommand, argEnvironment, envName, argServerID, serverID, argClientID, clientID, argTenantID, tenantID, argLoginMethod, "azurecli"},
			expectedArgs: []string{getTokenCommand, argServerID, serverID, argClientID, clientID, argTenantID, tenantID, argLoginMethod, token.DeviceCodeLogin},
			overrideFlags: map[string]string{
				flagLoginMethod: token.DeviceCodeLogin,
			},
			command: execName,
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {
			var authProviderName string
			if data.expectedArgs != nil {
				authProviderName = azureAuthProvider
			}
			config := createValidTestConfig(clusterName, data.command, authProviderName, data.authProviderConfig, data.execArgItems)
			fs := &pflag.FlagSet{}
			o := Options{
				Flags: fs,
				configFlags: genericclioptions.NewTestConfigFlags().
					WithClientConfig(clientcmd.NewNonInteractiveClientConfig(*config, clusterName, &clientcmd.ConfigOverrides{}, nil)),
			}
			o.AddFlags(fs)

			for k, v := range data.overrideFlags {
				if err := o.setFlag(k, v); err != nil {
					t.Fatalf("unable to add flag: %s, err: %s", k, err)
				}
			}

			Convert(o)

			validate(t, config.AuthInfos[clusterName], data.authProviderConfig, data.expectedArgs)
		})
	}
}

func createValidTestConfig(name, commandName, authProviderName string, authProviderConfig map[string]string, execArgItems []string) *clientcmdapi.Config {
	const server = "https://anything.com:8080"

	config := clientcmdapi.NewConfig()
	config.Clusters[name] = &clientcmdapi.Cluster{
		Server: server,
	}

	if authProviderConfig == nil && execArgItems != nil {
		config.AuthInfos[name] = &clientcmdapi.AuthInfo{
			Exec: &clientcmdapi.ExecConfig{
				Args:    execArgItems,
				Command: commandName,
			},
		}
	} else {
		config.AuthInfos[name] = &clientcmdapi.AuthInfo{
			AuthProvider: &clientcmdapi.AuthProviderConfig{
				Name:   authProviderName,
				Config: authProviderConfig,
			},
		}
	}

	config.Contexts[name] = &clientcmdapi.Context{
		Cluster:  name,
		AuthInfo: name,
	}
	config.CurrentContext = name

	return config
}

func validate(t *testing.T, authInfo *clientcmdapi.AuthInfo, authProviderConfig map[string]string, expectedArgs []string) {
	if expectedArgs == nil {
		if authInfo.AuthProvider == nil {
			t.Fatal("original auth provider should not be reset")
		}
		if authInfo.Exec != nil {
			t.Fatal("exec plugin should not be set")
		}
		return
	}

	if authInfo.AuthProvider != nil {
		t.Fatal("original auth provider should be reset")
	}
	exec := authInfo.Exec
	if exec == nil {
		t.Fatal("unable to find exec plugin")
	}

	if exec.Command != execName {
		t.Fatalf("expected exec command: %s, actual: %s", execName, exec.Command)
	}

	if exec.APIVersion != execAPIVersion {
		t.Fatalf("expected exec command: %s, actual: %s", execAPIVersion, exec.APIVersion)
	}

	if len(exec.Env) > 0 {
		t.Fatalf("expected 0 environment variable. actual: %d", len(exec.Env))
	}
	if len(exec.Args) != len(expectedArgs) {
		t.Fatalf("expected exec args: %v, actual: %v", expectedArgs, exec.Args)
	}
	for _, v := range expectedArgs {
		if !contains(exec.Args, v) {
			t.Fatalf("expected exec arg: %s not found in %v", v, exec.Args)
		}
	}
}

func (o *Options) setFlag(key, value string) error {
	return o.Flags.Set(key, value)
}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

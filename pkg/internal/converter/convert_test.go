package converter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Azure/kubelogin/pkg/internal/token"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestConvert(t *testing.T) {
	const (
		clusterName1       = "aks1"
		clusterName2       = "aks2"
		envName            = "foo"
		serverID           = "serverID"
		clientID           = "clientID"
		spClientID         = "spClientID"
		tenantID           = "tenantID"
		clientSecret       = "foosecret"
		clientCert         = "/tmp/clientcert"
		clientCertPassword = "clientcertsecret"
		username           = "foo123"
		password           = "foobar"
		loginMethod        = "devicecode"
		identityResourceID = "/msi/resource/id"
		authorityHost      = "https://login.microsoftonline.com/"
		federatedTokenFile = "/tmp/file"
		tokenCacheDir      = "/tmp/token_dir"
		azureCLIDir        = "/tmp/foo"
	)
	testData := []struct {
		name                string
		authProviderConfig  map[string]string
		overrideFlags       map[string]string
		expectedArgs        []string
		execArgItems        []string
		command             string
		expectedExecName    string
		installHint         string
		expectedInstallHint string
		expectedError       string
		expectedEnv         []clientcmdapi.ExecEnvVar
	}{
		{
			name: "non azure kubeconfig",
		},
		{
			name:             "non azure kubeconfig in exec format with install hint",
			command:          "foo",
			expectedExecName: "foo",
			execArgItems: []string{
				"--bar",
			},
			expectedArgs: []string{
				"--bar",
			},
			installHint:         "foo install hint",
			expectedInstallHint: "foo install hint",
		},
		{
			name: "using legacy azure auth to convert to msi",
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
			name: "using legacy azure auth to convert to msi will overwrite install hint",
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
			installHint: "Overwrite this install hint",
		},
		{
			name: "using legacy azure auth to convert to msi with client-id override",
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
			name: "using legacy azure auth to convert to workload identity",
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
			name: "using legacy azure auth to convert to workload identity with overrides",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "0",
			},
			overrideFlags: map[string]string{
				flagLoginMethod:        token.WorkloadIdentityLogin,
				flagClientID:           spClientID,
				flagTenantID:           tenantID,
				flagAuthorityHost:      authorityHost,
				flagFederatedTokenFile: federatedTokenFile,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, spClientID,
				argTenantID, tenantID,
				argAuthorityHost, authorityHost,
				argFederatedTokenFile, federatedTokenFile,
				argLoginMethod, token.WorkloadIdentityLogin,
			},
		},
		{
			name: "using legacy azure auth to convert to spn without setting environment",
			authProviderConfig: map[string]string{
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "1",
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.ServicePrincipalLogin,
				flagClientID:    spClientID,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, spClientID,
				argTenantID, tenantID,
				argLoginMethod, token.ServicePrincipalLogin,
			},
		},
		{
			name: "using legacy azure auth to convert to spn with clientSecret",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "1",
			},
			overrideFlags: map[string]string{
				flagLoginMethod:  token.ServicePrincipalLogin,
				flagClientID:     spClientID,
				flagClientSecret: clientSecret,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, spClientID,
				argClientSecret, clientSecret,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.ServicePrincipalLogin,
			},
		},
		{
			name: "using legacy azure auth to convert to spn with clientCert",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "1",
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.ServicePrincipalLogin,
				flagClientID:    spClientID,
				flagClientCert:  clientCert,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, spClientID,
				argClientCert, clientCert,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.ServicePrincipalLogin,
			},
		},
		{
			name: "using legacy azure auth to convert to spn with password-protected clientCert",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "1",
			},
			overrideFlags: map[string]string{
				flagLoginMethod:        token.ServicePrincipalLogin,
				flagClientID:           spClientID,
				flagClientCert:         clientCert,
				flagClientCertPassword: clientCertPassword,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, spClientID,
				argClientCert, clientCert,
				argClientCertPassword, clientCertPassword,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.ServicePrincipalLogin,
			},
		},
		{
			name: "using legacy azure auth to convert to ropc",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "1",
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.ROPCLogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.ROPCLogin,
			},
		},
		{
			name: "using legacy azure auth to convert to ropc with username and password",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "1",
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.ROPCLogin,
				flagUsername:    username,
				flagPassword:    password,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argUsername, username,
				argPassword, password,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.ROPCLogin,
			},
		},
		{
			name: "using legacy azure auth to convert to azurecli",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "1",
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.AzureCLILogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
			},
		},
		{
			name: "using legacy azure auth to convert to azurecli with --tenant-id override",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "1",
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.AzureCLILogin,
				flagTenantID:    tenantID,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
				argTenantID, tenantID,
			},
		},
		{
			name: "using legacy azure auth to convert to azurecli with --token-cache-dir override",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "1",
			},
			overrideFlags: map[string]string{
				flagLoginMethod:   token.AzureCLILogin,
				flagTokenCacheDir: tokenCacheDir,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
				argTokenCacheDir, tokenCacheDir,
			},
		},
		{
			name: "using legacy azure auth to convert to devicecode with redundant arguments",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "0",
			},
			overrideFlags: map[string]string{
				flagEnvironment:        envName,
				flagServerID:           serverID,
				flagClientID:           clientID,
				flagTenantID:           tenantID,
				flagClientSecret:       clientSecret,
				flagClientCert:         clientCert,
				flagClientCertPassword: clientCertPassword,
				flagUsername:           username,
				flagPassword:           password,
				flagLoginMethod:        loginMethod,
			},
			expectedArgs: []string{
				getTokenCommand,
				argEnvironment, envName,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argIsLegacy,
				argLoginMethod, loginMethod,
			},
		},
		{
			name: "using legacy azure auth with configMode: \"1\" to convert to devicecode with --legacy",
			authProviderConfig: map[string]string{
				cfgConfigMode: "1",
			},
			overrideFlags: map[string]string{
				flagEnvironment: envName,
				flagServerID:    serverID,
				flagClientID:    clientID,
				flagTenantID:    tenantID,
				flagLoginMethod: loginMethod,
				flagIsLegacy:    "true",
			},
			expectedArgs: []string{
				getTokenCommand,
				argEnvironment, envName,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argIsLegacy,
				argLoginMethod, loginMethod,
			},
		},
		{
			name: "using legacy azure auth to convert without --login should default to devicecode",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
			},
			expectedArgs: []string{
				getTokenCommand,
				argEnvironment, envName,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argIsLegacy,
				argLoginMethod, token.DeviceCodeLogin,
			},
		},
		{
			name: "using legacy azure auth with configMode: \"0\" to convert without --login should default to devicecode",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "0",
			},
			expectedArgs: []string{
				getTokenCommand,
				argEnvironment, envName,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argIsLegacy,
				argLoginMethod, token.DeviceCodeLogin,
			},
		},
		{
			name: "using legacy azure auth with configMode: \"1\" to convert without --login should result in devicecode without --legacy",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "1",
			},
			expectedArgs: []string{
				getTokenCommand,
				argEnvironment, envName,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argLoginMethod, token.DeviceCodeLogin,
			},
		},
		{
			name: "with exec format kubeconfig, convert from azurecli to azurecli",
			execArgItems: []string{
				getTokenCommand,
				argEnvironment, envName,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argLoginMethod, token.AzureCLILogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.AzureCLILogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from azurecli to azurecli with existing install hint",
			execArgItems: []string{
				getTokenCommand,
				argEnvironment, envName,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argLoginMethod, token.AzureCLILogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.AzureCLILogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
			},
			command:             execName,
			installHint:         "Preserve this install hint",
			expectedInstallHint: "Preserve this install hint",
		},
		{
			name: "with exec format kubeconfig, convert from azurecli to azurecli with --tenant-id",
			execArgItems: []string{
				getTokenCommand,
				argEnvironment, envName,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argLoginMethod, token.AzureCLILogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.AzureCLILogin,
				flagTenantID:    tenantID,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
				argTenantID, tenantID,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from azurecli to azurecli, with envName as overrides",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argLoginMethod, token.AzureCLILogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.AzureCLILogin,
				flagEnvironment: envName,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from azurecli to azurecli, with args as overrides",
			execArgItems: []string{
				getTokenCommand,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.AzureCLILogin,
				flagServerID:    serverID,
				flagClientID:    clientID,
				flagTenantID:    tenantID,
				flagEnvironment: envName,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
				argTenantID, tenantID,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from azurecli to devicecode",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
			},
			overrideFlags: map[string]string{
				flagClientID:    clientID,
				flagTenantID:    tenantID,
				flagLoginMethod: token.DeviceCodeLogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argLoginMethod, token.DeviceCodeLogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from azurecli to devicecode with existing install hint",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
			},
			overrideFlags: map[string]string{
				flagClientID:    clientID,
				flagTenantID:    tenantID,
				flagLoginMethod: token.DeviceCodeLogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argLoginMethod, token.DeviceCodeLogin,
			},
			command:             execName,
			installHint:         "Preserve this install hint",
			expectedInstallHint: "Preserve this install hint",
		},
		{
			name: "with exec format kubeconfig, convert from azurecli to devicecode, with args as overrides",
			execArgItems: []string{
				getTokenCommand,
				argLoginMethod, token.AzureCLILogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.DeviceCodeLogin,
				flagServerID:    serverID,
				flagClientID:    clientID,
				flagTenantID:    tenantID,
				flagEnvironment: envName,
			},
			expectedArgs: []string{
				getTokenCommand,
				argEnvironment, envName,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argLoginMethod, token.DeviceCodeLogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to devicecode without override",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to devicecode with --legacy",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagIsLegacy: "true",
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argIsLegacy,
				argLoginMethod, token.DeviceCodeLogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig using devicecode and --legacy, convert to devicecode should still have --legacy",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
				argIsLegacy,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.DeviceCodeLogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argIsLegacy,
				argLoginMethod, token.DeviceCodeLogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to azurecli",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.AzureCLILogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to azurecli with --token-cache-dir override",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod:   token.AzureCLILogin,
				flagTokenCacheDir: tokenCacheDir,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
				argTokenCacheDir, tokenCacheDir,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig already having --token-cache-dir, convert from devicecode to azurecli",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argTokenCacheDir, tokenCacheDir,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.AzureCLILogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
				argTokenCacheDir, tokenCacheDir,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to spn",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.ServicePrincipalLogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argEnvironment, envName,
				argServerID, serverID,
				argTenantID, tenantID,
				argClientID, clientID,
				argLoginMethod, token.ServicePrincipalLogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to spn without setting environment",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.ServicePrincipalLogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argTenantID, tenantID,
				argClientID, clientID,
				argLoginMethod, token.ServicePrincipalLogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to spn with clientID",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.ServicePrincipalLogin,
				flagClientID:    spClientID,
			},
			expectedArgs: []string{
				getTokenCommand,
				argEnvironment, envName,
				argServerID, serverID,
				argClientID, spClientID,
				argTenantID, tenantID,
				argLoginMethod, token.ServicePrincipalLogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to spn with --legacy",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.ServicePrincipalLogin,
				flagClientID:    spClientID,
				flagIsLegacy:    "true",
			},
			expectedArgs: []string{
				getTokenCommand,
				argEnvironment, envName,
				argServerID, serverID,
				argClientID, spClientID,
				argTenantID, tenantID,
				argIsLegacy,
				argLoginMethod, token.ServicePrincipalLogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to msi",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.MSILogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.MSILogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to msi with clientID override",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.MSILogin,
				flagClientID:    spClientID,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, spClientID,
				argLoginMethod, token.MSILogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to msi with identity-resource-id override",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod:        token.MSILogin,
				flagIdentityResourceID: identityResourceID,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argIdentityResourceID, identityResourceID,
				argLoginMethod, token.MSILogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to ropc",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.ROPCLogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.ROPCLogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to ropc with --legacy",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.ROPCLogin,
				flagIsLegacy:    "true",
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argIsLegacy,
				argLoginMethod, token.ROPCLogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to ropc with username and password",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.ROPCLogin,
				flagUsername:    username,
				flagPassword:    password,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argUsername, username,
				argPassword, password,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.ROPCLogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to workload identity",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.WorkloadIdentityLogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.WorkloadIdentityLogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to workload identity with override",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod:        token.WorkloadIdentityLogin,
				flagClientID:           spClientID,
				flagTenantID:           tenantID,
				flagAuthorityHost:      authorityHost,
				flagFederatedTokenFile: federatedTokenFile,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, spClientID,
				argTenantID, tenantID,
				argAuthorityHost, authorityHost,
				argFederatedTokenFile, federatedTokenFile,
				argLoginMethod, token.WorkloadIdentityLogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to interactive",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.InteractiveLogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.InteractiveLogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to interactive without setting environment",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.InteractiveLogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argTenantID, tenantID,
				argClientID, clientID,
				argLoginMethod, token.InteractiveLogin,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to interactive with override",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.InteractiveLogin,
				flagServerID:    serverID,
				flagClientID:    clientID,
				flagTenantID:    tenantID,
				flagEnvironment: envName,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.InteractiveLogin,
			},
			command: execName,
		},
		{
			name: "convert with context specified, auth info not specified by the context should not be changed",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "0",
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.MSILogin,
				flagContext:     clusterName1,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.MSILogin,
			},
		},
		{
			name: "convert with non-existent context specified, Convert should return error",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "0",
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.MSILogin,
				flagContext:     "badContext",
			},
			expectedError: "no context exists with the name: \"badContext\"",
		},
		{
			name: "with --azure-config-dir specified, exec.Env should be set accordingly",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "0",
			},
			overrideFlags: map[string]string{
				flagLoginMethod:    token.AzureCLILogin,
				flagContext:        clusterName1,
				flagAzureConfigDir: azureCLIDir,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
			},
			expectedEnv: []clientcmdapi.ExecEnvVar{
				{
					Name:  azureConfigDir,
					Value: azureCLIDir,
				},
			},
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to interactive with only pop-enabled specified, Convert should return error",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
				argIsPoPTokenEnabled,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.InteractiveLogin,
			},
			command:       execName,
			expectedError: "--pop-claims is required when specifying --pop-enabled",
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to interactive with only pop-claims specified, Convert should return error",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
				argPoPTokenClaims, "u=testhost",
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.InteractiveLogin,
			},
			command:       execName,
			expectedError: "--pop-enabled is required when specifying --pop-claims",
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to interactive with pop-enabled and pop-claims",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
				argIsPoPTokenEnabled,
				argPoPTokenClaims, "u=testhost, 1=2",
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.InteractiveLogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.InteractiveLogin,
				argIsPoPTokenEnabled,
				argPoPTokenClaims, "u=testhost, 1=2",
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to spn with pop-enabled and pop-claims as flags",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod:       token.ServicePrincipalLogin,
				flagIsPoPTokenEnabled: "true",
				flagPoPTokenClaims:    "u=testhost, 1=2",
			},
			expectedArgs: []string{
				getTokenCommand,
				argEnvironment, envName,
				argServerID, serverID,
				argTenantID, tenantID,
				argClientID, clientID,
				argLoginMethod, token.ServicePrincipalLogin,
				argIsPoPTokenEnabled,
				argPoPTokenClaims, "u=testhost, 1=2",
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from azurecli to devicecode with pop-enabled and pop-claims, expect pop args to be ignored",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
				argIsPoPTokenEnabled,
				argPoPTokenClaims, "u=testhost, 1=2",
			},
			overrideFlags: map[string]string{
				flagClientID:    clientID,
				flagTenantID:    tenantID,
				flagLoginMethod: token.DeviceCodeLogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argLoginMethod, token.DeviceCodeLogin,
			},
			command: execName,
		},
	}
	rootTmpDir, err := os.MkdirTemp("", "kubelogin-test")
	if err != nil {
		t.Fatalf("unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(rootTmpDir)
	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {
			var authProviderName string
			tmpDir, err := os.MkdirTemp(rootTmpDir, "config")
			if err != nil {
				t.Fatalf("%s", err)
			}
			if data.expectedArgs != nil {
				authProviderName = azureAuthProvider
			}
			kubeconfigFile := filepath.Join(tmpDir, "config")

			config := createValidTestConfigs(
				clusterName1,
				clusterName2,
				data.command,
				authProviderName,
				data.authProviderConfig,
				data.execArgItems,
				data.installHint,
			)
			fs := &pflag.FlagSet{}
			o := Options{
				Flags: fs,
				configFlags: genericclioptions.NewTestConfigFlags().
					WithClientConfig(clientcmd.NewNonInteractiveClientConfig(*config, clusterName1, &clientcmd.ConfigOverrides{}, nil)),
			}
			o.AddFlags(fs)

			for k, v := range data.overrideFlags {
				if err := o.setFlag(k, v); err != nil {
					t.Fatalf("unable to add flag: %s, err: %s", k, err)
				}
			}

			pathOptions := clientcmd.PathOptions{
				ExplicitFileFlag: "kubeconfig",
				LoadingRules: &clientcmd.ClientConfigLoadingRules{
					ExplicitPath: kubeconfigFile,
				},
			}
			err = Convert(o, &pathOptions)
			if data.expectedError == "" && err != nil {
				t.Fatalf("Unexpected error from Convert: %v", err)
			} else if data.expectedError != "" {
				if err == nil || err.Error() != data.expectedError {
					t.Fatalf("Expected error: %q, but got: %q", data.expectedError, err)
				}
			} else {
				// only need to validate fields if we're not expecting an error
				if o.context != "" {
					// when --context is specified, convert-kubeconfig will convert only the targeted context
					// hence, we expect the second auth info not to change
					validate(t, clusterName1, config.AuthInfos[clusterName1], data.authProviderConfig, data.expectedArgs, data.expectedExecName, data.expectedInstallHint, data.expectedEnv)
					validateAuthInfoThatShouldNotChange(t, clusterName2, config.AuthInfos[clusterName2], data.authProviderConfig)
				} else {
					// when --context is not specified, convert-kubeconfig will convert every auth info in the kubeconfig
					// hence, we expect the second auth info to be converted in the same way as the first one
					validate(t, clusterName1, config.AuthInfos[clusterName1], data.authProviderConfig, data.expectedArgs, data.expectedExecName, data.expectedInstallHint, data.expectedEnv)
					validate(t, clusterName2, config.AuthInfos[clusterName2], data.authProviderConfig, data.expectedArgs, data.expectedExecName, data.expectedInstallHint, data.expectedEnv)
				}
			}
		})
	}
}

func createValidTestConfigs(
	name1, name2, commandName, authProviderName string,
	authProviderConfig map[string]string,
	execArgItems []string,
	installHint string,
) *clientcmdapi.Config {
	const server = "https://anything.com:8080"

	config := clientcmdapi.NewConfig()
	for _, name := range []string{name1, name2} {
		config.Clusters[name] = &clientcmdapi.Cluster{
			Server: server,
		}

		if authProviderConfig == nil && execArgItems != nil {
			config.AuthInfos[name] = &clientcmdapi.AuthInfo{
				Exec: &clientcmdapi.ExecConfig{
					Args:        execArgItems,
					Command:     commandName,
					InstallHint: installHint,
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
	}
	config.CurrentContext = name1

	return config
}

func validate(
	t *testing.T,
	clusterName string,
	authInfo *clientcmdapi.AuthInfo,
	authProviderConfig map[string]string,
	expectedArgs []string,
	expectedExecName string,
	expectedInstallHint string,
	expectedEnv []clientcmdapi.ExecEnvVar,
) {
	if expectedArgs == nil {
		if authInfo.AuthProvider == nil {
			t.Fatalf("[context:%s]: %s", clusterName, "auth provider should not be reset")
		}
		if authInfo.Exec != nil {
			t.Fatalf("[context:%s]: %s", clusterName, "plugin should not be set")
		}
		return
	}

	if authInfo.AuthProvider != nil {
		t.Fatalf("[context:%s]: %s", clusterName, "auth provider should be reset")
	}
	exec := authInfo.Exec
	if exec == nil {
		t.Fatalf("[context:%s]: %s", clusterName, "unable to find exec plugin")
	}

	// default to the kubelogin exec name
	if expectedExecName == "" {
		expectedExecName = execName
	}

	if exec.Command != expectedExecName {
		t.Fatalf("[context:%s]: expected exec command: %s, actual: %s", clusterName, expectedExecName, exec.Command)
	}

	// default to the kubelogin install hint
	if expectedInstallHint == "" {
		expectedInstallHint = execInstallHint
	}

	if exec.InstallHint != expectedInstallHint {
		t.Fatalf("[context:%s]: expected install hint: %s, actual: %s", clusterName, expectedInstallHint, exec.InstallHint)
	}

	// Only validate the API version and first arg if exec is using kubelogin
	if exec.Command == execName {
		if exec.APIVersion != execAPIVersion {
			t.Fatalf("[context:%s]: expected API Version: %s, actual: %s", clusterName, execAPIVersion, exec.APIVersion)
		}

		if exec.Args[0] != getTokenCommand {
			t.Fatalf("[context:%s]: expected %s as first argument. actual: %s", clusterName, getTokenCommand, exec.Args[0])
		}
	}

	if len(exec.Args) != len(expectedArgs) {
		t.Fatalf("[context:%s]: expected exec args: %v, actual: %v", clusterName, expectedArgs, exec.Args)
	}
	for _, v := range expectedArgs {
		if !contains(exec.Args, v) {
			t.Fatalf("[context:%s]: expected exec arg: %s not found in %v", clusterName, v, exec.Args)
		}
	}
	if len(expectedEnv) != len(exec.Env) {
		t.Fatalf("[context:%s]: expected Env has %d entries, got %d", clusterName, len(expectedEnv), len(exec.Env))
	}
	for i, v := range expectedEnv {
		if exec.Env[i] != v {
			t.Fatalf("[context:%s]: for exec.Env, expected %q at index %d, got %q", clusterName, v, i, exec.Env[i])
		}
	}
}

func validateAuthInfoThatShouldNotChange(
	t *testing.T,
	clusterName string,
	authInfo *clientcmdapi.AuthInfo,
	authProviderConfig map[string]string,
) {
	if authInfo.AuthProvider == nil {
		t.Fatalf("[context:%s]: %s", clusterName, "auth provider should not be reset")
	}
	for k, v := range authInfo.AuthProvider.Config {
		if authProviderConfig[k] != v {
			t.Fatalf("[context:%s]: %s=%s does not match with input %s=%s", clusterName, k, v, k, authProviderConfig[k])
		}
	}
	for k, v := range authProviderConfig {
		if authInfo.AuthProvider.Config[k] != v {
			t.Fatalf("[context:%s]: %s=%s does not match with output %s=%s", clusterName, k, v, k, authInfo.AuthProvider.Config[k])
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

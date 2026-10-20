package converter

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
		subscriptionID     = "subscriptionID"
		clientSecret       = "foosecret"
		clientCert         = "/tmp/clientcert"
		clientCertPassword = "clientcertsecret"
		username           = "foo123"
		password           = "foobar"
		loginMethod        = "devicecode"
		identityResourceID = "/msi/resource/id"
		authorityHost      = "https://login.microsoftonline.com/"
		federatedTokenFile = "/tmp/file"
		authRecordCacheDir = "/tmp/token_dir"
		azureCLIDir        = "/tmp/foo"
		redirectURL        = "http://localhost:8000"
		usernameHint       = "username"
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
			name: "using legacy azure auth to convert to azd",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "1",
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.AzureDeveloperCLILogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureDeveloperCLILogin,
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
			name: "using legacy azure auth to convert to azd with --tenant-id override",
			authProviderConfig: map[string]string{
				cfgEnvironment: envName,
				cfgApiserverID: serverID,
				cfgClientID:    clientID,
				cfgTenantID:    tenantID,
				cfgConfigMode:  "1",
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.AzureDeveloperCLILogin,
				flagTenantID:    tenantID,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureDeveloperCLILogin,
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
				flagTokenCacheDir: authRecordCacheDir,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
				argAuthRecordCacheDir, authRecordCacheDir,
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
			name: "with exec format kubeconfig, convert from azurecli to azurecli with --subscription",
			execArgItems: []string{
				getTokenCommand,
				argEnvironment, envName,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argLoginMethod, token.AzureCLILogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod:    token.AzureCLILogin,
				flagSubscriptionID: subscriptionID,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
				argSubscriptionID, subscriptionID,
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
				flagTokenCacheDir: authRecordCacheDir,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
				argAuthRecordCacheDir, authRecordCacheDir,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to azurecli with --cache-dir",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
				argAuthRecordCacheDir, authRecordCacheDir,
			},
			overrideFlags: map[string]string{
				flagLoginMethod: token.AzureCLILogin,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
				argAuthRecordCacheDir, authRecordCacheDir,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to azurecli with --cache-dir override",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod:        token.AzureCLILogin,
				flagAuthRecordCacheDir: authRecordCacheDir,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argLoginMethod, token.AzureCLILogin,
				argAuthRecordCacheDir, authRecordCacheDir,
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
				argTokenCacheDir, authRecordCacheDir,
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
				argAuthRecordCacheDir, authRecordCacheDir,
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
			name: "with exec format kubeconfig, convert from devicecode to interactive with redirect url override",
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
				flagRedirectURL: redirectURL,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.InteractiveLogin,
				argRedirectURL, redirectURL,
			},
			command: execName,
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to interactive with login hint override",
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
				flagLoginHint:   usernameHint,
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.InteractiveLogin,
				argLoginHint, usernameHint,
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
			name: "with exec format kubeconfig, convert from devicecode to ropc with only pop-enabled specified, Convert should return error",
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
				flagLoginMethod: token.ROPCLogin,
			},
			command:       execName,
			expectedError: "--pop-claims is required when specifying --pop-enabled",
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to ropc with only pop-claims specified, Convert should return error",
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
				flagLoginMethod: token.ROPCLogin,
			},
			command:       execName,
			expectedError: "--pop-enabled is required when specifying --pop-claims",
		},
		{
			name: "with exec format kubeconfig, convert from devicecode to ropc with pop-enabled and pop-claims as flags",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod:       token.ROPCLogin,
				flagIsPoPTokenEnabled: "true",
				flagPoPTokenClaims:    "u=testhost, 1=2",
			},
			expectedArgs: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.ROPCLogin,
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
		{
			name: "test with exec format kubeconfig, convert from devicecode to spn with environment override flag disabled.",
			execArgItems: []string{
				getTokenCommand,
				argServerID, serverID,
				argClientID, clientID,
				argTenantID, tenantID,
				argEnvironment, envName,
				argLoginMethod, token.DeviceCodeLogin,
			},
			overrideFlags: map[string]string{
				flagLoginMethod:                token.ServicePrincipalLogin,
				flagDisableEnvironmentOverride: "true",
			},
			expectedArgs: []string{
				getTokenCommand,
				argEnvironment, envName,
				argServerID, serverID,
				argTenantID, tenantID,
				argClientID, clientID,
				argLoginMethod, token.ServicePrincipalLogin,
				argDisableEnvironmentOverride,
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
					validate(t, clusterName1, config.AuthInfos[clusterName1], data.expectedArgs, data.expectedExecName, data.expectedInstallHint, data.expectedEnv)
					validateAuthInfoThatShouldNotChange(t, clusterName2, config.AuthInfos[clusterName2], data.authProviderConfig)
				} else {
					// when --context is not specified, convert-kubeconfig will convert every auth info in the kubeconfig
					// hence, we expect the second auth info to be converted in the same way as the first one
					validate(t, clusterName1, config.AuthInfos[clusterName1], data.expectedArgs, data.expectedExecName, data.expectedInstallHint, data.expectedEnv)
					validate(t, clusterName2, config.AuthInfos[clusterName2], data.expectedArgs, data.expectedExecName, data.expectedInstallHint, data.expectedEnv)
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

func TestConvertOutFile(t *testing.T) {
	const (
		serverID = "serverID"
		clientID = "clientID"
		tenantID = "tenantID"
		caFile   = "ca.crt"
	)
	authProviderConfig := map[string]string{
		cfgApiserverID: serverID,
		cfgClientID:    clientID,
		cfgTenantID:    tenantID,
		cfgConfigMode:  "1",
	}
	expectedArgs := []string{
		getTokenCommand,
		argServerID, serverID,
		argClientID, clientID,
		argTenantID, tenantID,
		argLoginMethod, token.DeviceCodeLogin,
	}

	testData := []struct {
		name string
		// sources holds the user names stored in each source kubeconfig file.
		// More than one entry loads them as a merged KUBECONFIG.
		sources [][2]string
		// explicitKubeconfig passes the first source with --kubeconfig instead of KUBECONFIG
		explicitKubeconfig bool
		// missingKubeconfig points KUBECONFIG at a file that does not exist
		missingKubeconfig bool
		// outFile is relative to the test directory; empty leaves --outfile unset
		outFile string
		// alias makes outFile a "symlink" or "hardlink" to the first source before converting
		alias         string
		context       string
		expectedError string
	}{
		{
			name:               "outfile with --kubeconfig leaves the source untouched",
			sources:            [][2]string{{"aks1", "aks2"}},
			explicitKubeconfig: true,
			outFile:            "converted.yaml",
		},
		{
			name:    "outfile with merged KUBECONFIG files gets the merged result",
			sources: [][2]string{{"aks1", "aks2"}, {"aks3", "aks4"}},
			outFile: "converted.yaml",
		},
		{
			name:               "outfile with --context converts only the targeted user",
			sources:            [][2]string{{"aks1", "aks2"}},
			explicitKubeconfig: true,
			outFile:            "converted.yaml",
			context:            "aks1",
		},
		{
			name:               "outfile in a missing directory creates the directory",
			sources:            [][2]string{{"aks1", "aks2"}},
			explicitKubeconfig: true,
			outFile:            filepath.Join("generated", "team", "config"),
		},
		{
			name:               "outfile whose parent is a regular file fails",
			sources:            [][2]string{{"aks1", "aks2"}},
			explicitKubeconfig: true,
			outFile:            filepath.Join("source-0", "config", "converted.yaml"),
			expectedError:      "unable to write kubeconfig to",
		},
		{
			name:               "outfile that is the --kubeconfig source fails",
			sources:            [][2]string{{"aks1", "aks2"}},
			explicitKubeconfig: true,
			outFile:            filepath.Join("source-0", "config"),
			expectedError:      "is the same file as source kubeconfig",
		},
		{
			name:               "outfile that is a symlink to the source fails",
			sources:            [][2]string{{"aks1", "aks2"}},
			explicitKubeconfig: true,
			outFile:            "link.yaml",
			alias:              "symlink",
			expectedError:      "is the same file as source kubeconfig",
		},
		{
			name:               "outfile that is a hard link to the source fails",
			sources:            [][2]string{{"aks1", "aks2"}},
			explicitKubeconfig: true,
			outFile:            "link.yaml",
			alias:              "hardlink",
			expectedError:      "is the same file as source kubeconfig",
		},
		{
			name:          "outfile that is one of the merged KUBECONFIG files fails",
			sources:       [][2]string{{"aks1", "aks2"}, {"aks3", "aks4"}},
			outFile:       filepath.Join("source-1", "config"),
			expectedError: "is the same file as source kubeconfig",
		},
		{
			name:               "outfile that is an existing directory fails",
			sources:            [][2]string{{"aks1", "aks2"}},
			explicitKubeconfig: true,
			outFile:            "source-0",
			expectedError:      "unable to write kubeconfig to",
		},
		{
			name:              "outfile with no kubeconfig to convert fails and writes nothing",
			missingKubeconfig: true,
			outFile:           "converted.yaml",
			expectedError:     "no kubeconfig found to convert",
		},
		{
			name:               "without outfile the source kubeconfig is converted in place",
			sources:            [][2]string{{"aks1", "aks2"}},
			explicitKubeconfig: true,
		},
		{
			name:    "without outfile merged KUBECONFIG files are converted in place",
			sources: [][2]string{{"aks1", "aks2"}, {"aks3", "aks4"}},
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			var sourceFiles []string
			before := map[string][]byte{}
			for i, names := range data.sources {
				config := createValidTestConfigs(names[0], names[1], "", azureAuthProvider, authProviderConfig, nil, "")
				for _, cluster := range config.Clusters {
					// relative to the source file, so the outfile must not keep it as is
					cluster.CertificateAuthority = caFile
				}
				file := filepath.Join(tmpDir, fmt.Sprintf("source-%d", i), "config")
				if err := clientcmd.WriteToFile(*config, file); err != nil {
					t.Fatalf("unable to write source kubeconfig: %s", err)
				}
				content, err := os.ReadFile(file)
				if err != nil {
					t.Fatalf("unable to read source kubeconfig: %s", err)
				}
				sourceFiles = append(sourceFiles, file)
				before[file] = content
			}
			kubeconfigEnv := strings.Join(sourceFiles, string(os.PathListSeparator))
			if data.missingKubeconfig {
				kubeconfigEnv = filepath.Join(tmpDir, "missing", "config")
			}
			t.Setenv(clientcmd.RecommendedConfigPathEnvVar, kubeconfigEnv)

			o := New()
			fs := &pflag.FlagSet{}
			o.Flags = fs
			o.AddFlags(fs)
			overrideFlags := map[string]string{}
			if data.explicitKubeconfig {
				overrideFlags["kubeconfig"] = sourceFiles[0]
			}
			outFile := ""
			if data.outFile != "" {
				outFile = filepath.Join(tmpDir, data.outFile)
				switch data.alias {
				case "symlink":
					if err := os.Symlink(sourceFiles[0], outFile); err != nil {
						t.Skipf("symlinks are not supported: %s", err)
					}
				case "hardlink":
					if err := os.Link(sourceFiles[0], outFile); err != nil {
						t.Skipf("hard links are not supported: %s", err)
					}
				}
				overrideFlags[flagOutFile] = outFile
			}
			if data.context != "" {
				overrideFlags[flagContext] = data.context
			}
			for k, v := range overrideFlags {
				if err := o.setFlag(k, v); err != nil {
					t.Fatalf("unable to add flag: %s, err: %s", k, err)
				}
			}

			// same wiring as the convert-kubeconfig command
			pathOptions := clientcmd.NewDefaultPathOptions()
			pathOptions.LoadingRules.ExplicitPath, _ = o.Flags.GetString("kubeconfig")

			err := Convert(o, pathOptions)
			if data.expectedError != "" {
				if err == nil || !strings.Contains(err.Error(), data.expectedError) {
					t.Fatalf("expected error containing %q, got: %v", data.expectedError, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error from Convert: %v", err)
			}

			if outFile == "" {
				// old behaviour: every source file is converted in place
				for i, file := range sourceFiles {
					got, err := clientcmd.LoadFromFile(file)
					if err != nil {
						t.Fatalf("unable to load source kubeconfig: %s", err)
					}
					for _, name := range data.sources[i] {
						validate(t, name, got.AuthInfos[name], expectedArgs, "", "", nil)
					}
				}
				return
			}

			for _, file := range sourceFiles {
				after, err := os.ReadFile(file)
				if err != nil {
					t.Fatalf("unable to read source kubeconfig: %s", err)
				}
				if !bytes.Equal(before[file], after) {
					t.Fatalf("source kubeconfig %s was modified:\nbefore:\n%s\nafter:\n%s", file, before[file], after)
				}
			}

			if data.expectedError != "" {
				if data.missingKubeconfig {
					if _, err := os.Stat(outFile); !os.IsNotExist(err) {
						t.Fatalf("expected no outfile to be written, got: %v", err)
					}
				}
				return
			}

			info, err := os.Stat(outFile)
			if err != nil {
				t.Fatalf("outfile was not written: %s", err)
			}
			if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
				t.Fatalf("expected outfile mode 0600, got %o", info.Mode().Perm())
			}

			got, err := clientcmd.LoadFromFile(outFile)
			if err != nil {
				t.Fatalf("unable to load outfile: %s", err)
			}
			if got.CurrentContext != data.sources[0][0] {
				t.Fatalf("expected current context %q, got %q", data.sources[0][0], got.CurrentContext)
			}
			total := 0
			for i, names := range data.sources {
				for _, name := range names {
					total++
					if data.context != "" && name != data.context {
						validateAuthInfoThatShouldNotChange(t, name, got.AuthInfos[name], authProviderConfig)
					} else {
						validate(t, name, got.AuthInfos[name], expectedArgs, "", "", nil)
					}
					if got.Contexts[name] == nil || got.Clusters[name] == nil {
						t.Fatalf("[context:%s]: context or cluster missing from outfile", name)
					}
					expectedCA := filepath.Join(filepath.Dir(sourceFiles[i]), caFile)
					if got.Clusters[name].CertificateAuthority != expectedCA {
						t.Fatalf("[context:%s]: expected certificate-authority %q, got %q", name, expectedCA, got.Clusters[name].CertificateAuthority)
					}
				}
			}
			if len(got.AuthInfos) != total || len(got.Contexts) != total || len(got.Clusters) != total {
				t.Fatalf("expected %d users, contexts and clusters in outfile, got %d, %d, %d", total, len(got.AuthInfos), len(got.Contexts), len(got.Clusters))
			}
		})
	}
}

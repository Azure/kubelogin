package options

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestIsSet(t *testing.T) {
	tests := map[string]struct {
		flagName string
		setValue bool
		expected bool
	}{
		"flag explicitly set": {"client-id", true, true},
		"flag not set":        {"client-id", false, false},
		"flag does not exist": {"non-existent", false, false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			opts := NewUnifiedOptions(ConvertCommand)
			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)

			if tc.flagName == "client-id" {
				fs.String("client-id", "", "test flag")
				if tc.setValue {
					fs.Set("client-id", "test-value")
				}
			}
			opts.flags = fs

			result := opts.isSet(tc.flagName)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestShouldIncludeArgComprehensive(t *testing.T) {
	tests := map[string]struct {
		loginMethod string
		fieldName   string
		value       string
		clientCert  string // For SPN cert logic
		flagSet     bool   // Whether flag was explicitly set by user
		expected    bool
	}{
		// Azure CLI - excludes client-id/tenant-id, special cache-dir handling
		"azurecli excludes client-id":              {"azurecli", "ClientID", "test", "", false, false},
		"azurecli excludes tenant-id":              {"azurecli", "TenantID", "test", "", false, false},
		"azurecli includes server-id":              {"azurecli", "ServerID", "test", "", false, true},
		"azurecli includes cache-dir when set":     {"azurecli", "AuthRecordCacheDir", "/cache", "", true, true},
		"azurecli excludes cache-dir when not set": {"azurecli", "AuthRecordCacheDir", "/cache", "", false, false},

		// MSI - only includes client-id/identity-resource-id when explicitly set
		"msi includes client-id when set":                {"msi", "ClientID", "test", "", true, true},
		"msi excludes client-id when not set":            {"msi", "ClientID", "test", "", false, false},
		"msi includes identity-resource-id when set":     {"msi", "IdentityResourceID", "/sub/rg/id", "", true, true},
		"msi excludes identity-resource-id when not set": {"msi", "IdentityResourceID", "/sub/rg/id", "", false, false},
		"msi includes server-id":                         {"msi", "ServerID", "test", "", false, true},
		"msi excludes tenant-id":                         {"msi", "TenantID", "test", "", false, false},
		"msi excludes client-secret":                     {"msi", "ClientSecret", "test", "", false, false},

		// Service Principal - mutual exclusion of cert vs secret
		"spn includes client-id":                  {"spn", "ClientID", "test", "", false, true},
		"spn includes tenant-id":                  {"spn", "TenantID", "test", "", false, true},
		"spn includes server-id":                  {"spn", "ServerID", "test", "", false, true},
		"spn includes secret when no cert":        {"spn", "ClientSecret", "secret", "", false, true},
		"spn excludes secret when cert provided":  {"spn", "ClientSecret", "secret", "/cert.pfx", false, false},
		"spn includes cert when provided":         {"spn", "ClientCert", "/cert.pfx", "/cert.pfx", false, true},
		"spn excludes cert when not provided":     {"spn", "ClientCert", "", "", false, false},
		"spn includes cert password with cert":    {"spn", "ClientCertPassword", "pass", "/cert.pfx", false, true},
		"spn excludes cert password without cert": {"spn", "ClientCertPassword", "pass", "", false, false},

		// DeviceCode - includes client-id/tenant-id
		"devicecode includes client-id":     {"devicecode", "ClientID", "test", "", false, true},
		"devicecode includes tenant-id":     {"devicecode", "TenantID", "test", "", false, true},
		"devicecode excludes client-secret": {"devicecode", "ClientSecret", "test", "", false, false},

		// Interactive - includes client-id/tenant-id/login-hint
		"interactive includes client-id":  {"interactive", "ClientID", "test", "", false, true},
		"interactive includes tenant-id":  {"interactive", "TenantID", "test", "", false, true},
		"interactive includes login-hint": {"interactive", "LoginHint", "user@test.com", "", false, true},

		// ROPC - includes client-id/tenant-id/username/password
		"ropc includes client-id":     {"ropc", "ClientID", "test", "", false, true},
		"ropc includes tenant-id":     {"ropc", "TenantID", "test", "", false, true},
		"ropc includes username":      {"ropc", "Username", "user", "", false, true},
		"ropc includes password":      {"ropc", "Password", "pass", "", false, true},
		"ropc excludes client-secret": {"ropc", "ClientSecret", "test", "", false, false},

		// Workload Identity - includes client-id/tenant-id/federated-token/authority
		"workloadidentity includes client-id":       {"workloadidentity", "ClientID", "test", "", false, true},
		"workloadidentity includes tenant-id":       {"workloadidentity", "TenantID", "test", "", false, true},
		"workloadidentity includes federated-token": {"workloadidentity", "FederatedTokenFile", "/token", "", false, true},
		"workloadidentity includes authority-host":  {"workloadidentity", "AuthorityHost", "https://login.ms", "", false, true},
		"workloadidentity excludes client-secret":   {"workloadidentity", "ClientSecret", "test", "", false, false},

		// Azure Developer CLI - same as Azure CLI
		"azd excludes client-id":              {"azd", "ClientID", "test", "", false, false},
		"azd includes server-id":              {"azd", "ServerID", "test", "", false, true},
		"azd includes cache-dir when set":     {"azd", "AuthRecordCacheDir", "/cache", "", true, true},
		"azd excludes cache-dir when not set": {"azd", "AuthRecordCacheDir", "/cache", "", false, false},

		// ServerID always included for all methods
		"devicecode includes server-id":       {"devicecode", "ServerID", "test", "", false, true},
		"interactive includes server-id":      {"interactive", "ServerID", "test", "", false, true},
		"ropc includes server-id":             {"ropc", "ServerID", "test", "", false, true},
		"workloadidentity includes server-id": {"workloadidentity", "ServerID", "test", "", false, true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			opts := NewUnifiedOptions(ConvertCommand)
			opts.LoginMethod = tc.loginMethod
			opts.ClientCert = tc.clientCert

			// Setup flag tracking for tests that check isSet()
			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flagName := getCorrespondingFlagName(tc.fieldName)
			if flagName != "" {
				fs.String(flagName, "", "test flag")
				if tc.flagSet {
					fs.Set(flagName, tc.value)
				}
			}
			opts.flags = fs

			result := opts.shouldIncludeArg(tc.fieldName, tc.value)
			assert.Equal(t, tc.expected, result, "Login: %s, Field: %s", tc.loginMethod, tc.fieldName)
		})
	}
}

// Helper function to map field names to flag names
func getCorrespondingFlagName(fieldName string) string {
	flagMap := map[string]string{
		"ClientID":           "client-id",
		"IdentityResourceID": "identity-resource-id",
		"AuthRecordCacheDir": "cache-dir",
	}
	return flagMap[fieldName]
}

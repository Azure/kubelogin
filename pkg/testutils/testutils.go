package testutils

import "strings"

const (
	ClientID       = "AZURE_CLIENT_ID"
	ClientSecret   = "AAD_SERVICE_PRINCIPAL_CLIENT_SECRET"
	ClientCert     = "AZURE_CLIENT_CER"
	ClientCertPass = "AZURE_CLIENT_CERTIFICATE_PASSWORD"
	ResourceID     = "AZURE_RESOURCE_ID"
	TenantID       = "AZURE_TENANT_ID"
	BadSecret      = "Bad_Secret"
)

// ErrorContains takes an input error and a desired substring, checks if the string is present
// in the error message, and returns the boolean result
func ErrorContains(out error, want string) bool {
	substring := strings.TrimSpace(want)
	if out == nil {
		return substring == ""
	}
	if substring == "" {
		return false
	}
	return strings.Contains(out.Error(), substring)
}

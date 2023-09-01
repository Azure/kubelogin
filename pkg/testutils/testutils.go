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

func ErrorContains(out error, want string) bool {
	if out == nil {
		return want == ""
	}
	if want == "" {
		return false
	}
	return strings.Contains(out.Error(), want)
}

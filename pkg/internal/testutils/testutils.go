package testutils

import (
	"net/url"
	"strings"
)

const (
	ClientID       = "AZURE_CLIENT_ID"
	ClientSecret   = "AAD_SERVICE_PRINCIPAL_CLIENT_SECRET"
	ClientCert     = "AZURE_CLIENT_CER"
	ClientCertPass = "AZURE_CLIENT_CERTIFICATE_PASSWORD"
	ResourceID     = "AZURE_RESOURCE_ID"
	TenantID       = "AZURE_TENANT_ID"
	BadSecret      = "Bad_Secret"
	Username       = "USERNAME"
	Password       = "PASSWORD"
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

// ReplaceSecretValuesIncludingURLEscaped takes an input string, finds any instances of the
// input secret in the string (including in URL-escaped format), and replaces all instances
// with the given redaction token
// This is used for VCR tests as they sometimes include a URL-escaped version of the secret
// in the request body
func ReplaceSecretValuesIncludingURLEscaped(body, secret, redactionToken string) string {
	body = strings.ReplaceAll(body, secret, redactionToken)
	// get the URL-escaped version of the secret which replaces special characters with
	// the URL-safe "%AB" format
	escapedSecret := url.QueryEscape(secret)
	body = strings.ReplaceAll(body, escapedSecret, redactionToken)
	return body
}

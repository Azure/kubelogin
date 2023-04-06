package token

import (
	"strings"
	"testing"
)

func TestNewAzureCLITokenEmpty(t *testing.T) {
	_, err := newAzureCLIToken("", "")

	if !ErrorContains(err, "resourceID cannot be empty") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewAzureCLIToken(t *testing.T) {
	azcli := AzureCLIToken{}
	_, err := azcli.Token()

	if !ErrorContains(err, "expected an empty error but received:") {
		t.Errorf("unexpected error: %v", err)
	}
}

func ErrorContains(out error, want string) bool {
	if out == nil {
		return want == ""
	}
	if want == "" {
		return false
	}
	return strings.Contains(out.Error(), want)
}

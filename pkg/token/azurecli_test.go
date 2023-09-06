package token

import (
	"testing"

	"github.com/Azure/kubelogin/pkg/testutils"
)

func TestNewAzureCLITokenEmpty(t *testing.T) {
	_, err := newAzureCLIToken("", "")

	if !testutils.ErrorContains(err, "resourceID cannot be empty") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewAzureCLIToken(t *testing.T) {
	azcli := AzureCLIToken{}
	_, err := azcli.Token()

	if !testutils.ErrorContains(err, "expected an empty error but received:") {
		t.Errorf("unexpected error: %v", err)
	}
}

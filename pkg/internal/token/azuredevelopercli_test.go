package token

import (
	"context"
	"testing"

	"github.com/Azure/kubelogin/pkg/internal/testutils"
)

func TestNewAzureDeveloperCLITokenEmpty(t *testing.T) {
	// Using default timeout for testing
	_, err := newAzureDeveloperCLIToken("", "", defaultTimeout)

	if !testutils.ErrorContains(err, "resourceID cannot be empty") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewAzureDeveloperCLIToken(t *testing.T) {
	azd := AzureDeveloperCLIToken{}
	_, err := azd.Token(context.TODO())

	if !testutils.ErrorContains(err, "credential is nil") {
		t.Errorf("unexpected error: %v", err)
	}
}

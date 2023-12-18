package token

import (
	"context"
	"testing"

	"github.com/Azure/kubelogin/pkg/internal/testutils"
)

func TestNewAzureCLITokenEmpty(t *testing.T) {
	// Using default timeout for testing
	_, err := newAzureCLIToken("", "", defaultTimeout)

	if !testutils.ErrorContains(err, "resourceID cannot be empty") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewAzureCLIToken(t *testing.T) {
	azcli := AzureCLIToken{}
	_, err := azcli.Token(context.TODO())

	if !testutils.ErrorContains(err, "expected an empty error but received:") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOptionsValidateTimeout(t *testing.T) {
	// Assuming "devicecode" is a valid login method based on your error message
	validLoginMethod := "devicecode"

	// Test with Timeout set to 0
	options := Options{LoginMethod: validLoginMethod, Timeout: 0}
	err := options.Validate()
	if !testutils.ErrorContains(err, "timeout must be greater than 0") {
		t.Errorf("expected timeout error, got: %v", err)
	}

	// Test with Timeout set to a negative value
	options = Options{LoginMethod: validLoginMethod, Timeout: -1}
	err = options.Validate()
	if !testutils.ErrorContains(err, "timeout must be greater than 0") {
		t.Errorf("expected timeout error, got: %v", err)
	}

	// Test with a valid Timeout value
	options = Options{LoginMethod: validLoginMethod, Timeout: 10}
	err = options.Validate()
	if err != nil {
		t.Errorf("unexpected error for valid timeout: %v", err)
	}
}

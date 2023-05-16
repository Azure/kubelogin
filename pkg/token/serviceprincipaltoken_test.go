package token

import (
	"testing"
)

func TestNewServicePrincipalToken(t *testing.T) {
	servicePrincipalToken := servicePrincipalToken{}
	_, err := servicePrincipalToken.Token()

	if !ErrorContains(err, "expected an empty error but received:") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReadJWTFromSPEmptyString(t *testing.T) {
	_, err := readJWTFromFS("")
	if !ErrorContains(err, "no such file or directory") {
		t.Errorf("unexpected error: %v", err)
	}
}

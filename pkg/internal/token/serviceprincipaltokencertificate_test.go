package token

import (
	"context"
	"testing"

	"github.com/Azure/kubelogin/pkg/internal/testutils"
)

func TestMissingCertFile(t *testing.T) {
	p := &servicePrincipalToken{
		clientCert: "testdata/noCertHere.pfx",
	}
	expectedErr := "failed to read the certificate file"

	_, err := p.Token(context.TODO())
	if !testutils.ErrorContains(err, expectedErr) {
		t.Errorf("expected error %s, but got %s", expectedErr, err)
	}
}

func TestBadCertPassword(t *testing.T) {
	p := &servicePrincipalToken{
		clientCert:         "testdata/testCert.pfx",
		clientCertPassword: testutils.BadSecret,
	}
	expectedErr := "failed to decode pkcs12 certificate while creating spt: pkcs12: decryption password incorrect"

	_, err := p.Token(context.TODO())
	if !testutils.ErrorContains(err, expectedErr) {
		t.Errorf("expected error %s, but got %s", expectedErr, err)
	}
}

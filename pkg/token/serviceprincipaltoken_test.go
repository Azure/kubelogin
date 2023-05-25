package token

import (
	"os"
	"testing"
)

const (
	clientID       = "AZURE_CLIENT_ID"
	clientSecret   = "AAD_SERVICE_PRINCIPAL_CLIENT_SECRET"
	clientCert     = "AZURE_CLIENT_CER"
	clientCertPass = "AZURE_CLIENT_CERTIFICATE_PASSWORD"
	resourceID     = "AZURE_RESOURCE_ID"
	tenantID       = "AZURE_TENANT_ID"
)

func TestMissingLoginMethods(t *testing.T) {
	p := &servicePrincipalToken{
		clientID:   os.Getenv(clientID),
		resourceID: os.Getenv(resourceID),
		tenantID:   os.Getenv(tenantID),
	}
	expectedErr := "service principal token requires either client secret or certificate"

	_, err := p.Token()
	if !ErrorContains(err, expectedErr) {
		t.Errorf("expected error %v, but got %v", expectedErr, err)
	}
}

func TestBadSecret(t *testing.T) {
	p := &servicePrincipalToken{
		clientID:     os.Getenv(clientID),
		clientSecret: "Bad_Secret",
		resourceID:   os.Getenv(resourceID),
		tenantID:     os.Getenv(tenantID),
	}
	expectedErr := "ClientSecretCredential authentication failed"

	_, err := p.Token()
	if !ErrorContains(err, expectedErr) {
		t.Errorf("expected error %v, but got %v", expectedErr, err)
	}
}

func TestBadCertPassword(t *testing.T) {
	p := &servicePrincipalToken{
		clientID:           os.Getenv(clientID),
		clientCert:         os.Getenv(clientCert),
		clientCertPassword: "bad_password",
		resourceID:         os.Getenv(resourceID),
		tenantID:           os.Getenv(tenantID),
	}
	expectedErr := "failed to decode pkcs12 certificate while creating spt: pkcs12: decryption password incorrect"

	_, err := p.Token()
	if !ErrorContains(err, expectedErr) {
		t.Errorf("expected error %v, but got %v", expectedErr, err)
	}
}

func TestSecretToken(t *testing.T) {
	p := &servicePrincipalToken{
		clientID:     os.Getenv(clientID),
		clientSecret: os.Getenv(clientSecret),
		resourceID:   os.Getenv(resourceID),
		tenantID:     os.Getenv(tenantID),
	}

	_, err := p.Token()
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
}

func TestCertToken(t *testing.T) {
	p := &servicePrincipalToken{
		clientID:           os.Getenv(clientID),
		clientCert:         os.Getenv(clientCert),
		clientCertPassword: os.Getenv(clientCertPass),
		resourceID:         os.Getenv(resourceID),
		tenantID:           os.Getenv(tenantID),
	}

	_, err := p.Token()
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
}

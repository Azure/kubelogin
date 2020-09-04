package token

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/Azure/go-autorest/autorest/adal"
	"golang.org/x/crypto/pkcs12"
)

//pem block types
const (
	certificate = "CERTIFICATE"
	privateKey  = "PRIVATE KEY"
)

const (
	defaultEnvironment = "AzurePublicCloud"
)

type servicePrincipalToken struct {
	clientID     string
	clientSecret string
	clientCert   string
	resourceID   string
	tenantID     string
	oAuthConfig  adal.OAuthConfig
}

func newServicePrincipalToken(oAuthConfig adal.OAuthConfig, clientID, clientSecret, clientCert, resourceID, tenantID string) (TokenProvider, error) {
	if clientID == "" {
		return nil, errors.New("clientID cannot be empty")
	}
	if clientSecret == "" && clientCert == "" {
		return nil, errors.New("Both clientSecret and clientcert cannot be empty")
	}
	if clientSecret != "" && clientCert != "" {
		return nil, errors.New("Both clientSecret and clientcert cannot be set.Only one has to be specified")
	}
	if resourceID == "" {
		return nil, errors.New("resourceID cannot be empty")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}

	return &servicePrincipalToken{
		clientID:     clientID,
		clientSecret: clientSecret,
		clientCert:   clientCert,
		resourceID:   resourceID,
		tenantID:     tenantID,
		oAuthConfig:  oAuthConfig,
	}, nil
}

func (p *servicePrincipalToken) Token() (adal.Token, error) {
	emptyToken := adal.Token{}
	callback := func(t adal.Token) error {
		return nil
	}

	var (
		spt *adal.ServicePrincipalToken
		err error
	)

	if p.clientSecret != "" {
		spt, err = adal.NewServicePrincipalToken(
			p.oAuthConfig,
			p.clientID,
			p.clientSecret,
			p.resourceID,
			callback)
		if err != nil {
			return emptyToken, fmt.Errorf("failed to create service principal token using secret: %s", err)
		}
	} else if p.clientCert != "" {
		certData, err := ioutil.ReadFile(p.clientCert)
		if err != nil {
			return emptyToken, fmt.Errorf("failed to read the certificate file (%s): %v", p.clientCert, err)
		}

		// Get the certificate and private key from pfx file
		cert, rsaPrivateKey, err := decodePkcs12(certData, "")
		if err != nil {
			return emptyToken, fmt.Errorf("failed to decode pkcs12 certificate while creating spt: %v", err)
		}

		spt, err = adal.NewServicePrincipalTokenFromCertificate(
			p.oAuthConfig,
			p.clientID,
			cert,
			rsaPrivateKey,
			p.resourceID,
			callback)
		if err != nil {
			return emptyToken, fmt.Errorf("failed to create service principal token using cert: %s", err)
		}
	}

	err = spt.Refresh()
	if err != nil {
		return emptyToken, err
	}
	return spt.Token(), nil
}

func isPublicKeyEqual(key1, key2 *rsa.PublicKey) bool {
	if key1.N == nil || key2.N == nil {
		return false
	}
	return key1.E == key2.E && key1.N.Cmp(key2.N) == 0
}

func splitPEMBlock(pemBlock []byte) (certPEM []byte, keyPEM []byte) {
	for {
		var derBlock *pem.Block
		derBlock, pemBlock = pem.Decode(pemBlock)
		if derBlock == nil {
			break
		}
		if derBlock.Type == certificate {
			certPEM = append(certPEM, pem.EncodeToMemory(derBlock)...)
		} else if derBlock.Type == privateKey {
			keyPEM = append(keyPEM, pem.EncodeToMemory(derBlock)...)
		}
	}

	return certPEM, keyPEM
}

func parseRsaPrivateKey(privateKeyPEM []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("Failed to decode a pem block from private key")
	}

	privatePkcs1Key, errPkcs1 := x509.ParsePKCS1PrivateKey(block.Bytes)
	if errPkcs1 == nil {
		return privatePkcs1Key, nil
	}

	privatePkcs8Key, errPkcs8 := x509.ParsePKCS8PrivateKey(block.Bytes)
	if errPkcs8 == nil {
		privatePkcs8RsaKey, ok := privatePkcs8Key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("pkcs8 contained non-RSA key. Expected RSA key")
		}
		return privatePkcs8RsaKey, nil
	}

	return nil, fmt.Errorf("failed to parse private key as Pkcs#1 or Pkcs#8. (%s). (%s)", errPkcs1, errPkcs8)
}

func parseKeyPairFromPEMBlock(pemBlock []byte) (*x509.Certificate, *rsa.PrivateKey, error) {
	certPEM, keyPEM := splitPEMBlock(pemBlock)

	privateKey, err := parseRsaPrivateKey(keyPEM)
	if err != nil {
		return nil, nil, err
	}

	// Mooncake certificate signed by external provider is a certificate chain, we need to get the non-CA certificate,
	// i.e. the one with private key present.
	// Here is to locate the certificate only that matches the private key.
	found := false
	var cert *x509.Certificate
	for {
		var certBlock *pem.Block
		var err error
		certBlock, certPEM = pem.Decode(certPEM)
		if certBlock == nil {
			break
		}

		cert, err = x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to get client certificate PEM. %v", err)
		}

		certPublicKey, ok := cert.PublicKey.(*rsa.PublicKey)
		if ok {
			if isPublicKeyEqual(certPublicKey, &privateKey.PublicKey) {
				found = true
				break
			}
		}
	}

	if !found {
		return nil, nil, fmt.Errorf("Unable to find a matching public certificate")
	}

	return cert, privateKey, nil
}

func decodePkcs12(pkcs []byte, password string) (*x509.Certificate, *rsa.PrivateKey, error) {
	blocks, err := pkcs12.ToPEM(pkcs, password)
	if err != nil {
		return nil, nil, err
	}

	var (
		pemData []byte
	)

	for _, b := range blocks {
		pemData = append(pemData, pem.EncodeToMemory(b)...)
	}

	return parseKeyPairFromPEMBlock(pemData)
}

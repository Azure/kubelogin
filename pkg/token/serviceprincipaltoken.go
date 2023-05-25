package token

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/go-autorest/autorest/adal"
	"golang.org/x/crypto/pkcs12"
)

const (
	certificate = "CERTIFICATE"
	privateKey  = "PRIVATE KEY"
)

type servicePrincipalToken struct {
	clientID           string
	clientSecret       string
	clientCert         string
	clientCertPassword string
	resourceID         string
	tenantID           string
	oAuthConfig        adal.OAuthConfig
}

func newServicePrincipalToken(oAuthConfig adal.OAuthConfig, clientID, clientSecret, clientCert, clientCertPassword, resourceID, tenantID string) (TokenProvider, error) {
	if clientID == "" {
		return nil, errors.New("clientID cannot be empty")
	}
	if clientSecret == "" && clientCert == "" {
		return nil, errors.New("both clientSecret and clientcert cannot be empty")
	}
	if clientSecret != "" && clientCert != "" {
		return nil, errors.New("client secret and client certificate cannot be set at the same time. Only one has to be specified")
	}
	if resourceID == "" {
		return nil, errors.New("resourceID cannot be empty")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}

	return &servicePrincipalToken{
		clientID:           clientID,
		clientSecret:       clientSecret,
		clientCert:         clientCert,
		clientCertPassword: clientCertPassword,
		resourceID:         resourceID,
		tenantID:           tenantID,
		oAuthConfig:        oAuthConfig,
	}, nil
}

// Token fetches an azcore.AccessToken from the Azure SDK and converts it to an adal.Token for use with kubelogin.
func (p *servicePrincipalToken) Token() (adal.Token, error) {
	emptyToken := adal.Token{}
	var spnAccessToken azcore.AccessToken

	// Request a new Azure token provider for secret or certificate
	if p.clientSecret != "" {
		cred, err := azidentity.NewClientSecretCredential(
			p.tenantID,
			p.clientID,
			p.clientSecret,
			nil,
		)
		if err != nil {
			return emptyToken, fmt.Errorf("unable to create credential. Received: %v", err)
		}

		// Use the token provider to get a new token
		spnAccessToken, err = cred.GetToken(context.Background(), policy.TokenRequestOptions{Scopes: []string{p.resourceID + "/.default"}})
		if err != nil {
			return emptyToken, fmt.Errorf("expected an empty error but received: %v", err)
		}

	} else if p.clientCert != "" {
		certData, err := os.ReadFile(p.clientCert)
		if err != nil {
			return emptyToken, fmt.Errorf("failed to read the certificate file (%s): %w", p.clientCert, err)
		}

		// Get the certificate and private key from pfx file
		cert, rsaPrivateKey, err := decodePkcs12(certData, p.clientCertPassword)
		if err != nil {
			return emptyToken, fmt.Errorf("failed to decode pkcs12 certificate while creating spt: %w", err)
		}

		cred, err := azidentity.NewClientCertificateCredential(
			p.tenantID,
			p.clientID,
			[]*x509.Certificate{cert},
			rsaPrivateKey,
			nil,
		)
		if err != nil {
			return emptyToken, fmt.Errorf("unable to create credential. Received: %v", err)
		}
		spnAccessToken, err = cred.GetToken(context.Background(), policy.TokenRequestOptions{Scopes: []string{p.resourceID + "/.default"}})
		if err != nil {
			return emptyToken, fmt.Errorf("expected an empty error but received: %v", err)
		}
	} else {
		return emptyToken, errors.New("service principal token requires either client secret or certificate")
	}

	if spnAccessToken.Token == "" {
		return emptyToken, errors.New("did not receive a token")
	}

	// azurecore.AccessTokens have ExpiresOn as Time.Time. We need to convert it to JSON.Number
	// by fetching the time in seconds since the Unix epoch via Unix() and then converting to a
	// JSON.Number via formatting as a string using a base-10 int64 conversion.
	expiresOn := json.Number(strconv.FormatInt(spnAccessToken.ExpiresOn.Unix(), 10))

	// Re-wrap the azurecore.AccessToken into an adal.Token
	return adal.Token{
		AccessToken: spnAccessToken.Token,
		ExpiresOn:   expiresOn,
		Resource:    p.resourceID,
	}, nil
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
		return nil, fmt.Errorf("failed to decode a pem block from private key")
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
			return nil, nil, fmt.Errorf("unable to parse certificate. %w", err)
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
		return nil, nil, fmt.Errorf("unable to find a matching public certificate")
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

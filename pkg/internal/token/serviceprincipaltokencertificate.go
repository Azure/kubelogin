package token

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/kubelogin/pkg/internal/pop"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"golang.org/x/crypto/pkcs12"
)

// getTokenWithClientCert requests a token using the configured client ID/certificate
// and returns a PoP token if PoP claims are provided, otherwise returns a regular
// bearer token
func (p *servicePrincipalToken) getTokenWithClientCert(
	context context.Context,
	scopes []string,
	options *azcore.ClientOptions,
) (string, int64, error) {
	clientOptions := &azidentity.ClientCertificateCredentialOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: p.cloud,
		},
		SendCertificateChain: true,
	}
	if options != nil {
		clientOptions.ClientOptions = *options
	}
	certData, err := os.ReadFile(p.clientCert)
	if err != nil {
		return "", -1, fmt.Errorf("failed to read the certificate file (%s): %w", p.clientCert, err)
	}

	// Get the certificate and private key from pfx file
	cert, rsaPrivateKey, err := decodePkcs12(certData, p.clientCertPassword)
	if err != nil {
		return "", -1, fmt.Errorf("failed to decode pkcs12 certificate while creating spt: %w", err)
	}

	certArray := []*x509.Certificate{cert}
	if len(p.popClaims) > 0 {
		// if PoP token support is enabled, use the PoP token flow to request the token
		return p.getPoPTokenWithClientCert(context, scopes, certArray, rsaPrivateKey, options)
	}

	cred, err := azidentity.NewClientCertificateCredential(
		p.tenantID,
		p.clientID,
		certArray,
		rsaPrivateKey,
		clientOptions,
	)
	if err != nil {
		return "", -1, fmt.Errorf("unable to create credential. Received: %v", err)
	}
	spnAccessToken, err := cred.GetToken(context, policy.TokenRequestOptions{Scopes: scopes})
	if err != nil {
		return "", -1, fmt.Errorf("failed to create service principal token using cert: %s", err)
	}

	return spnAccessToken.Token, spnAccessToken.ExpiresOn.Unix(), nil
}

// getPoPTokenWithClientCert requests a PoP token using the given client ID/certificate
// and returns it
func (p *servicePrincipalToken) getPoPTokenWithClientCert(
	context context.Context,
	scopes []string,
	certArray []*x509.Certificate,
	rsaPrivateKey *rsa.PrivateKey,
	options *azcore.ClientOptions,
) (string, int64, error) {
	cred, err := confidential.NewCredFromCert(certArray, rsaPrivateKey)
	if err != nil {
		return "", -1, fmt.Errorf("unable to create credential from certificate. Received: %w", err)
	}

	accessToken, expiresOn, err := pop.AcquirePoPTokenConfidential(
		context,
		p.popClaims,
		scopes,
		cred,
		p.cloud.ActiveDirectoryAuthorityHost,
		p.clientID,
		p.tenantID,
		true,
		options,
		pop.GetSwPoPKey,
	)
	if err != nil {
		return "", -1, fmt.Errorf("failed to create service principal PoP token using certificate: %w", err)
	}

	return accessToken, expiresOn, nil
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

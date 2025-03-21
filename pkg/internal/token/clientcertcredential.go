package token

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"
	"golang.org/x/crypto/pkcs12"
	"k8s.io/klog/v2"
)

type ClientCertificateCredential struct {
	cred *azidentity.ClientCertificateCredential
}

var _ CredentialProvider = (*ClientCertificateCredential)(nil)

func newClientCertificateCredential(opts *Options) (CredentialProvider, error) {
	if opts.ClientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}
	if opts.TenantID == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	if opts.ClientCert == "" {
		return nil, fmt.Errorf("client certificate cannot be empty")
	}
	var (
		c   azidentity.Cache
		err error
	)
	if opts.UsePersistentCache {
		c, err = cache.New(nil)
		if err != nil {
			klog.V(5).Infof("failed to create cache: %v", err)
		}
	}

	// Get the certificate and private key from file
	cert, rsaPrivateKey, err := readCertificate(opts.ClientCert, opts.ClientCertPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	azOpts := &azidentity.ClientCertificateCredentialOptions{
		ClientOptions:            azcore.ClientOptions{Cloud: opts.GetCloudConfiguration()},
		Cache:                    c,
		SendCertificateChain:     true,
		DisableInstanceDiscovery: opts.DisableInstanceDiscovery,
	}

	if opts.httpClient != nil {
		azOpts.ClientOptions.Transport = opts.httpClient
	}

	cred, err := azidentity.NewClientCertificateCredential(
		opts.TenantID, opts.ClientID,
		[]*x509.Certificate{cert}, rsaPrivateKey,
		azOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create client certificate credential: %w", err)
	}
	return &ClientCertificateCredential{cred: cred}, nil
}

func (c *ClientCertificateCredential) Name() string {
	return "ClientCertificateCredential"
}

func (c *ClientCertificateCredential) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	return azidentity.AuthenticationRecord{}, errAuthenticateNotSupported
}

func (c *ClientCertificateCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return c.cred.GetToken(ctx, opts)
}

func (c *ClientCertificateCredential) NeedAuthenticate() bool {
	return false
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
		if derBlock.Type == "CERTIFICATE" {
			certPEM = append(certPEM, pem.EncodeToMemory(derBlock)...)
		} else if derBlock.Type == "PRIVATE KEY" {
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

	return nil, fmt.Errorf("failed to parse private key as Pkcs#1 or Pkcs#8. (%w), (%w)", errPkcs1, errPkcs8)
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
			return nil, nil, fmt.Errorf("unable to parse certificate: %w", err)
		}

		certPublicKey, ok := cert.PublicKey.(*rsa.PublicKey)
		if ok && isPublicKeyEqual(certPublicKey, &privateKey.PublicKey) {
			found = true
			break
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

	var pemData []byte
	for _, b := range blocks {
		pemData = append(pemData, pem.EncodeToMemory(b)...)
	}

	return parseKeyPairFromPEMBlock(pemData)
}

func readCertificate(certFile, password string) (*x509.Certificate, *rsa.PrivateKey, error) {
	if strings.HasSuffix(certFile, ".pfx") {
		cert, err := os.ReadFile(certFile)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read the certificate file (%s): %w", certFile, err)
		}
		return decodePkcs12(cert, password)
	} else {
		cert, err := os.ReadFile(certFile)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read the certificate file (%s): %w", certFile, err)
		}
		return parseKeyPairFromPEMBlock(cert)
	}
}

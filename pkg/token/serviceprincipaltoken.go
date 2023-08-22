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
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/kubelogin/pkg/pop"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
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
	cloud              cloud.Configuration
	popClaims          map[string]string
}

func newServicePrincipalToken(cloud cloud.Configuration, clientID, clientSecret, clientCert, clientCertPassword, resourceID, tenantID string, popClaims map[string]string) (TokenProvider, error) {
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
		cloud:              cloud,
		popClaims:          popClaims,
	}, nil
}

// Token fetches an azcore.AccessToken from the Azure SDK and converts it to an adal.Token for use with kubelogin.
func (p *servicePrincipalToken) Token() (adal.Token, error) {
	return p.TokenWithOptions(nil)
}

func (p *servicePrincipalToken) TokenWithOptions(options *azcore.ClientOptions) (adal.Token, error) {
	emptyToken := adal.Token{}
	var accessToken string
	var expirationTimeUnix int64
	var err error
	scopes := []string{p.resourceID + "/.default"}

	// Request a new Azure token provider for service principal
	if p.clientSecret != "" {
		accessToken, expirationTimeUnix, err = p.getTokenWithClientSecret(options, scopes)
		if err != nil {
			return emptyToken, fmt.Errorf("failed to create service principal token using secret: %w", err)
		}
	} else if p.clientCert != "" {
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
			clientOptions,
		)
		if err != nil {
			return emptyToken, fmt.Errorf("unable to create credential. Received: %v", err)
		}
		spnAccessToken, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{Scopes: []string{p.resourceID + "/.default"}})
		if err != nil {
			return emptyToken, fmt.Errorf("failed to create service principal token using cert: %s", err)
		}

		accessToken = spnAccessToken.Token
		expirationTimeUnix = spnAccessToken.ExpiresOn.Unix()
	} else {
		return emptyToken, errors.New("service principal token requires either client secret or certificate")
	}

	if accessToken == "" {
		return emptyToken, errors.New("unexpectedly got empty access token")
	}

	// azurecore.AccessTokens have ExpiresOn as Time.Time. We need to convert it to JSON.Number
	// by fetching the time in seconds since the Unix epoch via Unix() and then converting to a
	// JSON.Number via formatting as a string using a base-10 int64 conversion.
	expiresOn := json.Number(strconv.FormatInt(expirationTimeUnix, 10))

	// Re-wrap the azurecore.AccessToken into an adal.Token
	return adal.Token{
		AccessToken: accessToken,
		ExpiresOn:   expiresOn,
		Resource:    p.resourceID,
	}, nil
}

func (p *servicePrincipalToken) getTokenWithClientSecret(options *azcore.ClientOptions, scopes []string) (string, int64, error) {
	if p.popClaims != nil && len(p.popClaims) > 0 {
		// if PoP token support is enabled, use the PoP token flow to request the token
		return p.getPoPTokenWithClientSecret(scopes)
	}

	clientOptions := &azidentity.ClientSecretCredentialOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: p.cloud,
		},
	}
	if options != nil {
		clientOptions.ClientOptions = *options
	}
	cred, err := azidentity.NewClientSecretCredential(
		p.tenantID,
		p.clientID,
		p.clientSecret,
		clientOptions,
	)
	if err != nil {
		return "", -1, fmt.Errorf("unable to create credential. Received: %w", err)
	}

	// Use the token provider to get a new token
	spnAccessToken, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{Scopes: scopes})
	if err != nil {
		return "", -1, fmt.Errorf("failed to create service principal bearer token using secret: %w", err)
	}

	return spnAccessToken.Token, spnAccessToken.ExpiresOn.Unix(), nil
}

func (p *servicePrincipalToken) getPoPTokenWithClientSecret(scopes []string) (string, int64, error) {
	cred, err := confidential.NewCredFromSecret(p.clientSecret)
	if err != nil {
		return "", -1, fmt.Errorf("unable to create credential. Received: %w", err)
	}

	client, err := confidential.New(
		p.cloud.ActiveDirectoryAuthorityHost,
		p.clientID,
		cred,
	)
	if err != nil {
		return "", -1, fmt.Errorf("unable to create client. Received: %w", err)
	}

	result, err := client.AcquireTokenSilent(
		context.Background(),
		scopes,
		confidential.WithAuthenticationScheme(
			&pop.PopAuthenticationScheme{
				Host:   p.popClaims["u"],
				PoPKey: pop.GetSwPoPKey(),
			},
		),
		confidential.WithTenantID(p.tenantID),
	)
	if err != nil {
		result, err = client.AcquireTokenByCredential(
			context.Background(),
			scopes,
			confidential.WithAuthenticationScheme(
				&pop.PopAuthenticationScheme{
					Host:   p.popClaims["u"],
					PoPKey: pop.GetSwPoPKey(),
				},
			),
			confidential.WithTenantID(p.tenantID),
		)
		if err != nil {
			return "", -1, fmt.Errorf("failed to create service principal PoP token using secret: %w", err)
		}
	}

	return result.AccessToken, result.ExpiresOn.Unix(), nil
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

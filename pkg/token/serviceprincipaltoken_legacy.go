package token

import (
	"errors"
	"fmt"
	"os"

	"github.com/Azure/go-autorest/autorest/adal"
)

type legacyServicePrincipalToken struct {
	clientID           string
	clientSecret       string
	clientCert         string
	clientCertPassword string
	resourceID         string
	tenantID           string
	oAuthConfig        adal.OAuthConfig
}

func newLegacyServicePrincipalToken(oAuthConfig adal.OAuthConfig, clientID, clientSecret, clientCert, clientCertPassword, resourceID, tenantID string, isLegacy bool) (TokenProvider, error) {
	if !isLegacy {
		return nil, errors.New("legacy service principal token requires isLegacy being true")
	}
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

	return &legacyServicePrincipalToken{
		clientID:           clientID,
		clientSecret:       clientSecret,
		clientCert:         clientCert,
		clientCertPassword: clientCertPassword,
		resourceID:         resourceID,
		tenantID:           tenantID,
		oAuthConfig:        oAuthConfig,
	}, nil
}

func (p *legacyServicePrincipalToken) Token() (adal.Token, error) {
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
		certData, err := os.ReadFile(p.clientCert)
		if err != nil {
			return emptyToken, fmt.Errorf("failed to read the certificate file (%s): %w", p.clientCert, err)
		}

		// Get the certificate and private key from pfx file
		cert, rsaPrivateKey, err := decodePkcs12(certData, p.clientCertPassword)
		if err != nil {
			return emptyToken, fmt.Errorf("failed to decode pkcs12 certificate while creating spt: %w", err)
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

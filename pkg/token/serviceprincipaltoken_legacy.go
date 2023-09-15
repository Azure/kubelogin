package token

import (
	"errors"
	"fmt"
	"os"

	"github.com/Azure/go-autorest/autorest/adal"
)

var errInvalidOAuthConfig = errors.New("OAuthConfig needs to be configured with api-version=1.0")

type legacyServicePrincipalToken struct {
	clientID           string
	clientSecret       string
	clientCert         string
	clientCertPassword string
	resourceID         string
	tenantID           string
	oAuthConfig        adal.OAuthConfig
}

func newLegacyServicePrincipalToken(oAuthConfig adal.OAuthConfig, clientID, clientSecret, clientCert, clientCertPassword, resourceID, tenantID string) (TokenProvider, error) {
	if err := validateOAuthConfig(oAuthConfig); err != nil {
		return nil, err
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

	var (
		spt *adal.ServicePrincipalToken
		err error
	)

	if p.clientSecret != "" {
		spt, err = adal.NewServicePrincipalToken(
			p.oAuthConfig,
			p.clientID,
			p.clientSecret,
			p.resourceID)
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
			p.resourceID)
		if err != nil {
			return emptyToken, fmt.Errorf("failed to create service principal token using cert: %s", err)
		}
	}

	err = spt.EnsureFresh()
	if err != nil {
		return emptyToken, err
	}
	return spt.Token(), nil
}

func validateOAuthConfig(config adal.OAuthConfig) error {
	v := config.AuthorizeEndpoint.Query().Get("api-version")
	if v != "1.0" {
		return errInvalidOAuthConfig
	}

	v = config.TokenEndpoint.Query().Get("api-version")
	if v != "1.0" {
		return errInvalidOAuthConfig
	}

	return nil
}

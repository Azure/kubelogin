// Disclaimer: The PoPAuthenticationScheme implementation of the MSAL AuthenticationScheme
// interface is intended for the usage of Azure Arc.

package pop

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// type of a PoP token, as opposed to "JWT" for a regular bearer token
const popTokenType = "pop"

// PoPAuthenticationScheme is a PoP token implementation of the MSAL AuthenticationScheme interface
// used by the Azure Arc Platform team.
// This implementation will only use the passed-in u-claim (representing the ARM ID of the
// cluster/host); other claims passed in during a PoP token request will be disregarded
type PoPAuthenticationScheme struct {
	// host is the u claim we will add on the pop token
	Host   string
	PoPKey PoPKey
}

// TokenRequestParams returns the params to use when sending a request for a PoP token
func (as *PoPAuthenticationScheme) TokenRequestParams() map[string]string {
	return map[string]string{
		"token_type": popTokenType,
		"req_cnf":    as.PoPKey.ReqCnf(),
	}
}

// KeyID returns the key used to sign the PoP token
func (as *PoPAuthenticationScheme) KeyID() string {
	return as.PoPKey.KeyID()
}

// FormatAccessToken takes an access token, formats it as a PoP token,
// and returns it as a base-64 encoded string
func (as *PoPAuthenticationScheme) FormatAccessToken(accessToken string) (string, error) {
	timestamp := time.Now().Unix()
	nonce := uuid.NewString()
	nonce = strings.ReplaceAll(nonce, "-", "")

	return as.FormatAccessTokenWithOptions(accessToken, nonce, timestamp)
}

// FormatAccessTokenWithOptions takes an access token, nonce, and timestamp, formats
// the token as a PoP token containing the given fields, and returns it as a
// base-64 encoded string
func (as *PoPAuthenticationScheme) FormatAccessTokenWithOptions(accessToken, nonce string, timestamp int64) (string, error) {
	header := header{
		typ: popTokenType,
		alg: as.PoPKey.Alg(),
		kid: as.PoPKey.KeyID(),
	}
	payload := payload{
		at:    accessToken,
		ts:    timestamp,
		host:  as.Host,
		jwk:   as.PoPKey.JWK(),
		nonce: nonce,
	}

	popAccessToken, err := createPoPAccessToken(header, payload, as.PoPKey)
	if err != nil {
		return "", fmt.Errorf("error formatting PoP token: %w", err)
	}
	return popAccessToken.ToBase64(), nil
}

// AccessTokenType returns the PoP access token type
func (as *PoPAuthenticationScheme) AccessTokenType() string {
	return popTokenType
}

// type representing the header of a PoP access token
type header struct {
	typ string
	alg string
	kid string
}

// ToString returns a string representation of a header object
func (h *header) ToString() string {
	return fmt.Sprintf(`{"typ":"%s","alg":"%s","kid":"%s"}`, h.typ, h.alg, h.kid)
}

// ToBase64 returns a base-64 encoded string representation of a header object
func (h *header) ToBase64() string {
	return base64.RawURLEncoding.EncodeToString([]byte(h.ToString()))
}

// type representing the payload of a PoP token
type payload struct {
	at    string
	ts    int64
	host  string
	jwk   string
	nonce string
}

// ToString returns a string representation of a payload object
func (p *payload) ToString() string {
	return fmt.Sprintf(`{"at":"%s","ts":%d,"u":"%s","cnf":{"jwk":%s},"nonce":"%s"}`, p.at, p.ts, p.host, p.jwk, p.nonce)
}

// ToBase64 returns a base-64 encoded representation of a payload object
func (p *payload) ToBase64() string {
	return base64.RawURLEncoding.EncodeToString([]byte(p.ToString()))
}

// type representing the signature of a PoP token
type signature struct {
	sig []byte
}

// ToBase64 returns a base-64 encoded representation of a signature object
func (s *signature) ToBase64() string {
	return base64.RawURLEncoding.EncodeToString(s.sig)
}

// type representing a PoP access token
type popAccessToken struct {
	Header    header
	Payload   payload
	Signature signature
}

// given a header, payload, and PoP key, creates the signature for the token and returns
// a PoPAccessToken object representing the signed token
func createPoPAccessToken(h header, p payload, popKey PoPKey) (*popAccessToken, error) {
	token := &popAccessToken{
		Header:  h,
		Payload: p,
	}
	h256 := sha256.Sum256([]byte(h.ToBase64() + "." + p.ToBase64()))
	sig, err := popKey.Sign(h256[:])
	if err != nil {
		return nil, err
	}
	token.Signature = signature{
		sig: sig,
	}
	return token, nil
}

// ToBase64 returns a base-64 encoded representation of a PoP access token
func (p *popAccessToken) ToBase64() string {
	return fmt.Sprintf("%s.%s.%s", p.Header.ToBase64(), p.Payload.ToBase64(), p.Signature.ToBase64())
}

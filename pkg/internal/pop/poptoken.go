package pop

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
)

// PoPKey is a generic interface for PoP key properties and methods
type PoPKey interface {
	// encryption/signature algo
	Alg() string
	// kid
	KeyID() string
	// jwk that can be embedded in JWT w/ PoP token's cnf claim
	JWK() string
	// https://tools.ietf.org/html/rfc7638 compliant jwk thumbprint
	JWKThumbprint() string
	// req_cnf claim that can be included in access token request to AAD
	ReqCnf() string
	// sign payload using private key
	Sign([]byte) ([]byte, error)
}

// software based pop key implementation of PoPKey
type swKey struct {
	key    *rsa.PrivateKey
	keyID  string
	jwk    string
	jwkTP  string
	reqCnf string
}

// Alg returns the algorithm used to encrypt/sign the swKey
func (swk *swKey) Alg() string {
	return "RS256"
}

// KeyID returns the keyID of the swKey, representing the key used to sign the swKey
func (swk *swKey) KeyID() string {
	return swk.keyID
}

// JWK returns the JSON Web Key of the given swKey
func (swk *swKey) JWK() string {
	return swk.jwk
}

// JWKThumbprint returns the JWK thumbprint of the given swKey
func (swk *swKey) JWKThumbprint() string {
	return swk.jwkTP
}

// ReqCnf returns the req_cnf claim to send to AAD for the given swKey
func (swk *swKey) ReqCnf() string {
	return swk.reqCnf
}

// Sign uses the given swKey to sign the given payload and returns the signed payload
func (swk *swKey) Sign(payload []byte) ([]byte, error) {
	return swk.key.Sign(rand.Reader, payload, crypto.SHA256)
}

// init initializes the given swKey using the given private key
func (swk *swKey) init(key *rsa.PrivateKey) {
	swk.key = key

	eB64, nB64 := getRSAKeyExponentAndModulus(key)
	swk.jwkTP = computeJWKThumbprint(eB64, nB64)
	swk.reqCnf = getReqCnf(swk.jwkTP)

	// set keyID to jwkTP
	swk.keyID = swk.jwkTP

	// compute JWK to be included in JWT w/ PoP token's cnf claim
	// - https://tools.ietf.org/html/rfc7800#section-3.2
	swk.jwk = getJWK(eB64, nB64, swk.keyID)
}

// generateSwKey generates a new swkey and initializes it with required fields before returning it
func generateSwKey(key *rsa.PrivateKey) (*swKey, error) {
	swk := &swKey{}
	swk.init(key)
	return swk, nil
}

// GetSwPoPKey generates a new PoP key that rotates every 8 hours and returns it
func GetSwPoPKey() (*swKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("error generating RSA private key: %w", err)
	}
	return GetSwPoPKeyWithRSAKey(key)
}

func GetSwPoPKeyWithRSAKey(rsaKey *rsa.PrivateKey) (*swKey, error) {
	key, err := generateSwKey(rsaKey)
	if err != nil {
		return nil, fmt.Errorf("unable to generate PoP key. err: %w", err)
	}
	return key, nil
}

// getRSAKeyExponentAndModulus returns the exponent and modulus from the given RSA key
// as base-64 encoded strings
func getRSAKeyExponentAndModulus(rsaKey *rsa.PrivateKey) (string, string) {
	pubKey := rsaKey.PublicKey
	e := big.NewInt(int64(pubKey.E))
	eB64 := base64.RawURLEncoding.EncodeToString(e.Bytes())
	n := pubKey.N
	nB64 := base64.RawURLEncoding.EncodeToString(n.Bytes())
	return eB64, nB64
}

// computeJWKThumbprint returns a computed JWK thumbprint using the given base-64 encoded
// exponent and modulus
func computeJWKThumbprint(eB64 string, nB64 string) string {
	// compute JWK thumbprint
	// jwk format - e, kty, n - in lexicographic order
	// - https://tools.ietf.org/html/rfc7638#section-3.3
	// - https://tools.ietf.org/html/rfc7638#section-3.1
	jwk := fmt.Sprintf(`{"e":"%s","kty":"RSA","n":"%s"}`, eB64, nB64)
	jwkS256 := sha256.Sum256([]byte(jwk))
	return base64.RawURLEncoding.EncodeToString(jwkS256[:])
}

// getReqCnf computes and returns the value for the req_cnf claim to include when sending
// a request for the token
func getReqCnf(jwkTP string) string {
	// req_cnf - base64URL("{"kid":"jwkTP","xms_ksl":"sw"}")
	reqCnfJSON := fmt.Sprintf(`{"kid":"%s","xms_ksl":"sw"}`, jwkTP)
	return base64.RawURLEncoding.EncodeToString([]byte(reqCnfJSON))
}

// getJWK computes the JWK to be included in the PoP token's enclosed cnf claim and returns it
func getJWK(eB64 string, nB64 string, keyID string) string {
	// compute JWK to be included in JWT w/ PoP token's cnf claim
	// - https://tools.ietf.org/html/rfc7800#section-3.2
	return fmt.Sprintf(`{"e":"%s","kty":"RSA","n":"%s","alg":"RS256","kid":"%s"}`, eB64, nB64, keyID)
}

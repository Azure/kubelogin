// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pop

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// PoPKey - generic interface for PoP key properties and methods
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

// returns the algorithm used to encrypt/sign the swKey
func (swk *swKey) Alg() string {
	return "RS256"
}

// returns the keyID of the swKey, representing the key used to sign the swKey
func (swk *swKey) KeyID() string {
	return swk.keyID
}

// returns the JSON Web Key of the given swKey
func (swk *swKey) JWK() string {
	return swk.jwk
}

// returns the JWK thumbprint of the given swKey
func (swk *swKey) JWKThumbprint() string {
	return swk.jwkTP
}

// returns the req_cnf claim to send to AAD for the given swKey
func (swk *swKey) ReqCnf() string {
	return swk.reqCnf
}

// uses the given swKey to sign the given payload and returns the signed payload
func (swk *swKey) Sign(payload []byte) ([]byte, error) {
	return swk.key.Sign(rand.Reader, payload, crypto.SHA256)
}

// initializes the given swKey using the given private key
func (swk *swKey) init(key *rsa.PrivateKey) {
	swk.key = key

	pubKey := &swk.key.PublicKey
	e := big.NewInt(int64(pubKey.E))
	eB64 := base64.RawURLEncoding.EncodeToString(e.Bytes())
	n := pubKey.N
	nB64 := base64.RawURLEncoding.EncodeToString(n.Bytes())

	// compute JWK thumbprint
	// jwk format - e, kty, n - in lexicographic order
	// - https://tools.ietf.org/html/rfc7638#section-3.3
	// - https://tools.ietf.org/html/rfc7638#section-3.1
	jwk := fmt.Sprintf(`{"e":"%s","kty":"RSA","n":"%s"}`, eB64, nB64)
	jwkS256 := sha256.Sum256([]byte(jwk))
	swk.jwkTP = base64.RawURLEncoding.EncodeToString(jwkS256[:])

	// req_cnf - base64URL("{"kid":"jwkTP","xms_ksl":"sw"}")
	reqCnfJSON := fmt.Sprintf(`{"kid":"%s","xms_ksl":"sw"}`, swk.jwkTP)
	swk.reqCnf = base64.RawURLEncoding.EncodeToString([]byte(reqCnfJSON))

	// set keyID to jwkTP
	swk.keyID = swk.jwkTP

	// compute JWK to be included in JWT w/ PoP token's cnf claim
	// - https://tools.ietf.org/html/rfc7800#section-3.2
	swk.jwk = fmt.Sprintf(`{"e":"%s","kty":"RSA","n":"%s","alg":"RS256","kid":"%s"}`, eB64, nB64, swk.keyID)
}

// generates a new swkey and initializes it with required fields before returning it
func generateSwKey() (*swKey, error) {
	swk := &swKey{}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	swk.init(key)
	return swk, nil
}

var pswKey *swKey
var pwsKeyMutex sync.Mutex

// generates a new PoP key that rotates every 8 hours and returns it
func GetSwPoPKey() (*swKey, error) {
	pwsKeyMutex.Lock()
	defer pwsKeyMutex.Unlock()
	if pswKey != nil {
		return pswKey, nil
	}

	key, err := generateSwKey()
	if err != nil {
		return nil, fmt.Errorf("unable to generate popkey. err: %w", err)
	}
	pswKey = key

	// rotate key every 8 hours
	ticker := time.NewTicker(8 * time.Hour)
	go func() error {
		for {
			<-ticker.C
			key, err := generateSwKey()
			if err != nil {
				return fmt.Errorf("unable to generate popkey. err: %w", err)
			}
			pwsKeyMutex.Lock()
			pswKey = key
			pwsKeyMutex.Unlock()
		}
	}()
	if err != nil {
		return nil, err
	}

	return pswKey, nil
}

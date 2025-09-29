package pop

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
)

const popKeyFileName = "pop_key.pem"

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
type SwKey struct {
	key    *rsa.PrivateKey
	keyID  string
	jwk    string
	jwkTP  string
	reqCnf string
}

// Alg returns the algorithm used to encrypt/sign the SwKey
func (swk *SwKey) Alg() string {
	return "RS256"
}

// KeyID returns the keyID of the SwKey, representing the key used to sign the SwKey
func (swk *SwKey) KeyID() string {
	return swk.keyID
}

// JWK returns the JSON Web Key of the given SwKey
func (swk *SwKey) JWK() string {
	return swk.jwk
}

// JWKThumbprint returns the JWK thumbprint of the given SwKey
func (swk *SwKey) JWKThumbprint() string {
	return swk.jwkTP
}

// ReqCnf returns the req_cnf claim to send to AAD for the given SwKey
func (swk *SwKey) ReqCnf() string {
	return swk.reqCnf
}

// Sign uses the given SwKey to sign the given payload and returns the signed payload
func (swk *SwKey) Sign(payload []byte) ([]byte, error) {
	return swk.key.Sign(rand.Reader, payload, crypto.SHA256)
}

// init initializes the given SwKey using the given private key
func (swk *SwKey) init(key *rsa.PrivateKey) {
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

// generateSwKey generates a new SwKey and initializes it with required fields before returning it
func generateSwKey(key *rsa.PrivateKey) (*SwKey, error) {
	swk := &SwKey{}
	swk.init(key)
	return swk, nil
}

// GetSwPoPKey generates a new PoP key returns it
// Deprecated: This function generates a new key each time, breaking PoP token caching.
// Use GetSwPoPKeyPersistent() instead for proper caching support.
func GetSwPoPKey() (*SwKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("error generating RSA private key: %w", err)
	}
	return GetSwPoPKeyWithRSAKey(key)
}

// GetSwPoPKeyPersistent loads or generates a persistent PoP key for token caching.
// This ensures the same PoP key is used across multiple kubelogin invocations,
// which is required for PoP token caching to work correctly.
func GetSwPoPKeyPersistent(cacheDir string) (*SwKey, error) {
	key, err := loadOrGenerateRSAKey(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("error loading or generating persistent RSA private key: %w", err)
	}
	return GetSwPoPKeyWithRSAKey(key)
}

func GetSwPoPKeyWithRSAKey(rsaKey *rsa.PrivateKey) (*SwKey, error) {
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

// getPoPKeyFilePath returns the file path for the persistent PoP RSA key.
func getPoPKeyFilePath(cacheDir string) string {
	return filepath.Join(cacheDir, popKeyFileName)
}

// loadOrGenerateRSAKey loads an existing RSA key from disk or generates a new one if it doesn't exist.
// This ensures the same PoP key is used across multiple kubelogin invocations, which is required
// for PoP token caching to work correctly.
func loadOrGenerateRSAKey(cacheDir string) (*rsa.PrivateKey, error) {
	keyPath := getPoPKeyFilePath(cacheDir)

	// Try to load existing key first
	if key, err := loadRSAKey(keyPath); err == nil {
		return key, nil
	}

	// Generate new key if loading failed
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("error generating RSA private key: %w", err)
	}

	// Save the key for future use
	if err := saveRSAKey(keyPath, key); err != nil {
		// Log warning but don't fail - key generation succeeded
		// This allows the PoP token to work even if key persistence fails
		fmt.Fprintf(os.Stderr, "Warning: failed to persist PoP key to %s: %v\n", keyPath, err)
	}

	return key, nil
}

// loadRSAKey loads an RSA private key from a PEM file.
func loadRSAKey(keyPath string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("invalid PEM block type")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
	}

	return key, nil
}

// saveRSAKey saves an RSA private key to a PEM file with secure permissions.
func saveRSAKey(keyPath string, key *rsa.PrivateKey) error {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(keyPath), 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// Convert key to PEM format
	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	})

	// Write with secure permissions (readable only by owner)
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return nil
}

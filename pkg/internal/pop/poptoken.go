package pop

import (
	"context"
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

	"github.com/Azure/kubelogin/pkg/internal/pop/cache"
)

const popKeyFileName = "pop_rsa_key.cache"

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
func GetSwPoPKey() (*SwKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("error generating RSA private key: %w", err)
	}
	return GetSwPoPKeyWithRSAKey(key)
}

// GetSwPoPKeyPersistent loads or generates a persistent PoP key for token caching.
// This ensures the same PoP key is used across multiple kubelogin invocations,
// which is required for PoP token caching with MSAL to work correctly.
//
// This implementation uses platform-specific secure storage exclusively:
// - Linux: Kernel keyrings with encrypted files
// - macOS: macOS Keychain
// - Windows: Windows Credential Manager
func GetSwPoPKeyPersistent(cacheDir string) (*SwKey, error) {
	key, err := loadOrGenerateRSAKey(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("error loading or generating persistent RSA private key from secure storage: %w", err)
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

// loadOrGenerateRSAKey loads an existing RSA key from secure storage or generates a new one if it doesn't exist.
// This uses the same encrypted storage infrastructure as our PoP token cache, providing platform-specific secure storage:
// - Linux: Kernel keyrings with encrypted files
// - macOS: macOS Keychain
// - Windows: Windows Credential Manager
func loadOrGenerateRSAKey(cacheDir string) (*rsa.PrivateKey, error) {
	// Create a secure storage accessor using our cache infrastructure
	popKeyPath := getPoPKeyFilePath(cacheDir)
	accessor, err := cache.NewSecureAccessor(popKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create secure storage accessor: %w", err)
	}

	ctx := context.Background()

	// Try to load existing key from secure storage
	if keyData, err := accessor.Read(ctx); err == nil && len(keyData) > 0 {
		if key, err := parseRSAKeyFromPEM(keyData); err == nil {
			return key, nil
		}
		// If parsing fails, we'll generate a new key below
	}

	// Generate new key if loading failed
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("error generating RSA private key: %w", err)
	}

	// Save the key to secure storage
	keyPEM := marshalRSAKeyToPEM(key)
	if err := accessor.Write(ctx, keyPEM); err != nil {
		// Log warning but don't fail - key generation succeeded
		fmt.Fprintf(os.Stderr, "Warning: failed to persist PoP key to secure storage: %v\n", err)
	}

	return key, nil
}

// parseRSAKeyFromPEM parses an RSA private key from PEM data
func parseRSAKeyFromPEM(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("invalid PEM block type")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
	}

	return key, nil
}

// marshalRSAKeyToPEM converts an RSA private key to PEM format
func marshalRSAKeyToPEM(key *rsa.PrivateKey) []byte {
	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	})
}

// GetPoPKeyByPolicy returns a PoP key based on cache directory availability.
// Uses persistent key storage when cacheDir is provided, ephemeral keys otherwise.
// This centralizes the key selection logic used across all PoP credential implementations.
func GetPoPKeyByPolicy(cacheDir string) (*SwKey, error) {
	if cacheDir != "" {
		// Use persistent key storage when cache directory is available
		popKey, err := GetSwPoPKeyPersistent(cacheDir)
		if err != nil {
			return nil, fmt.Errorf("unable to get persistent PoP key: %w", err)
		}
		return popKey, nil
	} else {
		// Use ephemeral keys when no cache directory is available
		popKey, err := GetSwPoPKey()
		if err != nil {
			return nil, fmt.Errorf("unable to generate PoP key: %w", err)
		}
		return popKey, nil
	}
}

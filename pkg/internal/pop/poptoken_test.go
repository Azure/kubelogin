package pop

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"os"
	"strings"
	"testing"
)

func TestSwPoPKey(t *testing.T) {
	t.Run("GetSwPoPKeyWithRSAKey should return a key with all the expected fields", func(t *testing.T) {
		rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Errorf("expected no error generating RSA key but got: %s", err)
		}
		key, err := GetSwPoPKeyWithRSAKey(rsaKey)
		if err != nil {
			t.Errorf("expected no error but got: %s", err)
		}

		// validate key alg
		if key.Alg() != "RS256" {
			t.Errorf("expected key alg: %s but got: %s", "RS256", key.Alg())
		}

		// validate key jwk thumbprint
		eB64, nB64 := getRSAKeyExponentAndModulus(key.key)
		expectedJWKThumbprint := computeJWKThumbprint(eB64, nB64)
		if key.JWKThumbprint() != expectedJWKThumbprint {
			t.Errorf("expected key jwt thumbprint: %s but got: %s", expectedJWKThumbprint, key.JWKThumbprint())
		}

		// validate req_cnf
		expectedReqCnf := getReqCnf(expectedJWKThumbprint)
		if key.ReqCnf() != expectedReqCnf {
			t.Errorf("expected key req_cnf: %s but got: %s", expectedReqCnf, key.ReqCnf())
		}

		// validate key ID
		if key.KeyID() != expectedJWKThumbprint {
			t.Errorf("expected key ID: %s but got: %s", expectedJWKThumbprint, key.KeyID())
		}

		// validate jwk
		expectedJWK := getJWK(eB64, nB64, expectedJWKThumbprint)
		if key.JWK() != expectedJWK {
			t.Errorf("expected key JWK: %s but got: %s", expectedJWK, key.JWK())
		}
	})

	t.Run("GetSwPoPKeyWithRSAKey should return a key with all the expected fields", func(t *testing.T) {
		rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Errorf("expected no error generating RSA key but got: %s", err)
		}

		e, n := getRSAKeyExponentAndModulus(rsaKey)
		e2, n2 := getRSAKeyExponentAndModulus(rsaKey)
		if e2 != e {
			t.Errorf("%s but got: %s", e, e2)
		}
		if n2 != n {
			t.Errorf("%s but got: %s", n, n2)
		}

		tp1 := computeJWKThumbprint(e, n)
		tp2 := computeJWKThumbprint(e2, n2)
		if tp1 != tp2 {
			t.Errorf("%s but got: %s", tp1, tp2)
		}
	})
}

func TestSecureKeyStorage(t *testing.T) {
	// Create a temporary test directory
	testDir, err := os.MkdirTemp("", "kubelogin_secure_key_test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	t.Run("GetSwPoPKeyPersistent should use secure storage and persist keys", func(t *testing.T) {
		// Generate first key (should use secure storage)
		key1, err := GetSwPoPKeyPersistent(testDir)
		if err != nil {
			t.Fatalf("Failed to generate first key: %v", err)
		}

		// Load the same key again (should load from secure storage)
		key2, err := GetSwPoPKeyPersistent(testDir)
		if err != nil {
			t.Fatalf("Failed to load second key: %v", err)
		}

		// Verify they're the same key (same KeyID means same underlying RSA key)
		if key1.KeyID() != key2.KeyID() {
			t.Errorf("Keys don't match! First KeyID: %s, Second KeyID: %s", key1.KeyID(), key2.KeyID())
		}

		// Verify JWK thumbprints match (additional verification)
		if key1.JWKThumbprint() != key2.JWKThumbprint() {
			t.Errorf("JWK thumbprints don't match! First: %s, Second: %s", key1.JWKThumbprint(), key2.JWKThumbprint())
		}
	})

	t.Run("GetSwPoPKeyPersistent should handle non-existent cache directory gracefully", func(t *testing.T) {
		nonExistentDir := "/tmp/non_existent_cache_dir_12345"

		// Should create the directory and work fine
		key, err := GetSwPoPKeyPersistent(nonExistentDir)
		if err != nil {
			t.Fatalf("Failed to generate key with non-existent cache dir: %v", err)
		}

		if key == nil {
			t.Error("Key should not be nil")
		}

		// Clean up
		os.RemoveAll(nonExistentDir)
	})
}

func TestRSAKeyConversion(t *testing.T) {
	t.Run("parseRSAKeyFromPEM and marshalRSAKeyToPEM should be reversible", func(t *testing.T) {
		// Generate a test RSA key
		originalKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatalf("Failed to generate test RSA key: %v", err)
		}

		// Convert to PEM and back
		pemData := marshalRSAKeyToPEM(originalKey)
		parsedKey, err := parseRSAKeyFromPEM(pemData)
		if err != nil {
			t.Fatalf("Failed to parse PEM data: %v", err)
		}

		// Verify they're the same key by comparing modulus
		if originalKey.N.Cmp(parsedKey.N) != 0 {
			t.Error("Original and parsed keys have different modulus")
		}

		if originalKey.E != parsedKey.E {
			t.Error("Original and parsed keys have different exponent")
		}
	})

	t.Run("parseRSAKeyFromPEM should handle invalid PEM data", func(t *testing.T) {
		invalidPEMData := []byte("invalid pem data")

		_, err := parseRSAKeyFromPEM(invalidPEMData)
		if err == nil {
			t.Error("Expected error for invalid PEM data, but got none")
		}
	})

	t.Run("parseRSAKeyFromPEM should handle wrong PEM block type", func(t *testing.T) {
		// Create a PEM block with wrong type
		wrongPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: []byte("not an RSA key"),
		})

		_, err := parseRSAKeyFromPEM(wrongPEM)
		if err == nil {
			t.Error("Expected error for wrong PEM block type, but got none")
		}

		if !strings.Contains(err.Error(), "invalid PEM block type") {
			t.Errorf("Expected 'invalid PEM block type' error, got: %v", err)
		}
	})
}

package pop

import (
	"crypto/rand"
	"crypto/rsa"
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

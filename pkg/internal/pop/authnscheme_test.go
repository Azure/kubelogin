package pop

import (
	"crypto/rand"
	"crypto/rsa"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func TestAuthnScheme(t *testing.T) {
	t.Run("FormatAccessTokenWithOptions should return a correctly formatted PoP token", func(t *testing.T) {
		accessToken := uuid.NewString()
		timestamp := time.Now().Unix()
		nonce := uuid.NewString()
		nonce = strings.ReplaceAll(nonce, "-", "")
		host := "testresource"
		popKey, err := GetSwPoPKey()
		if err != nil {
			t.Errorf("expected no error but got: %s", err)
		}
		authnScheme := &PoPAuthenticationScheme{
			Host:   host,
			PoPKey: popKey,
		}

		formatted, err := authnScheme.FormatAccessTokenWithOptions(accessToken, nonce, timestamp)
		if err != nil {
			t.Errorf("expected no error but got: %s", err)
		}
		claims := jwt.MapClaims{}
		parsed, _ := jwt.ParseWithClaims(formatted, &claims, func(token *jwt.Token) (interface{}, error) {
			return authnScheme.PoPKey.KeyID(), nil
		})
		if claims["at"] != accessToken {
			t.Errorf("expected access token: %s but got: %s", accessToken, claims["at"])
		}
		if claims["u"] != host {
			t.Errorf("expected u-claim value: %s but got: %s", host, claims["u"])
		}
		ts := int64(math.Round(claims["ts"].(float64)))
		if ts != timestamp {
			t.Errorf("expected timestamp value: %d but got: %d", timestamp, ts)
		}
		if claims["nonce"] != nonce {
			t.Errorf("expected nonce value: %s but got: %s", nonce, claims["nonce"])
		}
		if parsed.Header["typ"] != popTokenType {
			t.Errorf("expected token type: %s but got: %s", popTokenType, parsed.Header["typ"])
		}
		if parsed.Header["alg"] != authnScheme.PoPKey.Alg() {
			t.Errorf("expected token alg: %s but got: %s", authnScheme.PoPKey.Alg(), parsed.Header["alg"])
		}
		if parsed.Header["kid"] != authnScheme.KeyID() {
			t.Errorf("expected token kid: %s but got: %s", authnScheme.PoPKey.KeyID(), parsed.Header["kid"])
		}

		header := header{
			typ: popTokenType,
			alg: authnScheme.PoPKey.Alg(),
			kid: authnScheme.PoPKey.KeyID(),
		}
		payload := payload{
			at:    accessToken,
			ts:    timestamp,
			host:  host,
			jwk:   authnScheme.PoPKey.JWK(),
			nonce: nonce,
		}
		popAccessToken, err := createPoPAccessToken(header, payload, authnScheme.PoPKey)
		if err != nil {
			t.Errorf("expected no error but got: %s", err)
		}
		if parsed.Signature != popAccessToken.Signature.ToBase64() {
			t.Errorf("expected token signature: %s but got: %s", popAccessToken.Signature.ToBase64(), parsed.Signature)
		}
	})

	t.Run("TokenRequestParams should return correct token_type and req_cnf claims", func(t *testing.T) {
		host := "testresource"
		rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Errorf("expected no error generating RSA key but got: %s", err)
		}
		popKey, err := GetSwPoPKeyWithRSAKey(rsaKey)
		if err != nil {
			t.Errorf("expected no error but got: %s", err)
		}
		authnScheme := &PoPAuthenticationScheme{
			Host:   host,
			PoPKey: popKey,
		}
		tokenRequestParams := authnScheme.TokenRequestParams()

		// validate token type
		if tokenRequestParams["token_type"] != "pop" {
			t.Errorf("expected req_cnf: %s but got: %s", "pop", tokenRequestParams["token_type"])
		}

		// validate req_cnf
		eB64, nB64 := getRSAKeyExponentAndModulus(popKey.key)
		jwktp := computeJWKThumbprint(eB64, nB64)
		expectedReqCnf := getReqCnf(jwktp)
		if tokenRequestParams["req_cnf"] != expectedReqCnf {
			t.Errorf("expected req_cnf: %s but got: %s", expectedReqCnf, tokenRequestParams["req_cnf"])
		}
	})

	t.Run("AccessTokenType should return correct type", func(t *testing.T) {
		host := "testresource"
		popKey, err := GetSwPoPKey()
		if err != nil {
			t.Errorf("expected no error but got: %s", err)
		}
		authnScheme := &PoPAuthenticationScheme{
			Host:   host,
			PoPKey: popKey,
		}

		if authnScheme.AccessTokenType() != "pop" {
			t.Errorf("expected req_cnf: %s but got: %s", "pop", authnScheme.AccessTokenType())
		}
	})
}

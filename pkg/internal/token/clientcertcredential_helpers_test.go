package token

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// certTestKeyPair generates an RSA key and a matching self-signed certificate
// (returned as DER) for exercising the PEM/PKCS certificate helpers.
func certTestKeyPair(t *testing.T) (*rsa.PrivateKey, []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "kubelogin-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	return key, der
}

func certTestCertPEM(der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func certTestPKCS8PEM(t *testing.T, key *rsa.PrivateKey) []byte {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

func TestIsPublicKeyEqual(t *testing.T) {
	key1, _ := certTestKeyPair(t)
	key2, _ := certTestKeyPair(t)

	assert.True(t, isPublicKeyEqual(&key1.PublicKey, &key1.PublicKey), "identical keys should be equal")
	assert.False(t, isPublicKeyEqual(&key1.PublicKey, &key2.PublicKey), "different keys should not be equal")
	assert.False(t, isPublicKeyEqual(&rsa.PublicKey{}, &key1.PublicKey), "nil modulus (left) should not be equal")
	assert.False(t, isPublicKeyEqual(&key1.PublicKey, &rsa.PublicKey{}), "nil modulus (right) should not be equal")
}

func TestParseRsaPrivateKey(t *testing.T) {
	key, _ := certTestKeyPair(t)

	pkcs1PEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	pkcs8DER, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	pkcs8PEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8DER})

	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	ecDER, err := x509.MarshalPKCS8PrivateKey(ecKey)
	require.NoError(t, err)
	ecPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: ecDER})

	testCases := []struct {
		name         string
		input        []byte
		expectErr    bool
		expectErrMsg string
	}{
		{name: "pkcs1 rsa key", input: pkcs1PEM},
		{name: "pkcs8 rsa key", input: pkcs8PEM},
		{name: "pkcs8 non-rsa key", input: ecPEM, expectErr: true, expectErrMsg: "non-RSA"},
		{name: "not a pem block", input: []byte("not a pem block"), expectErr: true, expectErrMsg: "failed to decode"},
		{name: "empty input", input: nil, expectErr: true, expectErrMsg: "failed to decode"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseRsaPrivateKey(tc.input)
			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectErrMsg)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, 0, key.PublicKey.N.Cmp(got.PublicKey.N), "parsed key should match the original")
		})
	}
}

func TestSplitPEMBlock(t *testing.T) {
	key, der := certTestKeyPair(t)
	certPEM := certTestCertPEM(der)
	keyPEM := certTestPKCS8PEM(t, key)
	unknownPEM := pem.EncodeToMemory(&pem.Block{Type: "SOMETHING ELSE", Bytes: []byte("ignored")})

	t.Run("separates cert and key and ignores unknown blocks", func(t *testing.T) {
		combined := bytesJoin(certPEM, keyPEM, unknownPEM)
		gotCert, gotKey := splitPEMBlock(combined)

		assert.NotEmpty(t, gotCert, "expected a certificate block")
		assert.NotEmpty(t, gotKey, "expected a private key block")
		assert.NotContains(t, string(gotCert), "SOMETHING ELSE")
		assert.NotContains(t, string(gotKey), "SOMETHING ELSE")

		parsedKey, err := parseRsaPrivateKey(gotKey)
		require.NoError(t, err)
		assert.Equal(t, 0, key.PublicKey.N.Cmp(parsedKey.PublicKey.N))
	})

	t.Run("returns nil for input without pem blocks", func(t *testing.T) {
		gotCert, gotKey := splitPEMBlock([]byte("no pem here"))
		assert.Nil(t, gotCert)
		assert.Nil(t, gotKey)
	})
}

func TestParseKeyPairFromPEMBlock(t *testing.T) {
	key, der := certTestKeyPair(t)
	certPEM := certTestCertPEM(der)
	keyPEM := certTestPKCS8PEM(t, key)

	_, otherDER := certTestKeyPair(t)
	otherCertPEM := certTestCertPEM(otherDER)

	t.Run("matching cert and key", func(t *testing.T) {
		cert, priv, err := parseKeyPairFromPEMBlock(bytesJoin(certPEM, keyPEM))
		require.NoError(t, err)
		require.NotNil(t, cert)
		require.NotNil(t, priv)
		assert.Equal(t, 0, key.PublicKey.N.Cmp(priv.PublicKey.N))
	})

	t.Run("cert does not match key", func(t *testing.T) {
		_, _, err := parseKeyPairFromPEMBlock(bytesJoin(otherCertPEM, keyPEM))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unable to find a matching public certificate")
	})

	t.Run("missing private key", func(t *testing.T) {
		_, _, err := parseKeyPairFromPEMBlock(certPEM)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode")
	})
}

func TestDecodePkcs12(t *testing.T) {
	t.Run("invalid pkcs12 data returns error", func(t *testing.T) {
		cert, key, err := decodePkcs12([]byte("not a valid pkcs12 blob"), "")
		require.Error(t, err)
		assert.Nil(t, cert)
		assert.Nil(t, key)
	})
}

func TestReadCertificate(t *testing.T) {
	key, der := certTestKeyPair(t)
	combined := bytesJoin(certTestCertPEM(der), certTestPKCS8PEM(t, key))
	dir := t.TempDir()

	t.Run("valid pem file", func(t *testing.T) {
		p := filepath.Join(dir, "cert.pem")
		require.NoError(t, os.WriteFile(p, combined, 0o600))

		cert, priv, err := readCertificate(p, "")
		require.NoError(t, err)
		require.NotNil(t, cert)
		assert.Equal(t, 0, key.PublicKey.N.Cmp(priv.PublicKey.N))
	})

	t.Run("missing pem file", func(t *testing.T) {
		_, _, err := readCertificate(filepath.Join(dir, "does-not-exist.pem"), "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read the certificate file")
	})

	t.Run("missing pfx file", func(t *testing.T) {
		_, _, err := readCertificate(filepath.Join(dir, "does-not-exist.pfx"), "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read the certificate file")
	})
}

// bytesJoin concatenates byte slices without a separator.
func bytesJoin(parts ...[]byte) []byte {
	var out []byte
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

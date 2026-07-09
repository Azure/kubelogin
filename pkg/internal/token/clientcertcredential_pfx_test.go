package token

import (
	"crypto/rsa"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testPFXBase64 is a throwaway, self-signed RSA certificate packaged as a
// password-protected PKCS#12 (PFX) file, base64-encoded. The password is
// testPFXPassword. It intentionally uses legacy SHA1/3DES PBE so it can be
// decoded by golang.org/x/crypto/pkcs12, which does not support the AES-based
// algorithms OpenSSL 3 emits by default.
//
// Regenerate with:
//
//	openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem \
//	    -days 3650 -nodes -subj "/CN=kubelogin-test"
//	openssl pkcs12 -export -inkey key.pem -in cert.pem -out test.pfx \
//	    -passout pass:kubelogin-test -legacy \
//	    -certpbe PBE-SHA1-3DES -keypbe PBE-SHA1-3DES -macalg sha1
//	base64 -w0 test.pfx
const testPFXBase64 = "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"

const testPFXPassword = "kubelogin-test"

func testPFXBytes(t *testing.T) []byte {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString(testPFXBase64)
	require.NoError(t, err)
	return data
}

func TestDecodePkcs12WithPassword(t *testing.T) {
	pfx := testPFXBytes(t)

	t.Run("correct password decodes cert and key", func(t *testing.T) {
		cert, key, err := decodePkcs12(pfx, testPFXPassword)
		require.NoError(t, err)
		require.NotNil(t, cert)
		require.NotNil(t, key)
		assert.Equal(t, "kubelogin-test", cert.Subject.CommonName)
		assert.IsType(t, &rsa.PrivateKey{}, key)
		require.NoError(t, key.Validate())
	})

	t.Run("wrong password returns error", func(t *testing.T) {
		cert, key, err := decodePkcs12(pfx, "not-the-password")
		require.Error(t, err)
		assert.Nil(t, cert)
		assert.Nil(t, key)
	})
}

func TestReadCertificatePFX(t *testing.T) {
	pfx := testPFXBytes(t)
	dir := t.TempDir()
	pfxPath := filepath.Join(dir, "client.pfx")
	require.NoError(t, os.WriteFile(pfxPath, pfx, 0o600))

	t.Run("valid pfx with correct password", func(t *testing.T) {
		cert, key, err := readCertificate(pfxPath, testPFXPassword)
		require.NoError(t, err)
		require.NotNil(t, cert)
		require.NotNil(t, key)
		assert.Equal(t, "kubelogin-test", cert.Subject.CommonName)
	})

	t.Run("valid pfx with wrong password", func(t *testing.T) {
		_, _, err := readCertificate(pfxPath, "not-the-password")
		require.Error(t, err)
	})
}

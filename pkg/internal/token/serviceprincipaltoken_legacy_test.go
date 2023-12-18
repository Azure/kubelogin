package token

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLegacyServicePrincipalToken(t *testing.T) {
	t.Run("new spn token provider with legacy should not result in error", func(t *testing.T) {
		_, err := newTokenProvider(&Options{
			LoginMethod:  ServicePrincipalLogin,
			IsLegacy:     true,
			TenantID:     "tenantID",
			ClientID:     "client-id",
			ClientSecret: "foobar",
			ServerID:     "server-id",
			Environment:  "AzurePublicCloud",
		})

		require.NoError(t, err)
	})

	t.Run("legacy spn token provider with incorrectly setup oauth config should result in error", func(t *testing.T) {
		oathConfig, err := getOAuthConfig("AzurePublicCloud", "tenantID", false)
		require.NoError(t, err)

		_, err = newLegacyServicePrincipalToken(*oathConfig, "client-id", "client-secret", "", "", "server-id", "tenant-id")
		require.ErrorIs(t, err, errInvalidOAuthConfig)
	})
}

package token

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/kubelogin/pkg/internal/token/mock_token"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetTokenProvider(t *testing.T) {
	t.Run("invalid login method", func(t *testing.T) {
		opts := &Options{
			LoginMethod: "invalid-login-method",
		}
		tp, err := GetTokenProvider(opts)
		assert.Error(t, err)
		assert.Nil(t, tp)
	})

	t.Run("basic", func(t *testing.T) {
		opts := &Options{
			LoginMethod:        MSILogin,
			ClientID:           "client-id",
			IdentityResourceID: "identity-resource-id",
			ServerID:           "server-id",
		}
		tp, err := GetTokenProvider(opts)
		assert.NoError(t, err)
		assert.NotNil(t, tp)
	})
}

func TestTokenProviderShim_GetAccessToken(t *testing.T) {
	t.Run("failure case", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockTokenProvider := mock_token.NewMockTokenProvider(mockCtrl)
		mockTokenProvider.EXPECT().Token(gomock.Any()).Return(adal.Token{}, assert.AnError)

		tp := &tokenProviderShim{
			impl: mockTokenProvider,
		}

		token, err := tp.GetAccessToken(context.Background())
		assert.Equal(t, AccessToken{}, token)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("success case", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		adalToken := adal.Token{
			AccessToken: "access-token",
			ExpiresOn:   json.Number("1700000000"),
		}
		mockTokenProvider := mock_token.NewMockTokenProvider(mockCtrl)
		mockTokenProvider.EXPECT().Token(gomock.Any()).Return(adalToken, nil)

		tp := &tokenProviderShim{
			impl: mockTokenProvider,
		}

		token, err := tp.GetAccessToken(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, adalToken.AccessToken, token.Token)
		assert.Equal(t, adalToken.Expires(), token.ExpiresOn)
	})
}

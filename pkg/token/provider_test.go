package token

import (
	"context"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/kubelogin/pkg/internal/token"
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

		credProvider := mock_token.NewMockCredentialProvider(mockCtrl)
		credProvider.EXPECT().GetToken(gomock.Any(), gomock.Any()).Return(azcore.AccessToken{}, assert.AnError)

		tp := &tokenProviderShim{
			cred: credProvider,
			opts: &token.Options{
				TenantID: "tenant-id",
				ServerID: "server-id",
			},
		}

		token, err := tp.GetAccessToken(context.Background())
		assert.Equal(t, AccessToken{}, token)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("success case", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		expectedToken := azcore.AccessToken{
			Token:     "access-token",
			ExpiresOn: time.Unix(1700000000, 0),
		}
		credProvider := mock_token.NewMockCredentialProvider(mockCtrl)
		credProvider.EXPECT().GetToken(gomock.Any(), gomock.Any()).Return(expectedToken, nil)

		tp := &tokenProviderShim{
			cred: credProvider,
			opts: &token.Options{
				TenantID: "tenant-id",
				ServerID: "server-id",
			},
		}

		token, err := tp.GetAccessToken(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expectedToken.Token, token.Token)
		assert.Equal(t, expectedToken.ExpiresOn, token.ExpiresOn)
	})
}

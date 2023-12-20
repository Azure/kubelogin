package token

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

// AccessToken represents an Azure service bearer access token with expiry information.
type AccessToken = azcore.AccessToken

// TokenProvider provides access to tokens.
type TokenProvider interface {
	// GetAccessToken returns an access token from given settings.
	GetAccessToken(ctx context.Context) (AccessToken, error)
}

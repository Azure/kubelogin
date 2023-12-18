package token

import (
	"errors"
	"testing"

	"github.com/Azure/go-autorest/autorest/adal"
)

func TestNewManualToken(t *testing.T) {
	testCases := []struct {
		name       string
		oAuthCfg   adal.OAuthConfig
		clientID   string
		resourceID string
		tenantID   string
		token      *adal.Token
		wantErr    error
	}{
		{
			name:       "Valid input",
			oAuthCfg:   adal.OAuthConfig{},
			clientID:   "clientID",
			resourceID: "resourceID",
			tenantID:   "tenantID",
			token:      &adal.Token{},
			wantErr:    nil,
		},
		{
			name:       "Nil token",
			oAuthCfg:   adal.OAuthConfig{},
			clientID:   "clientID",
			resourceID: "resourceID",
			tenantID:   "tenantID",
			token:      nil,
			wantErr:    errors.New("token cannot be nil"),
		},
		{
			name:       "Empty clientID",
			oAuthCfg:   adal.OAuthConfig{},
			clientID:   "",
			resourceID: "resourceID",
			tenantID:   "tenantID",
			token:      &adal.Token{},
			wantErr:    errors.New("clientID cannot be empty"),
		},
		{
			name:       "Empty resourceID",
			oAuthCfg:   adal.OAuthConfig{},
			clientID:   "clientID",
			resourceID: "",
			tenantID:   "tenantID",
			token:      &adal.Token{},
			wantErr:    errors.New("resourceID cannot be empty"),
		},
		{
			name:       "Empty tenantID",
			oAuthCfg:   adal.OAuthConfig{},
			clientID:   "clientID",
			resourceID: "resourceID",
			tenantID:   "",
			token:      &adal.Token{},
			wantErr:    errors.New("tenantID cannot be empty"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newManualToken(tc.oAuthCfg, tc.clientID, tc.resourceID, tc.tenantID, tc.token)
			if err != nil && err.Error() != tc.wantErr.Error() {
				t.Errorf("expected error %v, but got %v", tc.wantErr, err)
			}
			if err == nil && tc.wantErr != nil {
				t.Errorf("expected err: %v, got nil", tc.wantErr)
			}
		})
	}
}

func TestManualTokenToken(t *testing.T) {
	oAuthConfig := adal.OAuthConfig{}
	clientID := "test-client-id"
	resourceID := "test-resource-id"
	tenantID := "test-tenant-id"
	token := &adal.Token{AccessToken: "test-access-token"}

	provider, _ := newManualToken(oAuthConfig, clientID, resourceID, tenantID, token)

	// Test successful token refresh
	if _, err := provider.Token(); err == nil {
		if err == nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	}
}

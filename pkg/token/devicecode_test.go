package token

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/go-autorest/autorest/adal"
)

func TestNewDeviceCodeTokenProviderEmpty(t *testing.T) {
	testData := []struct {
		name string
	}{
		{
			name: "clientID cannot be empty",
		},
		{
			name: "resourceID cannot be empty",
		},
		{
			name: "tenantID cannot be empty",
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {

			name := data.name
			var err error

			switch {
			case strings.Contains(name, "clientID"):
				_, err = newDeviceCodeTokenProvider(adal.OAuthConfig{}, "", "", "", nil)
			case strings.Contains(name, "resourceID"):
				_, err = newDeviceCodeTokenProvider(adal.OAuthConfig{}, "test", "", "", nil)
			case strings.Contains(name, "tenantID"):
				_, err = newDeviceCodeTokenProvider(adal.OAuthConfig{}, "test", "test", "", nil)
			default:
				fmt.Println(false)
			}

			if !ErrorContains(err, data.name) {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestNewDeviceCodeToken(t *testing.T) {
	deviceCode := deviceCodeTokenProvider{}
	_, err := deviceCode.Token()

	if !ErrorContains(err, "initialing the device code authentication:") {
		t.Errorf("unexpected error: %v", err)
	}
}

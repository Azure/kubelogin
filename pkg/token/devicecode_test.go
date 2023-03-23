package token

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/go-autorest/autorest/adal"
)

func TestNewDeviceCodeTokenProviderEmpty(t *testing.T) {
	testData := []struct {
		name          string
		inputEmptyVar string
	}{
		{
			name:          "clientID cannot be empty",
			inputEmptyVar: "",
		},
		{
			name:          "resourceID cannot be empty",
			inputEmptyVar: "",
		},
		{
			name:          "tenantID cannot be empty",
			inputEmptyVar: "",
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {

			name := data.name
			var err error

			switch {
			case strings.Contains(name, "clientID"):
				_, err = newDeviceCodeTokenProvider(adal.OAuthConfig{}, "", "", "")
			case strings.Contains(name, "resourceID"):
				_, err = newDeviceCodeTokenProvider(adal.OAuthConfig{}, "test", "", "")
			case strings.Contains(name, "tenantID"):
				_, err = newDeviceCodeTokenProvider(adal.OAuthConfig{}, "test", "test", "")
			default:
				fmt.Println(false)
			}

			if !ErrorContains(err, data.name) {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

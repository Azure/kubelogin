package token

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewWorkloadIdentityTokenProvderEmpty(t *testing.T) {
	testData := []struct {
		name string
	}{
		{
			name: "clientID cannot be empty",
		},
		{
			name: "tenantID cannot be empty",
		},
		{
			name: "federatedTokenFile cannot be empty",
		},
		{
			name: "authorityHost cannot be empty",
		},
		{
			name: "serverID cannot be empty",
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {

			name := data.name
			var err error

			switch {
			case strings.Contains(name, "clientID"):
				_, err = newWorkloadIdentityToken("", "", "", "", "")
			case strings.Contains(name, "federatedTokenFile"):
				_, err = newWorkloadIdentityToken("test", "", "", "", "test")
			case strings.Contains(name, "authorityHost"):
				_, err = newWorkloadIdentityToken("test", "test", "", "", "test")
			case strings.Contains(name, "serverID"):
				_, err = newWorkloadIdentityToken("test", "test", "test", "", "test")
			case strings.Contains(name, "tenantID"):
				_, err = newWorkloadIdentityToken("test", "test", "test", "test", "")
			default:
				fmt.Println(false)
			}

			if !ErrorContains(err, data.name) {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestNewWorkloadIdentityToken(t *testing.T) {
	workloadIdentityToken := workloadIdentityToken{}
	_, err := workloadIdentityToken.Token()

	if !ErrorContains(err, "failed to read signed assertion from token file:") {
		t.Errorf("unexpected error: %v", err)
	}
}

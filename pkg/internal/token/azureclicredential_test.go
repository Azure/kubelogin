package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAzureCLICredential(t *testing.T) {
	testCases := []struct {
		name           string
		opts           *Options
		expectErrorMsg string
		expectName     string
	}{
		{
			name: "valid options",
			opts: &Options{
				TenantID: "test-tenant-id",
			},
			expectName: "AzureCLICredential",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cred, err := newAzureCLICredential(tc.opts)
			if tc.expectErrorMsg != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.expectErrorMsg, err.Error())
				assert.Nil(t, cred)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cred)
				assert.Equal(t, tc.expectName, cred.Name())
			}
		})
	}
}

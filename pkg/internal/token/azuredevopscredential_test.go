package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAzureDeveloperCLICredential(t *testing.T) {
	testCases := []struct {
		name           string
		opts           *Options
		expectError    bool
		expectErrorMsg string
		expectName     string
	}{
		{
			name: "valid options",
			opts: &Options{
				TenantID: "test-tenant-id",
			},
			expectError: false,
			expectName:  "AzureDeveloperCLICredential",
		},
		{
			name:           "missing tenant ID",
			opts:           &Options{},
			expectError:    true,
			expectErrorMsg: "tenant ID cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cred, err := newAzureDeveloperCLICredential(tc.opts)

			if tc.expectError {
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

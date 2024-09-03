package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAzureCLICredential(t *testing.T) {
	testCases := []struct {
		name           string
		opts           *Options
		expectError    bool
		expectErrorMsg string
		expectNil      bool
		expectName     string
	}{
		{
			name: "valid options",
			opts: &Options{
				TenantID: "test-tenant-id",
			},
			expectError: false,
			expectNil:   false,
			expectName:  "AzureCLICredential",
		},
		{
			name:           "missing tenant ID",
			opts:           &Options{},
			expectError:    true,
			expectErrorMsg: "tenant ID cannot be empty",
			expectNil:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cred, err := newAzureCLICredential(tc.opts)
			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, tc.expectErrorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
			if tc.expectNil {
				assert.Nil(t, cred)
			} else {
				assert.NotNil(t, cred)
				assert.Equal(t, tc.expectName, cred.Name())
			}
		})
	}
}

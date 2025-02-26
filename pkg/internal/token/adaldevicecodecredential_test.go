package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewADALDeviceCodeCredential(t *testing.T) {
	testCases := []struct {
		name     string
		opts     *Options
		expected string
	}{
		{
			name: "valid options",
			opts: &Options{
				ClientID: "test-client-id",
				TenantID: "test-tenant-id",
				IsLegacy: true,
			},
			expected: "ADALDeviceCodeCredential",
		},
		{
			name: "missing client ID",
			opts: &Options{
				TenantID: "test-tenant-id",
				IsLegacy: true,
			},
			expected: "client ID cannot be empty",
		},
		{
			name: "missing tenant ID",
			opts: &Options{
				ClientID: "test-client-id",
				IsLegacy: true,
			},
			expected: "tenant ID cannot be empty",
		},
		{
			name: "non-legacy mode",
			opts: &Options{
				ClientID: "test-client-id",
				TenantID: "test-tenant-id",
				IsLegacy: false,
			},
			expected: "ADALDeviceCodeCredential is not supported in non-legacy mode",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cred, err := newADALDeviceCodeCredential(tc.opts)
			if err != nil {
				assert.EqualError(t, err, tc.expected)
				assert.Nil(t, cred)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cred)
				assert.Equal(t, tc.expected, cred.Name())
			}
		})
	}
}

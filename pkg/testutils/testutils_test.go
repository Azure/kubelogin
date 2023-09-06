package testutils

import (
	"fmt"
	"testing"
)

func TestErrorContains(t *testing.T) {
	testCase := []struct {
		name             string
		err              error
		desiredSubstring string
		expectedResult   bool
	}{
		{
			name:             "should return true if error is nil and desired substring is empty string",
			desiredSubstring: "",
			err:              nil,
			expectedResult:   true,
		},
		{
			name:             "should return true if error is nil and desired substring is whitespace",
			desiredSubstring: "		   ",
			err:              nil,
			expectedResult:   true,
		},
		{
			name:             "should return false if error is not nil and desired substring is empty string",
			desiredSubstring: "",
			err:              fmt.Errorf("test error"),
			expectedResult:   false,
		},
		{
			name:             "should return false if error is not nil and desired substring is whitespace",
			desiredSubstring: "	   		",
			err:              fmt.Errorf("test error"),
			expectedResult:   false,
		},
		{
			name:             "should return false if error is not nil and desired substring is not contained in error",
			desiredSubstring: "not a test error",
			err:              fmt.Errorf("test error"),
			expectedResult:   false,
		},
		{
			name:             "should return true if error is not nil and desired substring is smaller than but contained in error",
			desiredSubstring: "error",
			err:              fmt.Errorf("test error"),
			expectedResult:   true,
		},
		{
			name:             "should return true if error is not nil and desired substring is the same as error string",
			desiredSubstring: "test error",
			err:              fmt.Errorf("test error"),
			expectedResult:   true,
		},
		{
			name:             "should return false if error is not nil and desired substring is the same as error string but has different casing",
			desiredSubstring: "Test Error",
			err:              fmt.Errorf("test error"),
			expectedResult:   false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			result := ErrorContains(tc.err, tc.desiredSubstring)
			if result != tc.expectedResult {
				t.Errorf(
					"comparing error: %s and desired substring: %s, expected %t but got %t",
					tc.err,
					tc.desiredSubstring,
					tc.expectedResult,
					result,
				)
			}
		})
	}
}

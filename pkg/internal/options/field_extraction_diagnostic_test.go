package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd/api"
)

// TestFieldExtractionValidationIssue demonstrates the core bug where unified mode
// validates fields before extracting them from kubeconfig, causing false validation errors
func TestFieldExtractionValidationIssue(t *testing.T) {
	tests := []struct {
		name             string
		options          *UnifiedOptions
		authInfo         *api.AuthInfo
		expectValidation bool // Should validation succeed?
		expectExtraction bool // Should field extraction work if validation is bypassed?
		description      string
	}{
		{
			name: "unified mode fails validation when tenant-id exists in kubeconfig but not provided by user",
			options: &UnifiedOptions{
				command:     ConvertCommand,
				LoginMethod: "devicecode", // Valid login method
				// TenantID is intentionally empty - should be extracted from kubeconfig
			},
			authInfo: &api.AuthInfo{
				Exec: &api.ExecConfig{
					Command: "kubelogin",
					Args: []string{
						"get-token",
						"--server-id", "test-server",
						"--client-id", "80faf920-1908-4b52-b5ef-a8e7bedfc67a",
						"--tenant-id", "test-tenant", // This exists in kubeconfig
						"--environment", "AzurePublicCloud",
						"--client-secret", "test-secret",
						"--login", "spn",
					},
				},
			},
			expectValidation: false, // Current bug: validation fails because TenantID="" and ClientID="" in options
			expectExtraction: true,  // Extraction logic works correctly
			description:      "Demonstrates that validation incorrectly fails when required fields exist in kubeconfig but not provided by user",
		},
		{
			name: "unified mode succeeds when all required fields are explicitly provided by user",
			options: &UnifiedOptions{
				command:     ConvertCommand,
				LoginMethod: "devicecode",
				TenantID:    "user-provided-tenant", // Explicitly provided by user
				ClientID:    "user-provided-client", // Also required
				Timeout:     300,                    // Also required (timeout > 0)
			},
			authInfo: &api.AuthInfo{
				Exec: &api.ExecConfig{
					Command: "kubelogin",
					Args: []string{
						"get-token",
						"--server-id", "test-server",
						"--client-id", "80faf920-1908-4b52-b5ef-a8e7bedfc67a",
						"--tenant-id", "test-tenant", // Different value in kubeconfig
						"--environment", "AzurePublicCloud",
						"--client-secret", "test-secret",
						"--login", "spn",
					},
				},
			},
			expectValidation: true, // Validation succeeds because user provided all required fields
			expectExtraction: true, // Extraction works but user value takes precedence
			description:      "Shows that validation works when user explicitly provides all required fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log("Description:", tt.description)

			// Test 1: Current validation behavior
			t.Run("validation_behavior", func(t *testing.T) {
				err := tt.options.ValidateForTokenExecution()
				if tt.expectValidation {
					assert.NoError(t, err, "Validation should succeed")
				} else {
					assert.Error(t, err, "Validation should fail (demonstrating the bug)")
					assert.Contains(t, err.Error(), "tenantid is required", "Should specifically fail on tenant-id requirement")
					assert.Contains(t, err.Error(), "clientid is required", "Should also fail on client-id requirement")
				}
			})

			// Test 2: Field extraction behavior (bypassing validation)
			t.Run("extraction_behavior", func(t *testing.T) {
				// Test the extraction logic directly
				extracted := tt.options.extractExistingValues(tt.authInfo)

				if tt.expectExtraction {
					assert.Contains(t, extracted, "--tenant-id", "Should extract tenant-id from kubeconfig")
					assert.Equal(t, "test-tenant", extracted["--tenant-id"], "Should extract correct tenant-id value")
					assert.Contains(t, extracted, "--client-id", "Should extract client-id from kubeconfig")
					assert.Equal(t, "80faf920-1908-4b52-b5ef-a8e7bedfc67a", extracted["--client-id"], "Should extract correct client-id value")
				}
			})

			// Test 3: buildExecConfig behavior (shows extraction works)
			t.Run("build_exec_config_behavior", func(t *testing.T) {
				// Skip if validation would fail (which is the bug we're testing)
				if !tt.expectValidation {
					t.Skip("Skipping buildExecConfig test because validation fails (this is the bug)")
					return
				}

				execConfig, err := tt.options.buildExecConfig(tt.authInfo)
				require.NoError(t, err, "buildExecConfig should succeed when validation passes")

				// Check that tenant-id is included in the args
				found := false
				for i, arg := range execConfig.Args {
					if arg == "--tenant-id" && i+1 < len(execConfig.Args) {
						found = true
						if tt.options.TenantID != "" {
							// User provided value should take precedence
							assert.Equal(t, tt.options.TenantID, execConfig.Args[i+1])
						} else {
							// Should use extracted value from kubeconfig
							assert.Equal(t, "test-tenant", execConfig.Args[i+1])
						}
						break
					}
				}
				assert.True(t, found, "Should include --tenant-id in exec config args")
			})
		})
	}
}

// TestDemonstrateBugFix will be used to test the fix once implemented
func TestDemonstrateBugFix(t *testing.T) {
	t.Skip("This test will be implemented once we fix the validation issue")

	// This test will demonstrate that after fixing the bug:
	// 1. Validation should succeed when fields exist in kubeconfig (even if not provided by user)
	// 2. Field extraction should work correctly
	// 3. buildExecConfig should include extracted values
	// 4. The overall conversion should succeed
}

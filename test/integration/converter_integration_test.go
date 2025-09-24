package integration

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// KubeconfigUser represents the user section of a kubeconfig
type KubeconfigUser struct {
	Name string `yaml:"name"`
	User struct {
		Exec *struct {
			APIVersion  string   `yaml:"apiVersion"`
			Command     string   `yaml:"command"`
			Args        []string `yaml:"args"`
			InstallHint string   `yaml:"installHint"`
		} `yaml:"exec,omitempty"`
	} `yaml:"user"`
}

// Kubeconfig represents a simplified kubeconfig structure for testing
type Kubeconfig struct {
	APIVersion string           `yaml:"apiVersion"`
	Kind       string           `yaml:"kind"`
	Users      []KubeconfigUser `yaml:"users"`
}

// TestKubeconfigConversion tests that kubeconfig conversion works consistently
func TestKubeconfigConversion(t *testing.T) {
	testCases := []struct {
		name           string
		fixture        string
		args           []string
		expectSuccess  bool
		validateResult func(t *testing.T, originalContent, resultContent string)
	}{
		// ===== Device Code Login Tests =====
		{
			name:          "convert to devicecode (default)",
			fixture:       "sample_kubeconfig.yaml",
			args:          []string{"--tenant-id", "new-tenant-id", "--client-id", "new-client-id", "--server-id", "6dae42f8-4368-4678-94ff-3960e28e3630"},
			expectSuccess: true,
			validateResult: func(t *testing.T, originalContent, resultContent string) {
				var config Kubeconfig
				err := yaml.Unmarshal([]byte(resultContent), &config)
				require.NoError(t, err)

				require.Len(t, config.Users, 1)
				user := config.Users[0]
				require.NotNil(t, user.User.Exec)
				assert.Equal(t, "kubelogin", user.User.Exec.Command)
				assert.Contains(t, user.User.Exec.Args, "get-token")
				assert.Contains(t, user.User.Exec.Args, "devicecode") // Default login method
			},
		},
		{
			name:          "convert to devicecode explicit",
			fixture:       "sample_kubeconfig.yaml",
			args:          []string{"--login", "devicecode", "--tenant-id", "new-tenant-id", "--client-id", "new-client-id", "--server-id", "6dae42f8-4368-4678-94ff-3960e28e3630"},
			expectSuccess: true,
			validateResult: func(t *testing.T, originalContent, resultContent string) {
				var config Kubeconfig
				err := yaml.Unmarshal([]byte(resultContent), &config)
				require.NoError(t, err)

				require.Len(t, config.Users, 1)
				user := config.Users[0]
				require.NotNil(t, user.User.Exec)
				assert.Contains(t, user.User.Exec.Args, "--login")
				assert.Contains(t, user.User.Exec.Args, "devicecode")
			},
		},

		// ===== Service Principal Tests =====
		{
			name:          "convert to spn with client secret",
			fixture:       "sample_kubeconfig.yaml",
			args:          []string{"--login", "spn", "--tenant-id", "new-tenant-id", "--client-id", "new-client-id", "--client-secret", "new-secret", "--server-id", "6dae42f8-4368-4678-94ff-3960e28e3630"},
			expectSuccess: true,
			validateResult: func(t *testing.T, originalContent, resultContent string) {
				var config Kubeconfig
				err := yaml.Unmarshal([]byte(resultContent), &config)
				require.NoError(t, err)

				require.Len(t, config.Users, 1)
				user := config.Users[0]
				require.NotNil(t, user.User.Exec)
				args := user.User.Exec.Args
				assert.Contains(t, args, "--login")
				assert.Contains(t, args, "spn")
				assert.Contains(t, args, "--client-secret")
				assert.Contains(t, args, "new-secret")
			},
		},
		{
			name:          "convert to spn with client certificate",
			fixture:       "sample_kubeconfig.yaml",
			args:          []string{"--login", "spn", "--tenant-id", "new-tenant-id", "--client-id", "new-client-id", "--client-certificate", "/path/to/cert.pem", "--server-id", "6dae42f8-4368-4678-94ff-3960e28e3630"},
			expectSuccess: true,
			validateResult: func(t *testing.T, originalContent, resultContent string) {
				var config Kubeconfig
				err := yaml.Unmarshal([]byte(resultContent), &config)
				require.NoError(t, err)

				require.Len(t, config.Users, 1)
				user := config.Users[0]
				require.NotNil(t, user.User.Exec)
				args := user.User.Exec.Args
				assert.Contains(t, args, "--login")
				assert.Contains(t, args, "spn")
				assert.Contains(t, args, "--client-certificate")
				assert.Contains(t, args, "/path/to/cert.pem")
				assert.NotContains(t, args, "--client-secret") // Should not have secret when using cert
			},
		},

		// ===== Interactive Login Tests =====
		{
			name:          "convert to interactive",
			fixture:       "sample_kubeconfig.yaml",
			args:          []string{"--login", "interactive", "--tenant-id", "new-tenant-id", "--client-id", "new-client-id", "--server-id", "6dae42f8-4368-4678-94ff-3960e28e3630"},
			expectSuccess: true,
			validateResult: func(t *testing.T, originalContent, resultContent string) {
				var config Kubeconfig
				err := yaml.Unmarshal([]byte(resultContent), &config)
				require.NoError(t, err)

				require.Len(t, config.Users, 1)
				user := config.Users[0]
				require.NotNil(t, user.User.Exec)
				args := user.User.Exec.Args
				assert.Contains(t, args, "--login")
				assert.Contains(t, args, "interactive")
			},
		},

		// ===== Managed Service Identity Tests =====
		{
			name:          "convert to msi (default identity)",
			fixture:       "sample_kubeconfig.yaml",
			args:          []string{"--login", "msi", "--server-id", "6dae42f8-4368-4678-94ff-3960e28e3630"},
			expectSuccess: true,
			validateResult: func(t *testing.T, originalContent, resultContent string) {
				var config Kubeconfig
				err := yaml.Unmarshal([]byte(resultContent), &config)
				require.NoError(t, err)

				require.Len(t, config.Users, 1)
				user := config.Users[0]
				require.NotNil(t, user.User.Exec)
				args := user.User.Exec.Args
				assert.Contains(t, args, "--login")
				assert.Contains(t, args, "msi")
			},
		},
		{
			name:          "convert to msi with specific client id",
			fixture:       "sample_kubeconfig.yaml",
			args:          []string{"--login", "msi", "--client-id", "msi-client-id", "--server-id", "6dae42f8-4368-4678-94ff-3960e28e3630"},
			expectSuccess: true,
			validateResult: func(t *testing.T, originalContent, resultContent string) {
				var config Kubeconfig
				err := yaml.Unmarshal([]byte(resultContent), &config)
				require.NoError(t, err)

				require.Len(t, config.Users, 1)
				user := config.Users[0]
				require.NotNil(t, user.User.Exec)
				args := user.User.Exec.Args
				assert.Contains(t, args, "--login")
				assert.Contains(t, args, "msi")
				assert.Contains(t, args, "--client-id")
				assert.Contains(t, args, "msi-client-id")
			},
		},

		// ===== Azure CLI Tests =====
		{
			name:          "convert to azurecli",
			fixture:       "sample_kubeconfig.yaml",
			args:          []string{"--login", "azurecli", "--server-id", "6dae42f8-4368-4678-94ff-3960e28e3630"},
			expectSuccess: true,
			validateResult: func(t *testing.T, originalContent, resultContent string) {
				var config Kubeconfig
				err := yaml.Unmarshal([]byte(resultContent), &config)
				require.NoError(t, err)

				require.Len(t, config.Users, 1)
				user := config.Users[0]
				require.NotNil(t, user.User.Exec)
				args := user.User.Exec.Args
				assert.Contains(t, args, "--login")
				assert.Contains(t, args, "azurecli")
			},
		},

		// ===== Azure Developer CLI Tests =====
		{
			name:          "convert to azd",
			fixture:       "sample_kubeconfig.yaml",
			args:          []string{"--login", "azd", "--server-id", "6dae42f8-4368-4678-94ff-3960e28e3630"},
			expectSuccess: true,
			validateResult: func(t *testing.T, originalContent, resultContent string) {
				var config Kubeconfig
				err := yaml.Unmarshal([]byte(resultContent), &config)
				require.NoError(t, err)

				require.Len(t, config.Users, 1)
				user := config.Users[0]
				require.NotNil(t, user.User.Exec)
				args := user.User.Exec.Args
				assert.Contains(t, args, "--login")
				assert.Contains(t, args, "azd")
			},
		},

		// ===== Workload Identity Tests =====
		{
			name:          "convert to workloadidentity",
			fixture:       "sample_kubeconfig.yaml",
			args:          []string{"--login", "workloadidentity", "--tenant-id", "new-tenant-id", "--client-id", "new-client-id", "--server-id", "6dae42f8-4368-4678-94ff-3960e28e3630"},
			expectSuccess: true,
			validateResult: func(t *testing.T, originalContent, resultContent string) {
				var config Kubeconfig
				err := yaml.Unmarshal([]byte(resultContent), &config)
				require.NoError(t, err)

				require.Len(t, config.Users, 1)
				user := config.Users[0]
				require.NotNil(t, user.User.Exec)
				args := user.User.Exec.Args
				assert.Contains(t, args, "--login")
				assert.Contains(t, args, "workloadidentity")
			},
		},

		// ===== ROPC (Resource Owner Password Credential) Tests =====
		{
			name:          "convert to ropc",
			fixture:       "sample_kubeconfig.yaml",
			args:          []string{"--login", "ropc", "--tenant-id", "new-tenant-id", "--client-id", "new-client-id", "--username", "test-user", "--password", "test-password", "--server-id", "6dae42f8-4368-4678-94ff-3960e28e3630"},
			expectSuccess: true,
			validateResult: func(t *testing.T, originalContent, resultContent string) {
				var config Kubeconfig
				err := yaml.Unmarshal([]byte(resultContent), &config)
				require.NoError(t, err)

				require.Len(t, config.Users, 1)
				user := config.Users[0]
				require.NotNil(t, user.User.Exec)
				args := user.User.Exec.Args
				assert.Contains(t, args, "--login")
				assert.Contains(t, args, "ropc")
				assert.Contains(t, args, "--username")
				assert.Contains(t, args, "test-user")
				assert.Contains(t, args, "--password")
				assert.Contains(t, args, "test-password")
			},
		},

		// ===== Environment Configuration Tests =====
		{
			name:          "convert with custom environment",
			fixture:       "sample_kubeconfig.yaml",
			args:          []string{"--login", "spn", "--tenant-id", "new-tenant-id", "--client-id", "new-client-id", "--client-secret", "new-secret", "--server-id", "6dae42f8-4368-4678-94ff-3960e28e3630", "--environment", "AzureUSGovernment"},
			expectSuccess: true,
			validateResult: func(t *testing.T, originalContent, resultContent string) {
				var config Kubeconfig
				err := yaml.Unmarshal([]byte(resultContent), &config)
				require.NoError(t, err)

				require.Len(t, config.Users, 1)
				user := config.Users[0]
				require.NotNil(t, user.User.Exec)
				args := user.User.Exec.Args
				assert.Contains(t, args, "--environment")
				assert.Contains(t, args, "AzureUSGovernment")
			},
		},

		// ===== Proof-of-Possession (PoP) Token Tests =====
		{
			name:          "convert spn with pop token",
			fixture:       "sample_kubeconfig.yaml",
			args:          []string{"--login", "spn", "--tenant-id", "new-tenant-id", "--client-id", "new-client-id", "--client-secret", "new-secret", "--server-id", "6dae42f8-4368-4678-94ff-3960e28e3630", "--pop-enabled", "--pop-claims", "u=/subscriptions/123/resourcegroups/rg/providers/Microsoft.ContainerService/managedClusters/cluster"},
			expectSuccess: true,
			validateResult: func(t *testing.T, originalContent, resultContent string) {
				var config Kubeconfig
				err := yaml.Unmarshal([]byte(resultContent), &config)
				require.NoError(t, err)

				require.Len(t, config.Users, 1)
				user := config.Users[0]
				require.NotNil(t, user.User.Exec)
				args := user.User.Exec.Args
				assert.Contains(t, args, "--pop-enabled")
				assert.Contains(t, args, "--pop-claims")
			},
		},

		// ===== Error Cases =====
		{
			name:           "fail with invalid login method",
			fixture:        "sample_kubeconfig.yaml",
			args:           []string{"--login", "invalid-method", "--server-id", "6dae42f8-4368-4678-94ff-3960e28e3630"},
			expectSuccess:  false,
			validateResult: nil,
		},

		// ===== Legacy Mode Tests =====
		{
			name:          "convert to devicecode with legacy flag",
			fixture:       "sample_kubeconfig.yaml",
			args:          []string{"--login", "devicecode", "--tenant-id", "new-tenant-id", "--client-id", "new-client-id", "--server-id", "6dae42f8-4368-4678-94ff-3960e28e3630", "--legacy"},
			expectSuccess: true,
			validateResult: func(t *testing.T, originalContent, resultContent string) {
				var config Kubeconfig
				err := yaml.Unmarshal([]byte(resultContent), &config)
				require.NoError(t, err)

				require.Len(t, config.Users, 1)
				user := config.Users[0]
				require.NotNil(t, user.User.Exec)
				args := user.User.Exec.Args
				assert.Contains(t, args, "--legacy")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test both legacy and unified modes
			modes := []struct {
				name    string
				unified bool
			}{
				{"legacy", false},
				{"unified", true},
			}

			// Store results for cross-mode comparison
			var legacyContent, unifiedContent string

			for _, mode := range modes {
				t.Run(mode.name, func(t *testing.T) {
					env := NewTestEnvironment(t)

					// Load fixture content
					fixtureContent, err := LoadFixture(tc.fixture)
					require.NoError(t, err)

					// Create temp kubeconfig
					kubeconfigPath := env.CreateTempKubeconfig(fixtureContent)

					// Prepare command arguments
					args := []string{"convert-kubeconfig", "--kubeconfig", kubeconfigPath}
					args = append(args, tc.args...)

					// Set environment for mode
					envVars := map[string]string{}
					if !mode.unified {
						envVars["KUBELOGIN_USE_LEGACY_OPTIONS"] = "true"
					}

					// Run conversion
					result := RunKubeloginCommandWithEnv(envVars, args...)

					if tc.expectSuccess {
						assert.Equal(t, 0, result.ExitCode, "Conversion should succeed in %s mode. Stderr: %s", mode.name, result.Stderr)

						// Save converted kubeconfig for manual verification (if enabled)
						if result.ExitCode == 0 && SaveOutputEnabled() {
							resultContent := env.GetKubeconfigContent()
							saveFileName := fmt.Sprintf("converted_%s_%s_mode.yaml",
								strings.ReplaceAll(tc.name, " ", "_"), mode.name)
							err := SaveOutput(saveFileName, resultContent)
							if err == nil {
								t.Logf("Saved converted kubeconfig to: _output/%s", saveFileName)
							}
						}

						// Validate the result
						if tc.validateResult != nil {
							resultContent := env.GetKubeconfigContent()
							tc.validateResult(t, fixtureContent, resultContent)
						}

						// Store result content for cross-mode comparison
						resultContent := env.GetKubeconfigContent()
						if mode.unified {
							unifiedContent = resultContent
						} else {
							legacyContent = resultContent
						}
					} else {
						assert.NotEqual(t, 0, result.ExitCode, "Conversion should fail in %s mode", mode.name)
						t.Logf("Expected failure in %s mode: %s", mode.name, result.Stderr)
					}
				})
			}

			// STRICT BACKWARD COMPATIBILITY CHECK
			// Only compare when both modes succeeded and this is a success test case
			if tc.expectSuccess && legacyContent != "" && unifiedContent != "" {
				// Parse both configs to get detailed error information
				var legacyConfig, unifiedConfig Kubeconfig
				err := yaml.Unmarshal([]byte(legacyContent), &legacyConfig)
				require.NoError(t, err, "Failed to parse legacy config")
				err = yaml.Unmarshal([]byte(unifiedContent), &unifiedConfig)
				require.NoError(t, err, "Failed to parse unified config")

				// Extract exec args for detailed error reporting
				var legacyArgs, unifiedArgs []string
				if len(legacyConfig.Users) > 0 && legacyConfig.Users[0].User.Exec != nil {
					legacyArgs = legacyConfig.Users[0].User.Exec.Args
				}
				if len(unifiedConfig.Users) > 0 && unifiedConfig.Users[0].User.Exec != nil {
					unifiedArgs = unifiedConfig.Users[0].User.Exec.Args
				}

				// FUNCTIONAL COMPATIBILITY CHECK: exec args must be equivalent regardless of order
				// This ensures functional backward compatibility while allowing for ordering differences
				assert.ElementsMatch(t, legacyArgs, unifiedArgs,
					"Legacy and unified modes must produce functionally identical exec arguments.\n"+
						"Test case: %s\n"+
						"Legacy args: %v\n"+
						"Unified args: %v\n"+
						"This indicates a functional difference in the unified options implementation.",
					tc.name, legacyArgs, unifiedArgs)
			}
		})
	}
}

// TestExecConfigGeneration tests the exec config generation specifically
func TestExecConfigGeneration(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		unexpectedArgs []string
	}{
		{
			name: "devicecode with basic args",
			args: []string{"--login", "devicecode", "--tenant-id", "test-tenant", "--client-id", "test-client"},
		},
		{
			name:           "service principal with certificate",
			args:           []string{"--login", "spn", "--tenant-id", "test-tenant", "--client-id", "test-client", "--client-certificate", "/path/to/cert.pem"},
			unexpectedArgs: []string{"--client-secret"}, // Should not include secret when using cert
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test unified mode (since this tests the reflection-based arg building)
			env := NewTestEnvironment(t)

			// Create sample kubeconfig with existing exec config
			fixtureContent, err := LoadFixture("sample_kubeconfig.yaml")
			require.NoError(t, err)
			kubeconfigPath := env.CreateTempKubeconfig(fixtureContent)

			// Add required args
			args := []string{"convert-kubeconfig", "--kubeconfig", kubeconfigPath}
			args = append(args, tc.args...)
			if !contains(tc.args, "--server-id") {
				args = append(args, "--server-id", "test-server")
			}

			// Run with unified options (default behavior)
			result := RunKubeloginCommandWithEnv(
				map[string]string{},
				args...,
			)

			assert.Equal(t, 0, result.ExitCode, "Conversion should succeed")

			// Get result content for validation and optional saving
			resultContent := env.GetKubeconfigContent()

			// Save converted kubeconfig for manual verification (if enabled)
			if SaveOutputEnabled() {
				testNameClean := strings.ReplaceAll(tc.name, " ", "_")
				saveFileName := fmt.Sprintf("exec_config_%s.yaml", testNameClean)
				err := SaveOutput(saveFileName, resultContent)
				if err == nil {
					t.Logf("Saved exec config test result to: _output/%s", saveFileName)
				}
			}

			// Parse the result to check exec args
			var config Kubeconfig
			err = yaml.Unmarshal([]byte(resultContent), &config)
			require.NoError(t, err)

			require.Len(t, config.Users, 1)
			user := config.Users[0]
			require.NotNil(t, user.User.Exec)

			execArgs := user.User.Exec.Args

			// Verify basic structure
			assert.Contains(t, execArgs, "get-token", "Should contain get-token command")

			// Check for login method and other key args
			if contains(tc.args, "--login") {
				loginIndex := indexOf(tc.args, "--login")
				if loginIndex >= 0 && loginIndex+1 < len(tc.args) {
					loginMethod := tc.args[loginIndex+1]
					assert.Contains(t, execArgs, "--login", "Should contain --login flag")
					assert.Contains(t, execArgs, loginMethod, "Should contain login method %s", loginMethod)
				}
			}

			// Check that specified arguments appear in exec config
			for i := 0; i < len(tc.args); i += 2 {
				if i+1 < len(tc.args) && strings.HasPrefix(tc.args[i], "--") {
					flag := tc.args[i]
					assert.Contains(t, execArgs, flag, "Should contain flag %s", flag)
					// Note: We don't check exact position since order may vary between implementations
				}
			}

			// Check unexpected args are not present
			for _, unexpectedArg := range tc.unexpectedArgs {
				assert.NotContains(t, execArgs, unexpectedArg, "Should not contain %s", unexpectedArg)
			}
		})
	}
}

// TestModeBehaviorComparison tests that both modes produce identical results
func TestModeBehaviorComparison(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "devicecode conversion",
			args: []string{"--login", "devicecode", "--tenant-id", "test-tenant", "--client-id", "test-client", "--server-id", "test-server"},
		},
		{
			name: "service principal conversion",
			args: []string{"--login", "spn", "--tenant-id", "test-tenant", "--client-id", "test-client", "--client-secret", "test-secret", "--server-id", "test-server"},
		},
		{
			name: "azurecli conversion",
			args: []string{"--login", "azurecli", "--server-id", "test-server"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create two separate environments for comparison
			legacyEnv := NewTestEnvironment(t)
			unifiedEnv := NewTestEnvironment(t)

			// Load same fixture for both
			fixtureContent, err := LoadFixture("sample_kubeconfig.yaml")
			require.NoError(t, err)

			legacyKubeconfig := legacyEnv.CreateTempKubeconfig(fixtureContent)
			unifiedKubeconfig := unifiedEnv.CreateTempKubeconfig(fixtureContent)

			// Prepare args
			legacyArgs := []string{"convert-kubeconfig", "--kubeconfig", legacyKubeconfig}
			legacyArgs = append(legacyArgs, tc.args...)

			unifiedArgs := []string{"convert-kubeconfig", "--kubeconfig", unifiedKubeconfig}
			unifiedArgs = append(unifiedArgs, tc.args...)

			// Run both modes
			legacyResult := RunKubeloginCommandWithEnv(
				map[string]string{"KUBELOGIN_USE_LEGACY_OPTIONS": "true"},
				legacyArgs...,
			)

			unifiedResult := RunKubeloginCommandWithEnv(
				map[string]string{},
				unifiedArgs...,
			)

			// Both should succeed
			assert.Equal(t, 0, legacyResult.ExitCode, "Legacy mode should succeed")
			assert.Equal(t, 0, unifiedResult.ExitCode, "Unified mode should succeed")

			// Save converted kubeconfigs for manual verification (if enabled)
			if legacyResult.ExitCode == 0 && unifiedResult.ExitCode == 0 && SaveOutputEnabled() {
				legacyContent := legacyEnv.GetKubeconfigContent()
				unifiedContent := unifiedEnv.GetKubeconfigContent()

				testNameClean := strings.ReplaceAll(tc.name, " ", "_")
				SaveOutput(fmt.Sprintf("comparison_%s_legacy.yaml", testNameClean), legacyContent)
				SaveOutput(fmt.Sprintf("comparison_%s_unified.yaml", testNameClean), unifiedContent)
				t.Logf("Saved comparison files: _output/comparison_%s_legacy.yaml and _output/comparison_%s_unified.yaml", testNameClean, testNameClean)
			}

			// Compare the resulting kubeconfigs
			legacyContent := legacyEnv.GetKubeconfigContent()
			unifiedContent := unifiedEnv.GetKubeconfigContent()

			// Parse both configs for comparison
			var legacyConfig, unifiedConfig Kubeconfig
			err = yaml.Unmarshal([]byte(legacyContent), &legacyConfig)
			require.NoError(t, err)
			err = yaml.Unmarshal([]byte(unifiedContent), &unifiedConfig)
			require.NoError(t, err)

			// Both should have the same exec configuration structure
			require.Len(t, legacyConfig.Users, 1)
			require.Len(t, unifiedConfig.Users, 1)

			legacyExec := legacyConfig.Users[0].User.Exec
			unifiedExec := unifiedConfig.Users[0].User.Exec

			require.NotNil(t, legacyExec)
			require.NotNil(t, unifiedExec)

			// Command should be the same
			assert.Equal(t, legacyExec.Command, unifiedExec.Command)
			assert.Equal(t, legacyExec.APIVersion, unifiedExec.APIVersion)

			// FUNCTIONAL COMPATIBILITY CHECK: exec args must be equivalent regardless of order
			// This ensures functional backward compatibility while allowing for ordering differences
			assert.ElementsMatch(t, legacyExec.Args, unifiedExec.Args,
				"Legacy and unified modes must produce functionally identical exec arguments.\n"+
					"This test detected a functional difference in arguments for %s conversion.\n"+
					"Legacy args: %v\n"+
					"Unified args: %v",
				tc.name, legacyExec.Args, unifiedExec.Args)

			// Args should contain the same essential elements
			// (order might differ due to reflection vs manual processing)
			for i := 0; i < len(legacyExec.Args); i += 2 {
				if i+1 < len(legacyExec.Args) {
					flag := legacyExec.Args[i]
					value := legacyExec.Args[i+1]

					if strings.HasPrefix(flag, "--") {
						// Find this flag in unified args
						unifiedIndex := indexOf(unifiedExec.Args, flag)
						if unifiedIndex >= 0 && unifiedIndex+1 < len(unifiedExec.Args) {
							assert.Equal(t, value, unifiedExec.Args[unifiedIndex+1],
								"Flag %s should have same value in both modes", flag)
						}
					}
				}
			}
		})
	}
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}

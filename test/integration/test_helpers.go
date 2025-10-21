package integration

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SaveOutputEnabled returns true if output saving is enabled via environment variable
func SaveOutputEnabled() bool {
	return os.Getenv("KUBELOGIN_SAVE_TEST_OUTPUT") == "true"
}

// TestEnvironment manages test isolation and cleanup
type TestEnvironment struct {
	t              *testing.T
	originalEnv    map[string]string
	tempDir        string
	kubeconfigPath string
}

// NewTestEnvironment creates an isolated test environment
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	tempDir, err := os.MkdirTemp("", "kubelogin-integration-test")
	require.NoError(t, err)

	env := &TestEnvironment{
		t:           t,
		originalEnv: make(map[string]string),
		tempDir:     tempDir,
	}

	t.Cleanup(env.Cleanup)
	return env
}

// SetEnv sets an environment variable and remembers the original value
func (te *TestEnvironment) SetEnv(key, value string) {
	if _, exists := te.originalEnv[key]; !exists {
		te.originalEnv[key] = os.Getenv(key)
	}
	os.Setenv(key, value)
}

// ClearEnv clears an environment variable and remembers the original value
func (te *TestEnvironment) ClearEnv(key string) {
	if _, exists := te.originalEnv[key]; !exists {
		te.originalEnv[key] = os.Getenv(key)
	}
	os.Unsetenv(key)
}

// CreateTempKubeconfig creates a temporary kubeconfig file
func (te *TestEnvironment) CreateTempKubeconfig(content string) string {
	kubeconfigPath := filepath.Join(te.tempDir, "kubeconfig")
	err := os.WriteFile(kubeconfigPath, []byte(content), 0644)
	require.NoError(te.t, err)
	te.kubeconfigPath = kubeconfigPath
	return kubeconfigPath
}

// GetKubeconfigContent reads the current kubeconfig content
func (te *TestEnvironment) GetKubeconfigContent() string {
	if te.kubeconfigPath == "" {
		return ""
	}
	content, err := os.ReadFile(te.kubeconfigPath)
	require.NoError(te.t, err)
	return string(content)
}

// Cleanup restores the original environment
func (te *TestEnvironment) Cleanup() {
	// Restore environment variables
	for key, value := range te.originalEnv {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}

	// Clean up temp directory
	os.RemoveAll(te.tempDir)
}

// CommandResult holds the result of running a command
type CommandResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Err      error
}

// RunKubeloginCommand runs a kubelogin command and returns the result
func RunKubeloginCommand(args ...string) *CommandResult {
	// Use relative path to kubelogin binary from integration test directory
	cmd := exec.Command("../../kubelogin", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &CommandResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}

	if exitError, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitError.ExitCode()
	} else if err == nil {
		result.ExitCode = 0
	} else {
		result.ExitCode = -1
	}

	return result
}

// RunKubeloginCommandWithEnv runs a kubelogin command with specific environment variables
func RunKubeloginCommandWithEnv(env map[string]string, args ...string) *CommandResult {
	// Use relative path to kubelogin binary from integration test directory
	cmd := exec.Command("../../kubelogin", args...)

	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &CommandResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}

	if exitError, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitError.ExitCode()
	} else if err == nil {
		result.ExitCode = 0
	} else {
		result.ExitCode = -1
	}

	return result
}

// CompareHelpOutput compares help output between legacy and unified modes
func CompareHelpOutput(t *testing.T, command string) {
	// Get legacy help output
	legacyResult := RunKubeloginCommandWithEnv(
		map[string]string{"KUBELOGIN_USE_LEGACY_OPTIONS": "true"},
		command, "--help",
	)
	require.Equal(t, 0, legacyResult.ExitCode, "Legacy help command should succeed")

	// Get unified help output (default behavior, no env var needed)
	unifiedResult := RunKubeloginCommandWithEnv(
		map[string]string{},
		command, "--help",
	)
	require.Equal(t, 0, unifiedResult.ExitCode, "Unified help command should succeed")

	// Compare flag availability (both should have same flags)
	legacyFlags := extractFlags(legacyResult.Stdout)
	unifiedFlags := extractFlags(unifiedResult.Stdout)

	assert.Equal(t, legacyFlags, unifiedFlags, "Both modes should have identical flags")
}

// extractFlags extracts flag names from help output
func extractFlags(helpOutput string) []string {
	var flags []string
	lines := strings.Split(helpOutput, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-") {
			// Extract flag name (handle both -f and --flag formats)
			parts := strings.Fields(line)
			if len(parts) > 0 {
				flag := parts[0]
				// Remove trailing comma if present
				flag = strings.TrimSuffix(flag, ",")
				flags = append(flags, flag)
			}
		}
	}

	return flags
}

// WaitForFileChange waits for a file to be modified (useful for async operations)
func WaitForFileChange(filePath string, timeout time.Duration) error {
	initialStat, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		currentStat, err := os.Stat(filePath)
		if err != nil {
			return err
		}

		if currentStat.ModTime() != initialStat.ModTime() {
			return nil
		}

		time.Sleep(10 * time.Millisecond)
	}

	return fmt.Errorf("file %s was not modified within %v", filePath, timeout)
}

// LoadFixture loads a test fixture file
func LoadFixture(filename string) (string, error) {
	fixturePath := filepath.Join("fixtures", filename)
	content, err := os.ReadFile(fixturePath)
	if err != nil {
		return "", fmt.Errorf("failed to load fixture %s: %w", filename, err)
	}
	return string(content), nil
}

// SaveFixture saves content to a fixture file (useful for creating expected outputs)
func SaveFixture(filename string, content string) error {
	fixturePath := filepath.Join("fixtures", filename)
	err := os.WriteFile(fixturePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to save fixture %s: %w", filename, err)
	}
	return nil
}

// SaveOutput saves content to the output directory for manual verification
// Only saves if KUBELOGIN_SAVE_TEST_OUTPUT environment variable is set to "true"
func SaveOutput(filename string, content string) error {
	if !SaveOutputEnabled() {
		return nil // Skip saving if not enabled
	}

	outputPath := filepath.Join("convert", "_output", filename)

	// Ensure the output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	err := os.WriteFile(outputPath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to save output %s: %w", filename, err)
	}
	return nil
}

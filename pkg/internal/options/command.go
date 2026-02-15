package options

import (
	"os"

	"github.com/spf13/cobra"
)

const (
	// Environment variable name for falling back to legacy options
	legacyOptionsEnvVar = "KUBELOGIN_USE_LEGACY_OPTIONS"
	// String value for true comparison
	trueValue = "true"
)

// NewUnifiedCommand creates a new cobra command using the unified options pattern
func NewUnifiedCommand(cmdType CommandType) *cobra.Command {
	opts := NewUnifiedOptions(cmdType)

	cmd := &cobra.Command{
		Use:               getCommandUse(cmdType),
		Short:             getCommandShort(cmdType),
		SilenceUsage:      true,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(c *cobra.Command, args []string) error {
			return opts.ExecuteCommand(c.Context(), c.Flags())
		},
	}

	// Auto-register flags using reflection and struct tags
	if err := opts.RegisterFlags(cmd.Flags()); err != nil {
		// This should not happen in practice, but we'll handle it gracefully
		panic("Failed to register flags: " + err.Error())
	}

	// Register completions
	if err := opts.RegisterCompletions(cmd); err != nil {
		// This should not happen in practice, but we'll handle it gracefully
		panic("Failed to register completions: " + err.Error())
	}

	return cmd
}

// getCommandUse returns the command use string based on command type
func getCommandUse(cmdType CommandType) string {
	switch cmdType {
	case ConvertCommand:
		return "convert-kubeconfig"
	case TokenCommand:
		return "get-token"
	default:
		return "unknown"
	}
}

// getCommandShort returns the command short description based on command type
func getCommandShort(cmdType CommandType) string {
	switch cmdType {
	case ConvertCommand:
		return "convert kubeconfig to use exec auth module"
	case TokenCommand:
		return "get AAD token"
	default:
		return "unknown command"
	}
}

// UseUnifiedOptions returns true by default, false only if legacy options are explicitly requested
// This can be controlled by environment variable for gradual rollout
func UseUnifiedOptions() bool {
	return os.Getenv(legacyOptionsEnvVar) != trueValue
}

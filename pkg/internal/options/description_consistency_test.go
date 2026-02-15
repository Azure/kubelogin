package options

import (
	"strings"
	"testing"

	"github.com/Azure/kubelogin/pkg/internal/converter"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFlagDescriptionConsistency validates that unified and legacy modes have consistent flag descriptions
func TestFlagDescriptionConsistency(t *testing.T) {
	// Get legacy flag descriptions
	legacyOptions := converter.New()
	legacyFlags := pflag.NewFlagSet("legacy", pflag.ContinueOnError)
	legacyOptions.AddFlags(legacyFlags)

	// Get unified flag descriptions
	unifiedOptions := NewUnifiedOptions(ConvertCommand)
	unifiedFlags := pflag.NewFlagSet("unified", pflag.ContinueOnError)
	unifiedOptions.RegisterFlags(unifiedFlags)

	// Get all flags from both flag sets
	var legacyFlagNames []string
	var unifiedFlagNames []string
	
	legacyFlags.VisitAll(func(f *pflag.Flag) {
		legacyFlagNames = append(legacyFlagNames, f.Name)
	})
	
	unifiedFlags.VisitAll(func(f *pflag.Flag) {
		unifiedFlagNames = append(unifiedFlagNames, f.Name)
	})

	// Check that both flag sets have the same flags
	t.Run("flag_sets_should_have_same_flags", func(t *testing.T) {
		assert.ElementsMatch(t, legacyFlagNames, unifiedFlagNames,
			"Legacy and unified modes should have the exact same set of flags.\nLegacy: %v\nUnified: %v",
			legacyFlagNames, unifiedFlagNames)
	})

	var inconsistencies []string

	// Compare each flag that exists in both sets
	for _, flagName := range legacyFlagNames {
		t.Run("flag_"+flagName, func(t *testing.T) {
			// Get legacy description
			legacyFlagObj := legacyFlags.Lookup(flagName)
			require.NotNil(t, legacyFlagObj, "Legacy flag %s should exist", flagName)

			// Get unified description  
			unifiedFlagObj := unifiedFlags.Lookup(flagName)
			if unifiedFlagObj == nil {
				t.Errorf("Unified mode is missing flag %s that exists in legacy mode", flagName)
				return
			}

			legacyDesc := legacyFlagObj.Usage
			unifiedDesc := unifiedFlagObj.Usage

			// Assert exact description match
			if legacyDesc != unifiedDesc {
				inconsistencies = append(inconsistencies, 
					"Flag: "+flagName+"\n"+
					"  Legacy:  "+legacyDesc+"\n"+
					"  Unified: "+unifiedDesc+"\n")
				
				t.Errorf("Description mismatch for flag %s:\nLegacy:  %s\nUnified: %s", 
					flagName, legacyDesc, unifiedDesc)
			}
		})
	}

	// Check for any flags that exist only in unified mode
	for _, flagName := range unifiedFlagNames {
		if legacyFlags.Lookup(flagName) == nil {
			t.Errorf("Unified mode has extra flag %s that doesn't exist in legacy mode", flagName)
		}
	}

	// Summary logging (for informational purposes)
	if len(inconsistencies) > 0 {
		t.Logf("Found %d flag description inconsistencies:\n%s", 
			len(inconsistencies), strings.Join(inconsistencies, "\n"))
	} else {
		t.Logf("All %d flags have consistent descriptions between legacy and unified modes", len(legacyFlagNames))
	}
}

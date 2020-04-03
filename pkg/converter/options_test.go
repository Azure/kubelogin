package converter

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestOptions(t *testing.T) {
	o := New()
	flags := &pflag.FlagSet{}
	flags.Set("login", "devicecode")
	o.AddFlags(flags)
	o.UpdateFromEnv()
	if err := o.Validate(); err != nil {
		t.Fatalf("option validation failed: %s", err)
	}
}

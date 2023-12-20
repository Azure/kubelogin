package converter

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestOptions(t *testing.T) {
	o := New()
	o.AddFlags(&pflag.FlagSet{})
	o.UpdateFromEnv()
	if err := o.Validate(); err != nil {
		t.Fatalf("option validation failed: %s", err)
	}
}

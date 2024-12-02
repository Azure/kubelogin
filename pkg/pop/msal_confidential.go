package pop

import (
	"github.com/Azure/kubelogin/pkg/internal/pop"
)

// AcquirePoPTokenConfidential retrieves a Proof of Possession (PoP) token using confidential client credentials.
// It utilizes the internal pop.AcquirePoPTokenConfidential function to obtain the token.
var AcquirePoPTokenConfidential = pop.AcquirePoPTokenConfidential

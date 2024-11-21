package pop

import "github.com/Azure/kubelogin/pkg/internal/pop"

// GetSwPoPKey retrieves a software Proof of Possession (PoP) key using RSA encryption.
// It utilizes the internal pop.GetSwPoPKey function to obtain the key.
var GetSwPoPKey = pop.GetSwPoPKey

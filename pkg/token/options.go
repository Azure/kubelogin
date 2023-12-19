package token

import "github.com/Azure/kubelogin/pkg/internal/token"

// list of supported login methods for library consumers

const (
	ServicePrincipalLogin = token.ServicePrincipalLogin
	ROPCLogin             = token.ROPCLogin
	MSILogin              = token.MSILogin
	WorkloadIdentityLogin = token.WorkloadIdentityLogin
)

// Options defines the options for getting token.
type Options struct {
	LoginMethod string

	// shared login settings

	Environment string
	TenantID    string
	ServerID    string
	ClientID    string

	// for ServicePrincipalLogin & ROPCLogin

	ClientSecret       string
	ClientCert         string
	ClientCertPassword string
	IsPopTokenEnabled  bool
	PoPTokenClaims     string

	// for ROPCLogin
	Username string
	Password string

	// for MSILogin

	IdentityResourceID string

	// for WorkloadIdentityLogin

	AuthorityHost      string
	FederatedTokenFile string
}

package token

import "github.com/Azure/kubelogin/pkg/internal/token"

// list of supported login methods for library consumers

const (
	ServicePrincipalLogin = token.ServicePrincipalLogin
	MSILogin              = token.MSILogin
	WorkloadIdentityLogin = token.WorkloadIdentityLogin
)

// Options defines the options for getting token.
// This struct is a subset of internal/token.Options where its values are copied
// to internal type. See internal/token/options.go for details
type Options struct {
	LoginMethod string

	// shared login settings

	Environment string
	TenantID    string
	ServerID    string
	ClientID    string

	// for ServicePrincipalLogin

	ClientSecret       string
	ClientCert         string
	ClientCertPassword string
	IsPoPTokenEnabled  bool
	PoPTokenClaims     string

	// for MSILogin

	IdentityResourceID string

	// for WorkloadIdentityLogin

	AuthorityHost      string
	FederatedTokenFile string
}

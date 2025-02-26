package token

//go:generate sh -c "mockgen -destination mock_$GOPACKAGE/provider.go github.com/Azure/kubelogin/pkg/internal/token CredentialProvider"

import (
	"context"
	"errors"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type CredentialProvider interface {
	GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error)

	Authenticate(ctx context.Context, options *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error)

	NeedAuthenticate() bool

	Name() string
}

func NewAzIdentityCredential(record azidentity.AuthenticationRecord, o *Options) (CredentialProvider, error) {
	switch o.LoginMethod {
	case AzureCLILogin:
		return newAzureCLICredential(o)

	case AzureDeveloperCLILogin:
		return newAzureDeveloperCLICredential(o)

	case DeviceCodeLogin:
		switch {
		case o.IsLegacy:
			return newADALDeviceCodeCredential(o)
		default:
			return newDeviceCodeCredential(o, record)
		}

	case InteractiveLogin:
		switch {
		case o.IsPoPTokenEnabled:
			return newInteractiveBrowserCredentialWithPoP(o)
		default:
			return newInteractiveBrowserCredential(o, record)
		}

	case MSILogin:
		return newManagedIdentityCredential(o)

	case ROPCLogin:
		switch {
		case o.IsPoPTokenEnabled:
			return newUsernamePasswordCredentialWithPoP(o)
		default:
			return newUsernamePasswordCredential(o, record)
		}

	case ServicePrincipalLogin:
		switch {
		case o.IsLegacy && o.ClientCert != "":
			return newADALClientCertCredential(o)
		case o.IsLegacy:
			return newADALClientSecretCredential(o)
		case o.ClientCert != "" && o.IsPoPTokenEnabled:
			return newClientCertificateCredentialWithPoP(o)
		case o.ClientCert != "":
			return newClientCertificateCredential(o)
		case o.IsPoPTokenEnabled:
			return newClientSecretCredentialWithPoP(o)
		default:
			return newClientSecretCredential(o)
		}

	case WorkloadIdentityLogin:
		switch {
		case os.Getenv(actionsIDTokenRequestToken) != "" && os.Getenv(actionsIDTokenRequestURL) != "":
			return newGithubActionsCredential(o)
		default:
			return newWorkloadIdentityCredential(o)
		}
	}

	return nil, errors.New("unsupported token provider")
}

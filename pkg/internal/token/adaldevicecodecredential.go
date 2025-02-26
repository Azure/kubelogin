package token

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
)

type ADALDeviceCodeCredential struct {
	oAuthConfig adal.OAuthConfig
	clientID    string
}

var _ CredentialProvider = (*ADALDeviceCodeCredential)(nil)

func newADALDeviceCodeCredential(opts *Options) (CredentialProvider, error) {
	if !opts.IsLegacy {
		return nil, fmt.Errorf("ADALDeviceCodeCredential is not supported in non-legacy mode")
	}
	if opts.ClientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}
	if opts.TenantID == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	cloud := opts.GetCloudConfiguration()
	oAuthConfig, err := adal.NewOAuthConfig(cloud.ActiveDirectoryAuthorityHost, opts.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth config: %w", err)
	}
	return &ADALDeviceCodeCredential{
		oAuthConfig: *oAuthConfig,
		clientID:    opts.ClientID,
	}, nil
}

func (c *ADALDeviceCodeCredential) Name() string {
	return "ADALDeviceCodeCredential"
}

func (c *ADALDeviceCodeCredential) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	return azidentity.AuthenticationRecord{}, errAuthenticateNotSupported
}

func (c *ADALDeviceCodeCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	client := &autorest.Client{}
	// to keep backward compatibility,
	// 1. we only support one resource
	// 2. we remove the "/.default" suffix from the resource
	resource := strings.Replace(opts.Scopes[0], "/.default", "", 1)
	deviceCode, err := adal.InitiateDeviceAuth(client, c.oAuthConfig, c.clientID, resource)
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("initialing the device code authentication: %w", err)
	}

	if _, err := fmt.Fprintln(os.Stderr, *deviceCode.Message); err != nil {
		return azcore.AccessToken{}, fmt.Errorf("prompting the device code message: %w", err)
	}

	token, err := adal.WaitForUserCompletionWithContext(ctx, client, deviceCode)
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("waiting for device code authentication to complete: %w", err)
	}

	return azcore.AccessToken{Token: token.AccessToken, ExpiresOn: token.Expires()}, nil
}

func (c *ADALDeviceCodeCredential) NeedAuthenticate() bool {
	return false
}

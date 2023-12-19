package token

import (
	"testing"

	"github.com/Azure/kubelogin/pkg/internal/testutils"
	"github.com/google/go-cmp/cmp"
)

func TestNewTokenProvider(t *testing.T) {
	t.Run("NewTokenProvider should return error on failure to get oAuthConfig", func(t *testing.T) {
		options := &Options{
			Environment: "badenvironment",
			TenantID:    "testtenant",
			IsLegacy:    false,
		}
		provider, err := NewTokenProvider(options)
		if err == nil || provider != nil {
			t.Errorf("expected error but got nil")
		}
		if !testutils.ErrorContains(err, "autorest/azure: There is no cloud environment matching the name") {
			t.Errorf("expected error getting environment but got: %s", err)
		}
	})

	t.Run("NewTokenProvider should return error on failure to parse PoP claims", func(t *testing.T) {
		options := &Options{
			TenantID:          "testtenant",
			IsLegacy:          false,
			IsPoPTokenEnabled: true,
			PoPTokenClaims:    "1=2",
		}
		provider, err := NewTokenProvider(options)
		if err == nil || provider != nil {
			t.Errorf("expected error but got nil")
		}
		if !testutils.ErrorContains(err, "required u-claim not provided for PoP token flow") {
			t.Errorf("expected error parsing PoP claims but got: %s", err)
		}
	})

	t.Run("NewTokenProvider should return error on invalid login method", func(t *testing.T) {
		options := &Options{
			LoginMethod: "unsupported",
		}
		provider, err := NewTokenProvider(options)
		if err == nil || provider != nil {
			t.Errorf("expected error but got nil")
		}
		if !testutils.ErrorContains(err, "unsupported token provider") {
			t.Errorf("expected unsupported token provider error but got: %s", err)
		}
	})

	t.Run("NewTokenProvider should return interactive token provider with correct fields", func(t *testing.T) {
		options := &Options{
			TenantID:          "testtenant",
			ClientID:          "testclient",
			ServerID:          "testserver",
			IsPoPTokenEnabled: true,
			PoPTokenClaims:    "u=testhost",
			LoginMethod:       "interactive",
		}
		provider, err := NewTokenProvider(options)
		if err != nil || provider == nil {
			t.Errorf("expected no error but got: %s", err)
		}
		interactive := provider.(*InteractiveToken)
		if interactive.clientID != options.ClientID {
			t.Errorf("expected provider client ID to be: %s but got: %s", options.ClientID, interactive.clientID)
		}
		if interactive.resourceID != options.ServerID {
			t.Errorf("expected provider resource ID to be: %s but got: %s", options.ServerID, interactive.resourceID)
		}
		if interactive.tenantID != options.TenantID {
			t.Errorf("expected provider tenant ID to be: %s but got: %s", options.TenantID, interactive.tenantID)
		}
		expectedPoPClaims := map[string]string{"u": "testhost"}
		if !cmp.Equal(interactive.popClaims, expectedPoPClaims) {
			t.Errorf("expected provider PoP claims to be: %v but got: %v", expectedPoPClaims, interactive.popClaims)
		}
	})

	t.Run("NewTokenProvider should return SPN token provider using client secret with correct fields", func(t *testing.T) {
		options := &Options{
			TenantID:          "testtenant",
			ClientID:          "testclient",
			ServerID:          "testserver",
			ClientSecret:      "testsecret",
			IsPoPTokenEnabled: true,
			PoPTokenClaims:    "u=testhost, 1=2",
			LoginMethod:       "spn",
		}
		provider, err := NewTokenProvider(options)
		if err != nil || provider == nil {
			t.Errorf("expected no error but got: %s", err)
		}
		spn := provider.(*servicePrincipalToken)
		if spn.clientID != options.ClientID {
			t.Errorf("expected provider client ID to be: %s but got: %s", options.ClientID, spn.clientID)
		}
		if spn.resourceID != options.ServerID {
			t.Errorf("expected provider resource ID to be: %s but got: %s", options.ServerID, spn.resourceID)
		}
		if spn.tenantID != options.TenantID {
			t.Errorf("expected provider tenant ID to be: %s but got: %s", options.TenantID, spn.tenantID)
		}
		if spn.clientSecret != options.ClientSecret {
			t.Errorf("expected provider client secret to be: %s but got: %s", options.ClientSecret, spn.clientSecret)
		}
		expectedPoPClaims := map[string]string{"u": "testhost", "1": "2"}
		if !cmp.Equal(spn.popClaims, expectedPoPClaims) {
			t.Errorf("expected provider PoP claims to be: %v but got: %v", expectedPoPClaims, spn.popClaims)
		}
	})

	t.Run("NewTokenProvider should return SPN token provider using client cert with correct fields", func(t *testing.T) {
		options := &Options{
			TenantID:           "testtenant",
			ClientID:           "testclient",
			ServerID:           "testserver",
			ClientCert:         "testcert",
			ClientCertPassword: "testcertpass",
			LoginMethod:        "spn",
		}
		provider, err := NewTokenProvider(options)
		if err != nil || provider == nil {
			t.Errorf("expected no error but got: %s", err)
		}
		spn := provider.(*servicePrincipalToken)
		if spn.clientID != options.ClientID {
			t.Errorf("expected provider client ID to be: %s but got: %s", options.ClientID, spn.clientID)
		}
		if spn.resourceID != options.ServerID {
			t.Errorf("expected provider resource ID to be: %s but got: %s", options.ServerID, spn.resourceID)
		}
		if spn.tenantID != options.TenantID {
			t.Errorf("expected provider tenant ID to be: %s but got: %s", options.TenantID, spn.tenantID)
		}
		if spn.clientCert != options.ClientCert {
			t.Errorf("expected provider client cert to be: %s but got: %s", options.ClientCert, spn.clientCert)
		}
		if spn.clientCertPassword != options.ClientCertPassword {
			t.Errorf("expected provider client cert password to be: %s but got: %s", options.ClientCertPassword, spn.clientCertPassword)
		}
		if spn.popClaims != nil {
			t.Errorf("expected provider PoP claims to be nil but got: %v", spn.popClaims)
		}
	})

	t.Run("NewTokenProvider should return resource owner token provider with correct fields", func(t *testing.T) {
		options := &Options{
			TenantID:    "testtenant",
			ClientID:    "testclient",
			ServerID:    "testserver",
			Username:    "testuser",
			Password:    "testpass",
			LoginMethod: "ropc",
		}
		provider, err := NewTokenProvider(options)
		if err != nil || provider == nil {
			t.Errorf("expected no error but got: %s", err)
		}
		ropc := provider.(*resourceOwnerToken)
		if ropc.clientID != options.ClientID {
			t.Errorf("expected provider client ID to be: %s but got: %s", options.ClientID, ropc.clientID)
		}
		if ropc.resourceID != options.ServerID {
			t.Errorf("expected provider resource ID to be: %s but got: %s", options.ServerID, ropc.resourceID)
		}
		if ropc.tenantID != options.TenantID {
			t.Errorf("expected provider tenant ID to be: %s but got: %s", options.TenantID, ropc.tenantID)
		}
		if ropc.username != options.Username {
			t.Errorf("expected provider username to be: %s but got: %s", options.Username, ropc.username)
		}
		if ropc.password != options.Password {
			t.Errorf("expected provider password to be: %s but got: %s", options.Password, ropc.password)
		}
	})

	t.Run("NewTokenProvider should return resource owner token provider with correct fields", func(t *testing.T) {
		options := &Options{
			ClientID:           "testclient",
			ServerID:           "testserver",
			IdentityResourceID: "testidentity",
			LoginMethod:        "msi",
		}
		provider, err := NewTokenProvider(options)
		if err != nil || provider == nil {
			t.Errorf("expected no error but got: %s", err)
		}
		msi := provider.(*managedIdentityToken)
		if msi.clientID != options.ClientID {
			t.Errorf("expected provider client ID to be: %s but got: %s", options.ClientID, msi.clientID)
		}
		if msi.resourceID != options.ServerID {
			t.Errorf("expected provider resource ID to be: %s but got: %s", options.ServerID, msi.resourceID)
		}
		if msi.identityResourceID != options.IdentityResourceID {
			t.Errorf("expected provider identity resource ID to be: %s but got: %s", options.IdentityResourceID, msi.identityResourceID)
		}
	})

	t.Run("NewTokenProvider should return azure CLI token provider with correct fields", func(t *testing.T) {
		options := &Options{
			ServerID:    "testserver",
			TenantID:    "testtenant",
			LoginMethod: "azurecli",
		}
		provider, err := NewTokenProvider(options)
		if err != nil || provider == nil {
			t.Errorf("expected no error but got: %s", err)
		}
		msi := provider.(*AzureCLIToken)
		if msi.tenantID != options.TenantID {
			t.Errorf("expected provider tenant ID to be: %s but got: %s", options.TenantID, msi.tenantID)
		}
		if msi.resourceID != options.ServerID {
			t.Errorf("expected provider resource ID to be: %s but got: %s", options.ServerID, msi.resourceID)
		}
	})

	t.Run("NewTokenProvider should return workload identity token provider with correct fields", func(t *testing.T) {
		options := &Options{
			TenantID:           "testtenant",
			ClientID:           "testclient",
			ServerID:           "testserver",
			FederatedTokenFile: "testfile",
			AuthorityHost:      "https://testauthority",
			LoginMethod:        "workloadidentity",
		}
		t.Run("with token file", func(t *testing.T) {
			provider, err := NewTokenProvider(options)
			if err != nil || provider == nil {
				t.Errorf("expected no error but got: %s", err)
			}
			workloadId := provider.(*workloadIdentityToken)
			if workloadId.serverID != options.ServerID {
				t.Errorf("expected provider server ID to be: %s but got: %s", options.ServerID, workloadId.serverID)
			}
		})
		t.Run("with Github token", func(t *testing.T) {
			options.FederatedTokenFile = ""
			t.Setenv(actionsIDTokenRequestToken, "fake-token")
			t.Setenv(actionsIDTokenRequestURL, "fake-url")
			provider, err := NewTokenProvider(options)
			if err != nil || provider == nil {
				t.Errorf("expected no error but got: %s", err)
			}
		})
	})
}

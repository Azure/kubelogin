package token

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/kubelogin/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestWorkloadIdentityCredential_GetToken(t *testing.T) {
	var tokenFile string

	liveTokenFile := os.Getenv("KUBELOGIN_LIVETEST_FEDERATED_TOKEN_FILE")
	if liveTokenFile == "" {
		tempDir, err := os.MkdirTemp("", "kubelogin")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}

		tokenFile = filepath.Join(tempDir, "token")
		outFile, err := os.Create(tokenFile)
		if err != nil {
			t.Fatalf("failed to create token file: %v", err)
		}
		_, err = outFile.WriteString("[REDACTED]")
		if err != nil {
			t.Fatalf("failed to write token file: %v", err)
		}
		outFile.Close()
	} else {
		tokenFile = liveTokenFile
	}

	rec, err := testutils.GetVCRHttpClient("fixtures/workloadidentity_credential", testutils.TestTenantID)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}
	defer rec.Stop()

	opts := &Options{
		ClientID:           testutils.TestClientID,
		ServerID:           testutils.TestServerID,
		TenantID:           testutils.TestTenantID,
		FederatedTokenFile: tokenFile,
		httpClient:         rec.GetDefaultClient(),
	}
	cred, err := newWorkloadIdentityCredential(opts)
	if err != nil {
		t.Fatalf("failed to create credential: %v", err)
	}

	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{opts.ServerID + "/.default"},
	})
	assert.NoError(t, err)
	assert.Equal(t, testutils.TestToken, token.Token)
}

func Test_newWorkloadIdentityCredential(t *testing.T) {
	type args struct {
		opts *Options
	}
	tests := []struct {
		name    string
		args    args
		want    CredentialProvider
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newWorkloadIdentityCredential(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("newWorkloadIdentityCredential() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newWorkloadIdentityCredential() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorkloadIdentityCredential_Name(t *testing.T) {
	tests := []struct {
		name string
		c    *WorkloadIdentityCredential
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Name(); got != tt.want {
				t.Errorf("WorkloadIdentityCredential.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorkloadIdentityCredential_Authenticate(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts *policy.TokenRequestOptions
	}
	tests := []struct {
		name    string
		c       *WorkloadIdentityCredential
		args    args
		want    azidentity.AuthenticationRecord
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Authenticate(tt.args.ctx, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("WorkloadIdentityCredential.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WorkloadIdentityCredential.Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorkloadIdentityCredential_NeedAuthenticate(t *testing.T) {
	tests := []struct {
		name string
		c    *WorkloadIdentityCredential
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.NeedAuthenticate(); got != tt.want {
				t.Errorf("WorkloadIdentityCredential.NeedAuthenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

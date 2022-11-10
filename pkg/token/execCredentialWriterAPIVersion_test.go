package token

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/Azure/go-autorest/autorest/adal"
	"k8s.io/client-go/pkg/apis/clientauthentication"
)

func TestExecCredentialWriterAPIVersion(t *testing.T) {
	testData := []struct {
		name               string
		execInfoEnvTest    string
		expectedAPIVersion string
	}{
		{
			name:               "KUBERNETES_EXEC_INFO is empty",
			execInfoEnvTest:    "",
			expectedAPIVersion: "client.authentication.k8s.io/v1beta1",
		},
		{
			name:               "KUBERNETES_EXEC_INFO is present and apiVersion is absent",
			execInfoEnvTest:    `{"kind":"ExecCredential","spec":{"interactive":true},"apiVersion":""}`,
			expectedAPIVersion: "client.authentication.k8s.io/v1beta1",
		},
		{
			name:               "KUBERNETES_EXEC_INFO is present and apiVersion is neither v1 or v1beta1",
			execInfoEnvTest:    `{"kind":"ExecCredential","apiVersion":"client.authentication.k8s.io/v1alpha1","spec":{"interactive":true}}`,
			expectedAPIVersion: "",
		},
		{
			name:               "KUBERNETES_EXEC_INFO is present and apiVersion is v1beta1",
			execInfoEnvTest:    `{"kind":"ExecCredential","apiVersion":"client.authentication.k8s.io/v1beta1","spec":{"interactive":true}}`,
			expectedAPIVersion: "client.authentication.k8s.io/v1beta1",
		},
		{
			name:               "KUBERNETES_EXEC_INFO is present and apiVersion is v1",
			execInfoEnvTest:    `{"kind":"ExecCredential","apiVersion":"client.authentication.k8s.io/v1","spec":{"interactive":true}}`,
			expectedAPIVersion: "client.authentication.k8s.io/v1",
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {
			os.Setenv("KUBERNETES_EXEC_INFO", data.execInfoEnvTest)
			defer os.Unsetenv("KUBERNETES_EXEC_INFO")
			ecw := execCredentialWriter{}
			tokenTest := adal.Token{}
			stringBufferTest := new(bytes.Buffer)
			ecw.Write(tokenTest, stringBufferTest)
			var execCredential clientauthentication.ExecCredential
			json.Unmarshal(stringBufferTest.Bytes(), &execCredential)
			if execCredential.TypeMeta.APIVersion != data.expectedAPIVersion {
				t.Fatalf("expected: %s, actual: %s", data.expectedAPIVersion, execCredential.TypeMeta.APIVersion)
			}
		})
	}
}

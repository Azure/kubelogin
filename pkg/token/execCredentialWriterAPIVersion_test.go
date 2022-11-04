package token

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/Azure/go-autorest/autorest/adal"
	"k8s.io/client-go/pkg/apis/clientauthentication"
)


func TestExecCredentialWriterAPIVersion() (t *testing.T) {
	testData1 := []struct {
		name               string
		execInfoEnvTest    string
		stringBufferTest   bytes.Buffer
		expectedAPIVersion string
	}{
		{
			name: "KUBERNETES_EXEC_INFO is empty",
			execInfoEnvTest: "",
			stringBufferTest: bytes.NewBuffer([]byte(""),
			expectedAPIVersion: "client.authentication.k8s.io/v1beta1",
		},
		{
			name: "KUBERNETES_EXEC_INFO is is present and apiVersion is absent",
			execInfoEnvTest: "{"kind":"ExecCredential","apiVersion":"","spec"{"interactive":true}}",
			stringBufferTest: bytes.NewBuffer([]byte(""),
			expectedAPIVersion: "client.authentication.k8s.io/v1beta1",
		},
		{
			name: "KUBERNETES_EXEC_INFO is is present and apiVersion is neither v1 or v1beta1",
			execInfoEnvTest: "{"kind":"ExecCredential","apiVersion":"client.authentication.k8s.io/v1alpha1","spec"{"interactive":true}}",
			stringBufferTest: bytes.NewBuffer([]byte(""),
			expectedAPIVersion: "",
		},
		{
			name: "KUBERNETES_EXEC_INFO is is present and apiVersion is v1beta1",
			execInfoEnvTest: "{"kind":"ExecCredential","apiVersion":"client.authentication.k8s.io/v1beta1","spec"{"interactive":true}}",
			stringBufferTest: bytes.NewBuffer([]byte(""),
			expectedAPIVersion: "client.authentication.k8s.io/v1beta1",
		},
		{
			name: "KUBERNETES_EXEC_INFO is is present and apiVersion is v1",
			execInfoEnvTest: "{"kind":"ExecCredential","apiVersion":"client.authentication.k8s.io/v1","spec"{"interactive":true}}",
			stringBufferTest: bytes.NewBuffer([]byte(""),
			expectedAPIVersion: "client.authentication.k8s.io/v1",
		},

	}

	for _, data := range testData1 {
		t.Run(data.name, func(t *testing.T) {
			os.SetEnv("KUBERNETES_EXEC_INFO", testData.execInfoEnvTest)
			defer os.UnsetEnv("KUBERNETES_EXEC_INFO")
			execCredentialWriter.Write(adal.Token, data.stringBufferTest)
			var execCredential clientauthentication.ExecCredential
			json.Unmarshal([]byte(data.stringBufferTest.String()), &execCredential)
			if !execCredential.TypeMeta.APIVersion.equal(data.expectedAPIVersion) {
				t.Fatalf("expected: %s, actual: %s", data.expectedAPIVersion, execCredential.TypeMeta.APIVersion)
			}
		})
	}
}


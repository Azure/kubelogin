package token

//go:generate sh -c "mockgen -destination mock_$GOPACKAGE/execCredentialWriter.go github.com/Azure/kubelogin/pkg/token ExecCredentialWriter"

import (
	"bytes"
	"encoding/json"
	"fmt"

	//"io"
	"os"

	"github.com/Azure/go-autorest/autorest/adal"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/apis/clientauthentication"
	v1 "k8s.io/client-go/pkg/apis/clientauthentication/v1"
	"k8s.io/client-go/pkg/apis/clientauthentication/v1beta1"
)

const (
	apiV1       string = "client.authentication.k8s.io/v1"
	apiV1beta1  string = "client.authentication.k8s.io/v1beta1"
	execInfoEnv string = "KUBERNETES_EXEC_INFO"
)

type ExecCredentialWriter interface {
	Write(token adal.Token, buffer *bytes.Buffer) error
}

type execCredentialWriter struct{}

// Write writes the ExecCredential to standard output for kubectl.
func (*execCredentialWriter) Write(token adal.Token, buffer *bytes.Buffer) error {
	apiVersionFromEnv, err := getAPIVersionFromExecInfoEnv()
	if err != nil {
		return err
	}

	var ec interface{}
	t := metav1.NewTime(token.Expires())
	switch apiVersionFromEnv {
	case apiV1beta1:
		ec = &v1beta1.ExecCredential{
			TypeMeta: metav1.TypeMeta{
				APIVersion: apiV1beta1,
				Kind:       "ExecCredential",
			},
			Status: &v1beta1.ExecCredentialStatus{
				Token:               token.AccessToken,
				ExpirationTimestamp: &t,
			},
		}
	case apiV1:
		ec = &v1.ExecCredential{
			TypeMeta: metav1.TypeMeta{
				APIVersion: apiV1,
				Kind:       "ExecCredential",
			},
			Status: &v1.ExecCredentialStatus{
				Token:               token.AccessToken,
				ExpirationTimestamp: &t,
			},
		}
	}
	var ecCopy interface{} = ec
	content, _ := json.Marshal(ecCopy)
	//fmt.Fprintln(os.Stderr, string(content))
	buffer.WriteString(string(content))
	//fmt.Fprintln(os.Stderr, buffer.String())
	e := json.NewEncoder(os.Stdout)
	if err := e.Encode(ec); err != nil {
		return fmt.Errorf("could not write the ExecCredential: %s", err)
	}
	return nil
}

func getAPIVersionFromExecInfoEnv() (string, error) {
	env := os.Getenv(execInfoEnv)
	if env == "" {
		return apiV1beta1, nil
	}
	var execCredential clientauthentication.ExecCredential
	error := json.Unmarshal([]byte(env), &execCredential)
	if error != nil {
		return "", fmt.Errorf("cannot unmarshall %q to ExecCredential: %w", env, error)
	}
	switch execCredential.TypeMeta.APIVersion {
	case "":
		return apiV1beta1, nil
	case apiV1, apiV1beta1:
		return execCredential.TypeMeta.APIVersion, nil
	default:
		return "", fmt.Errorf("api version: %s is not supported", execCredential.TypeMeta.APIVersion)
	}
}

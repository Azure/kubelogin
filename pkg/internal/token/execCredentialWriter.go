package token

//go:generate sh -c "mockgen -destination mock_$GOPACKAGE/execCredentialWriter.go github.com/Azure/kubelogin/pkg/internal/token ExecCredentialWriter"

import (
	"encoding/json"
	"fmt"
	"io"
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
	Write(token adal.Token, writer io.Writer) error
}

type execCredentialWriter struct{}

// Write writes the ExecCredential to standard output for kubectl.
func (*execCredentialWriter) Write(token adal.Token, writer io.Writer) error {
	apiVersionFromEnv, err := getAPIVersionFromExecInfoEnv()
	if err != nil {
		return err
	}
	// Support both apiVersions of client.authentication.k8s.io/v1beta1 and client.authentication.k8s.io/v1
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

	e := json.NewEncoder(writer)
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
	if err := json.Unmarshal([]byte(env), &execCredential); err != nil {
		return "", fmt.Errorf("cannot unmarshal %q to ExecCredential: %w", env, err)
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

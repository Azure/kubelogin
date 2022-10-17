package token

//go:generate sh -c "mockgen -destination mock_$GOPACKAGE/execCredentialWriter.go github.com/Azure/kubelogin/pkg/token ExecCredentialWriter"

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/Azure/go-autorest/autorest/adal"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	//"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/pkg/apis/clientauthentication"
	v1 "k8s.io/client-go/pkg/apis/clientauthentication/v1"
	"k8s.io/client-go/pkg/apis/clientauthentication/v1beta1"
)

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)

const execInfoEnv = "KUBERNETES_EXEC_INFO"

type ExecCredentialWriter interface {
	Write(token adal.Token) error
}

type execCredentialWriter struct{}

// Write writes the ExecCredential to standard output for kubectl.
func (*execCredentialWriter) Write(token adal.Token) error {
	fmt.Fprintln(os.Stderr, os.Getenv(execInfoEnv))
	//os.Setenv(execInfoEnv, "TEST")
	//fmt.Println("KUBERNETES_EXEC_INFO:", os.Getenv(execInfoEnv))
	apiVersionFromEnv, err := helperGetApiVersionFromEnv()
	if err != nil {
		return err
	}

	var ec interface{}
	t := metav1.NewTime(token.Expires())
	if apiVersionFromEnv == "client.authentication.k8s.io/v1beta1" {
		ec = &v1beta1.ExecCredential{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "client.authentication.k8s.io/v1beta1",
				Kind:       "ExecCredential",
			},
			Status: &v1beta1.ExecCredentialStatus{
				Token:               token.AccessToken,
				ExpirationTimestamp: &t,
			},
		}
	} else {
		ec = &v1.ExecCredential{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "client.authentication.k8s.io/v1",
				Kind:       "ExecCredential",
			},
			Status: &v1.ExecCredentialStatus{
				Token:               token.AccessToken,
				ExpirationTimestamp: &t,
			},
		}
	}

	e := json.NewEncoder(os.Stdout)
	if err := e.Encode(ec); err != nil {
		return fmt.Errorf("could not write the ExecCredential: %s", err)
	}
	return nil
}

func helperGetApiVersionFromEnv() (string, error) {
	env := os.Getenv(execInfoEnv)
	if env == "" {
		return "client.authentication.k8s.io/v1beta1", nil
	} else {
		obj, _, err := codecs.UniversalDeserializer().Decode([]byte(env), nil, nil)
		if err != nil {
			return "", fmt.Errorf("decode: %w", err)
		}
		var execCredential clientauthentication.ExecCredential
		if err := scheme.Convert(obj, &execCredential, nil); err != nil {
			return "", fmt.Errorf("cannot convert to ExecCredential: %w", err)
		}
		if execCredential.TypeMeta.APIVersion == "" {
			return "client.authentication.k8s.io/v1beta1", nil
		}
		if execCredential.TypeMeta.APIVersion == "client.authentication.k8s.io/v1beta1" || execCredential.TypeMeta.APIVersion == "client.authentication.k8s.io/v1" {
			return execCredential.TypeMeta.APIVersion, nil
		} else {
			return "", errors.New("This api version is not supported")
		}
	}
}

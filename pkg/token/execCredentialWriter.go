package token

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Azure/go-autorest/autorest/adal"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/apis/clientauthentication/v1beta1"
)

type ExecCredentialWriter interface {
	Write(token adal.Token) error
}

type execCredentialWriter struct{}

// Write writes the ExecCredential to standard output for kubectl.
func (*execCredentialWriter) Write(token adal.Token) error {
	t := v1.NewTime(token.Expires())
	ec := &v1beta1.ExecCredential{
		TypeMeta: v1.TypeMeta{
			APIVersion: "client.authentication.k8s.io/v1beta1",
			Kind:       "ExecCredential",
		},
		Status: &v1beta1.ExecCredentialStatus{
			Token:               token.AccessToken,
			ExpirationTimestamp: &t,
		},
	}
	e := json.NewEncoder(os.Stdout)
	if err := e.Encode(ec); err != nil {
		return fmt.Errorf("could not write the ExecCredential: %s", err)
	}
	return nil
}

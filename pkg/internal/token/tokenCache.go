package token

//go:generate sh -c "mockgen -destination mock_$GOPACKAGE/tokenCache.go github.com/Azure/kubelogin/pkg/internal/token TokenCache"

import (
	"os"
	"time"

	"github.com/Azure/go-autorest/autorest/adal"

	"gopkg.in/retry.v1"
)

type TokenCache interface {
	Read(string) (adal.Token, error)
	Write(string, adal.Token) error
}

type defaultTokenCache struct{}

func (*defaultTokenCache) Read(file string) (adal.Token, error) {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return adal.Token{}, nil
	}
	token, err := adal.LoadToken(file)
	if err != nil {
		return adal.Token{}, err
	}

	return *token, nil
}

func (*defaultTokenCache) Write(file string, token adal.Token) error {
	attempts := retry.Regular{
		Total: 1 * time.Second,
		Delay: 250 * time.Millisecond,
	}

	for attempt := attempts.Start(nil); attempt.Next(); {
		err := adal.SaveToken(file, 0700, token)

		if err != nil && attempt.More() {
			continue
		}

		return err
	}
	return nil
}

package token

import (
	"os"

	"github.com/Azure/go-autorest/autorest/adal"
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
	return adal.SaveToken(file, 0700, token)
}

package token

import (
	"encoding/json"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type CachedRecordProvider interface {
	// Retrieve reads the authentication record from the file.
	Retrieve() (azidentity.AuthenticationRecord, error)
	// Store writes the authentication record to the file.
	Store(record azidentity.AuthenticationRecord) error
}

type defaultCachedRecordProvider struct {
	file string
}

func (c *defaultCachedRecordProvider) Retrieve() (azidentity.AuthenticationRecord, error) {
	record := azidentity.AuthenticationRecord{}
	b, err := os.ReadFile(c.file)
	if err == nil {
		err = json.Unmarshal(b, &record)
	}
	return record, err
}

func (c *defaultCachedRecordProvider) Store(record azidentity.AuthenticationRecord) error {
	b, err := json.Marshal(record)
	if err == nil {
		err = os.WriteFile(c.file, b, 0600)
	}
	return err
}

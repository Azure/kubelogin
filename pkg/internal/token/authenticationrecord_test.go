package token

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/stretchr/testify/assert"
)

func TestDefaultCachedRecordProvider(t *testing.T) {
	testCases := []struct {
		name           string
		fileContent    string
		expectErrorMsg string
	}{
		{
			name:        "valid record",
			fileContent: `{"tenantID":"test-tenant-id","clientID":"test-client-id","authority":"https://login.microsoftonline.com/","homeAccountID":"test-home-account-id","username":"test-username","version":"1.0"}`,
		},
		{
			name:           "invalid JSON",
			fileContent:    `invalid-json-content`,
			expectErrorMsg: "invalid character",
		},
		{
			name:           "empty file",
			fileContent:    ``,
			expectErrorMsg: "unexpected end of JSON input",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "test-record-*.json")
			assert.NoError(t, err)
			defer os.Remove(file.Name())

			_, err = file.WriteString(tc.fileContent)
			assert.NoError(t, err)
			file.Close()

			provider := &defaultCachedRecordProvider{file: file.Name()}
			record, err := provider.Retrieve()
			if tc.expectErrorMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
			}
		})
	}

	record := azidentity.AuthenticationRecord{
		TenantID:      "test-tenant-id",
		ClientID:      "test-client-id",
		Authority:     "https://login.microsoftonline.com/",
		HomeAccountID: "test-home-account-id",
		Username:      "test-username",
		Version:       "1.0",
	}

	file, err := os.CreateTemp("", "test-record-*.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	provider := &defaultCachedRecordProvider{file: file.Name()}
	err = provider.Store(record)
	assert.NoError(t, err)

	storedRecord, err := provider.Retrieve()
	assert.NoError(t, err)
	assert.Equal(t, record, storedRecord)
}

func TestDefaultCachedRecordProvider_NonExistentDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-record-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	nonExistentDir := filepath.Join(tempDir, "subdir", "nested")
	filePath := filepath.Join(nonExistentDir, "record.json")

	record := azidentity.AuthenticationRecord{
		TenantID:      "test-tenant-id",
		ClientID:      "test-client-id",
		Authority:     "https://login.microsoftonline.com/",
		HomeAccountID: "test-home-account-id",
		Username:      "test-username",
		Version:       "1.0",
	}

	provider := &defaultCachedRecordProvider{file: filePath}
	err = provider.Store(record)
	assert.NoError(t, err)

	// Verify the file was created and can be read
	storedRecord, err := provider.Retrieve()
	assert.NoError(t, err)
	assert.Equal(t, record, storedRecord)

	// Verify the directory was created with correct permissions
	fileInfo, err := os.Stat(nonExistentDir)
	assert.NoError(t, err)
	assert.True(t, fileInfo.IsDir())
}

func Test_defaultCachedRecordProvider_Retrieve(t *testing.T) {
	tests := []struct {
		name    string
		c       *defaultCachedRecordProvider
		want    azidentity.AuthenticationRecord
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Retrieve()
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultCachedRecordProvider.Retrieve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("defaultCachedRecordProvider.Retrieve() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultCachedRecordProvider_Store(t *testing.T) {
	type args struct {
		record azidentity.AuthenticationRecord
	}
	tests := []struct {
		name    string
		c       *defaultCachedRecordProvider
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.Store(tt.args.record); (err != nil) != tt.wantErr {
				t.Errorf("defaultCachedRecordProvider.Store() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

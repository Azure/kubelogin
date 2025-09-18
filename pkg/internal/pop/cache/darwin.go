//go:build go1.23 && darwin && cgo

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package cache

import (
	"os"

	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache/accessor"
)

// cacheDir returns the cache directory for macOS
func cacheDir() (string, error) {
	return os.UserHomeDir()
}

// storage creates a platform-specific accessor for macOS
func storage(name string) (accessor.Accessor, error) {
	return accessor.New(name, accessor.WithAccount("MSALCache"))
}

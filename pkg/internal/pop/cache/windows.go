//go:build go1.23 && windows

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package cache

import (
	"path/filepath"

	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache/accessor"
	"golang.org/x/sys/windows"
)

// cacheDir returns the cache directory for Windows
func cacheDir() (string, error) {
	return windows.KnownFolderPath(windows.FOLDERID_LocalAppData, 0)
}

// storage creates a platform-specific accessor for Windows
func storage(name string) (accessor.Accessor, error) {
	p, err := cacheFilePath(name)
	if err != nil {
		return nil, err
	}
	return accessor.New(p)
}

// cacheFilePath constructs the cache file path for Windows
func cacheFilePath(name string) (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}

	// Following azidentity pattern: LocalAppData + name
	return filepath.Join(dir, name), nil
}

//go:build go1.23 && linux

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package cache

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache/accessor"
	"golang.org/x/sys/unix"
)

// cacheDir returns the cache directory for Linux
func cacheDir() (string, error) {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return xdg, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".cache"), nil
}

// storage creates a platform-specific accessor for Linux
func storage(name string) (accessor.Accessor, error) {
	// Try kernel keyring first, fallback to file-based storage
	if keyringAccessor, err := newKeyringAccessor(name); err == nil {
		return keyringAccessor, nil
	}

	// Fallback to file-based accessor if keyring fails
	dir, err := cacheDir()
	if err != nil {
		return nil, err
	}

	cachePath := filepath.Join(dir, "kubelogin", "pop", fmt.Sprintf("%s.cache", name))
	if err := os.MkdirAll(filepath.Dir(cachePath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return accessor.New(cachePath)
}

// keyringAccessor implements accessor.Accessor using Linux kernel keyrings
type keyringAccessor struct {
	keyringID int
	keyName   string
}

// newKeyringAccessor creates a new keyring-based accessor
func newKeyringAccessor(name string) (accessor.Accessor, error) {
	// Try to get or create a user session keyring
	keyringID, err := unix.KeyctlGetKeyringID(unix.KEY_SPEC_USER_SESSION_KEYRING, true)
	if err != nil {
		return nil, fmt.Errorf("failed to access user keyring: %w", err)
	}

	return &keyringAccessor{
		keyringID: keyringID,
		keyName:   fmt.Sprintf("kubelogin-pop-%s", name),
	}, nil
}

// Read retrieves data from the kernel keyring
func (k *keyringAccessor) Read(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Search for the key
	keyID, err := unix.KeyctlSearch(k.keyringID, "user", k.keyName, 0)
	if err != nil {
		if err == syscall.ENOKEY {
			return nil, fmt.Errorf("key not found in keyring")
		}
		return nil, fmt.Errorf("keyring search failed: %w", err)
	}

	// Get the key size first
	size, err := unix.KeyctlBuffer(keyID, unix.KEYCTL_READ, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get key size: %w", err)
	}

	// Read the actual data
	data := make([]byte, size)
	_, err = unix.KeyctlBuffer(keyID, unix.KEYCTL_READ, data, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read from keyring: %w", err)
	}

	return data, nil
}

// Write stores data in the kernel keyring
func (k *keyringAccessor) Write(ctx context.Context, data []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Add the key to the keyring
	_, err := unix.AddKey("user", k.keyName, data, k.keyringID)
	if err != nil {
		return fmt.Errorf("failed to write to keyring: %w", err)
	}

	return nil
}

// Delete removes data from the kernel keyring
func (k *keyringAccessor) Delete(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Search for the key
	keyID, err := unix.KeyctlSearch(k.keyringID, "user", k.keyName, 0)
	if err != nil {
		if err == syscall.ENOKEY {
			return nil // Key doesn't exist, consider deletion successful
		}
		return fmt.Errorf("keyring search failed: %w", err)
	}

	// Revoke the key
	_, err = unix.KeyctlInt(unix.KEYCTL_REVOKE, keyID, 0, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to delete from keyring: %w", err)
	}

	return nil
}

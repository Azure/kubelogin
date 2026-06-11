package cmd

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestRemoveCacheDirReturnsErrorOnFailure(t *testing.T) {
	badPath := pathWithFileAsParent(t)
	cmd := newRemoveAuthRecordCacheCmd()
	err := executeCommand(cmd, "--cache-dir", badPath)

	if err == nil {
		t.Fatal("expected remove-cache-dir to return an error when cache deletion fails")
	}
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("expected wrapped cache deletion path error, got: %v", err)
	}
}

func TestRemoveTokensReturnsErrorOnFailure(t *testing.T) {
	badPath := pathWithFileAsParent(t)
	cmd := newRemoveAuthRecordCacheCmdDeprecated()
	err := executeCommand(cmd, "--token-cache-dir", badPath)

	if err == nil {
		t.Fatal("expected remove-tokens to return an error when cache deletion fails")
	}
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("expected wrapped cache deletion path error, got: %v", err)
	}
}

func TestRemoveCacheDirSucceedsForNonexistentPath(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "missing-cache")
	cmd := newRemoveAuthRecordCacheCmd()
	if err := executeCommand(cmd, "--cache-dir", cacheDir); err != nil {
		t.Fatalf("expected remove-cache-dir to succeed for a nonexistent path, got: %v", err)
	}
}

func TestRemoveTokensSucceedsForNonexistentPath(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "missing-cache")
	cmd := newRemoveAuthRecordCacheCmdDeprecated()
	if err := executeCommand(cmd, "--token-cache-dir", cacheDir); err != nil {
		t.Fatalf("expected remove-tokens to succeed for a nonexistent path, got: %v", err)
	}
}

func executeCommand(cmd *cobra.Command, args ...string) error {
	// Discard command output, including deprecation warnings
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func pathWithFileAsParent(t *testing.T) string {
	t.Helper()

	// Use a file parent to make os.RemoveAll fail deterministically
	parentFile := filepath.Join(t.TempDir(), "cache-parent")
	if err := os.WriteFile(parentFile, []byte("not a directory"), 0600); err != nil {
		t.Fatalf("failed to create parent file: %v", err)
	}
	return filepath.Join(parentFile, "child")
}

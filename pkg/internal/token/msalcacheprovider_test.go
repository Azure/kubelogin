package token

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
)

// stubMarshaler implements cache.Marshaler for testing Export
type stubMarshaler struct {
	data []byte
	err  error
}

func (s *stubMarshaler) Marshal() ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.data, nil
}

// stubUnmarshaler implements cache.Unmarshaler for testing Replace
type stubUnmarshaler struct {
	got []byte
	err error
}

func (s *stubUnmarshaler) Unmarshal(b []byte) error {
	s.got = append([]byte(nil), b...)
	return s.err
}

func TestExport_CreatesDirAndWritesFile(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "sub", "cache.bin")
	p := &defaultMSALCacheProvider{file: file}

	m := &stubMarshaler{data: []byte("hello-cache")}
	if err := p.Export(context.Background(), m, cache.ExportHints{}); err != nil {
		t.Fatalf("Export() unexpected error: %v", err)
	}

	// Verify file exists and contents match
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("reading exported file failed: %v", err)
	}
	if string(b) != "hello-cache" {
		t.Fatalf("unexpected file content: %q", string(b))
	}
}

func TestExport_MarshalError(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "cache.bin")
	p := &defaultMSALCacheProvider{file: file}

	wantErr := errors.New("boom")
	m := &stubMarshaler{err: wantErr}
	err := p.Export(context.Background(), m, cache.ExportHints{})
	if err == nil || !strings.Contains(err.Error(), "failed to marshal cache data") {
		t.Fatalf("expected marshal error, got: %v", err)
	}
}

func TestExport_MkdirAllError_IncludesDirInMessage(t *testing.T) {
	tmp := t.TempDir()
	// Create a regular file and then try to make it a parent directory to force MkdirAll to fail
	fileAsDir := filepath.Join(tmp, "notadir")
	if err := os.WriteFile(fileAsDir, []byte("x"), 0600); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	// c.file under the path that's a file; Dir(c.file) will be that file path
	target := filepath.Join(fileAsDir, "child", "cache.bin")
	p := &defaultMSALCacheProvider{file: target}

	m := &stubMarshaler{data: []byte("data")}
	err := p.Export(context.Background(), m, cache.ExportHints{})
	if err == nil {
		t.Fatalf("expected error from MkdirAll, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create cache directory") {
		t.Fatalf("expected mkdir error message, got: %v", err)
	}
	if !strings.Contains(err.Error(), filepath.Dir(target)) {
		t.Fatalf("expected path %q in error, got: %v", filepath.Dir(target), err)
	}
}

func TestExport_WriteFileError_IncludesFileInMessage(t *testing.T) {
	tmp := t.TempDir()
	// Create a directory and attempt to write to that directory path as if it were a file
	dir := filepath.Join(tmp, "dirfile")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	p := &defaultMSALCacheProvider{file: dir} // writing to a directory should fail

	m := &stubMarshaler{data: []byte("data")}
	err := p.Export(context.Background(), m, cache.ExportHints{})
	if err == nil {
		t.Fatalf("expected write error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to write cache file") {
		t.Fatalf("expected write error message, got: %v", err)
	}
	if !strings.Contains(err.Error(), dir) {
		t.Fatalf("expected file path %q in error, got: %v", dir, err)
	}
}

func TestExport_ContextCanceled(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "cache.bin")
	p := &defaultMSALCacheProvider{file: file}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	m := &stubMarshaler{data: []byte("data")}
	err := p.Export(ctx, m, cache.ExportHints{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
	if _, statErr := os.Stat(file); !os.IsNotExist(statErr) {
		t.Fatalf("file should not be created when context is canceled")
	}
}

func TestReplace_NoFile_ReturnsNil(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "missing.bin")
	p := &defaultMSALCacheProvider{file: file}

	u := &stubUnmarshaler{}
	if err := p.Replace(context.Background(), u, cache.ReplaceHints{}); err != nil {
		t.Fatalf("Replace() unexpected error for missing file: %v", err)
	}
}

func TestReplace_ReadFileError(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "adir")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	p := &defaultMSALCacheProvider{file: dir} // attempt to read a directory as a file

	u := &stubUnmarshaler{}
	err := p.Replace(context.Background(), u, cache.ReplaceHints{})
	if err == nil || !strings.Contains(err.Error(), "failed to read cache file") {
		t.Fatalf("expected read error, got: %v", err)
	}
}

func TestReplace_UnmarshalError(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "cache.bin")
	content := []byte("cached-bytes")
	if err := os.WriteFile(file, content, 0600); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}
	p := &defaultMSALCacheProvider{file: file}

	want := errors.New("unmarshal-bad")
	u := &stubUnmarshaler{err: want}
	err := p.Replace(context.Background(), u, cache.ReplaceHints{})
	if err == nil || !strings.Contains(err.Error(), "failed to unmarshal cache data") {
		t.Fatalf("expected unmarshal error, got: %v", err)
	}
}

func TestReplace_Success(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "cache.bin")
	content := []byte("cached-bytes")
	if err := os.WriteFile(file, content, 0600); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}
	p := &defaultMSALCacheProvider{file: file}

	u := &stubUnmarshaler{}
	if err := p.Replace(context.Background(), u, cache.ReplaceHints{}); err != nil {
		t.Fatalf("Replace() unexpected error: %v", err)
	}
	if string(u.got) != string(content) {
		t.Fatalf("unmarshal received %q, want %q", string(u.got), string(content))
	}
}

func TestReplace_ContextCanceled(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "cache.bin")
	// Even if file exists, canceled context should short-circuit
	if err := os.WriteFile(file, []byte("data"), 0600); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}
	p := &defaultMSALCacheProvider{file: file}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	u := &stubUnmarshaler{}
	err := p.Replace(ctx, u, cache.ReplaceHints{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

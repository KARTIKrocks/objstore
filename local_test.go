package objstore

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestLocalStorage(t *testing.T) *LocalStorage {
	t.Helper()
	dir := t.TempDir()
	store, err := NewLocalStorage(LocalConfig{
		BasePath:        dir,
		CreateDirs:      true,
		FilePermissions: 0644,
		DirPermissions:  0755,
	})
	if err != nil {
		t.Fatalf("NewLocalStorage: %v", err)
	}
	return store
}

func TestLocalStorage_PutAndGet(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	content := "hello local storage"
	info, err := store.Put(ctx, "test.txt", strings.NewReader(content))
	if err != nil {
		t.Fatalf("Put: %v", err)
	}

	if info.Path != "test.txt" {
		t.Errorf("Path = %q", info.Path)
	}
	if info.Name != "test.txt" {
		t.Errorf("Name = %q", info.Name)
	}
	if info.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", info.Size, len(content))
	}
	if info.ContentType != "text/plain" {
		t.Errorf("ContentType = %q", info.ContentType)
	}

	reader, err := store.Get(ctx, "test.txt")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer reader.Close()

	data, _ := io.ReadAll(reader)
	if string(data) != content {
		t.Errorf("content = %q, want %q", string(data), content)
	}
}

func TestLocalStorage_PutNestedDirectories(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	_, err := store.Put(ctx, "a/b/c/deep.txt", strings.NewReader("deep"))
	if err != nil {
		t.Fatalf("Put nested: %v", err)
	}

	data, err := GetString(ctx, store, "a/b/c/deep.txt")
	if err != nil {
		t.Fatalf("GetString: %v", err)
	}
	if data != "deep" {
		t.Errorf("content = %q", data)
	}
}

func TestLocalStorage_PutNoOverwrite(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	store.Put(ctx, "file.txt", strings.NewReader("first"))

	_, err := store.Put(ctx, "file.txt", strings.NewReader("second"), WithOverwrite(false))
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestLocalStorage_PutOverwrite(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	store.Put(ctx, "file.txt", strings.NewReader("first"))
	store.Put(ctx, "file.txt", strings.NewReader("second"))

	data, _ := GetString(ctx, store, "file.txt")
	if data != "second" {
		t.Errorf("content = %q, want %q", data, "second")
	}
}

func TestLocalStorage_PutCustomContentType(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	info, _ := store.Put(ctx, "data.bin", strings.NewReader("binary"),
		WithContentType("application/custom"),
	)

	if info.ContentType != "application/custom" {
		t.Errorf("ContentType = %q", info.ContentType)
	}
}

func TestLocalStorage_GetNotFound(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	_, err := store.Get(ctx, "missing.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestLocalStorage_Delete(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	store.Put(ctx, "file.txt", strings.NewReader("data"))

	err := store.Delete(ctx, "file.txt")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	exists, _ := store.Exists(ctx, "file.txt")
	if exists {
		t.Error("file still exists")
	}
}

func TestLocalStorage_DeleteNotFound(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	err := store.Delete(ctx, "missing.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestLocalStorage_Exists(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	exists, _ := store.Exists(ctx, "file.txt")
	if exists {
		t.Error("expected false")
	}

	store.Put(ctx, "file.txt", strings.NewReader("data"))

	exists, _ = store.Exists(ctx, "file.txt")
	if !exists {
		t.Error("expected true")
	}
}

func TestLocalStorage_Stat(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	store.Put(ctx, "photo.jpg", strings.NewReader("jpeg-data"))

	info, err := store.Stat(ctx, "photo.jpg")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	if info.Name != "photo.jpg" {
		t.Errorf("Name = %q", info.Name)
	}
	if info.Size != 9 {
		t.Errorf("Size = %d", info.Size)
	}
	if info.ContentType != "image/jpeg" {
		t.Errorf("ContentType = %q", info.ContentType)
	}
	if info.IsDir {
		t.Error("expected IsDir = false")
	}
}

func TestLocalStorage_StatNotFound(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	_, err := store.Stat(ctx, "missing.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestLocalStorage_Copy(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	store.Put(ctx, "src.txt", strings.NewReader("copy-me"))

	err := store.Copy(ctx, "src.txt", "dst.txt")
	if err != nil {
		t.Fatalf("Copy: %v", err)
	}

	srcData, _ := GetString(ctx, store, "src.txt")
	dstData, _ := GetString(ctx, store, "dst.txt")

	if srcData != "copy-me" {
		t.Errorf("source changed: %q", srcData)
	}
	if dstData != "copy-me" {
		t.Errorf("dst = %q", dstData)
	}
}

func TestLocalStorage_CopyToNestedDir(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	store.Put(ctx, "src.txt", strings.NewReader("data"))

	err := store.Copy(ctx, "src.txt", "a/b/dst.txt")
	if err != nil {
		t.Fatalf("Copy nested: %v", err)
	}

	data, _ := GetString(ctx, store, "a/b/dst.txt")
	if data != "data" {
		t.Errorf("dst = %q", data)
	}
}

func TestLocalStorage_CopyNotFound(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	err := store.Copy(ctx, "missing.txt", "dst.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestLocalStorage_Move(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	store.Put(ctx, "src.txt", strings.NewReader("move-me"))

	err := store.Move(ctx, "src.txt", "dst.txt")
	if err != nil {
		t.Fatalf("Move: %v", err)
	}

	exists, _ := store.Exists(ctx, "src.txt")
	if exists {
		t.Error("source still exists")
	}

	data, _ := GetString(ctx, store, "dst.txt")
	if data != "move-me" {
		t.Errorf("dst = %q", data)
	}
}

func TestLocalStorage_MoveToNestedDir(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	store.Put(ctx, "src.txt", strings.NewReader("data"))

	err := store.Move(ctx, "src.txt", "new/dir/dst.txt")
	if err != nil {
		t.Fatalf("Move nested: %v", err)
	}

	data, _ := GetString(ctx, store, "new/dir/dst.txt")
	if data != "data" {
		t.Errorf("dst = %q", data)
	}
}

func TestLocalStorage_List(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	store.Put(ctx, "images/a.jpg", strings.NewReader("a"))
	store.Put(ctx, "images/b.jpg", strings.NewReader("b"))
	store.Put(ctx, "images/sub/c.jpg", strings.NewReader("c"))
	store.Put(ctx, "docs/d.txt", strings.NewReader("d"))

	t.Run("list directory", func(t *testing.T) {
		result, err := store.List(ctx, "images/")
		if err != nil {
			t.Fatalf("List: %v", err)
		}

		if len(result.Files) != 2 {
			t.Errorf("got %d files, want 2: %+v", len(result.Files), result.Files)
		}
	})

	t.Run("recursive", func(t *testing.T) {
		result, err := store.List(ctx, "images/", WithRecursive(true))
		if err != nil {
			t.Fatalf("List: %v", err)
		}

		if len(result.Files) != 3 {
			t.Errorf("got %d files, want 3", len(result.Files))
		}
	})

	t.Run("with max keys", func(t *testing.T) {
		result, err := store.List(ctx, "images/", WithRecursive(true), WithMaxKeys(2))
		if err != nil {
			t.Fatalf("List: %v", err)
		}

		if len(result.Files) != 2 {
			t.Errorf("got %d files, want 2", len(result.Files))
		}
		if !result.IsTruncated {
			t.Error("expected IsTruncated")
		}
	})
}

func TestLocalStorage_URL(t *testing.T) {
	ctx := context.Background()

	t.Run("no base URL", func(t *testing.T) {
		store := newTestLocalStorage(t)
		_, err := store.URL(ctx, "file.txt")
		if !errors.Is(err, ErrNotImplemented) {
			t.Errorf("expected ErrNotImplemented, got %v", err)
		}
	})

	t.Run("with base URL", func(t *testing.T) {
		dir := t.TempDir()
		store, _ := NewLocalStorage(LocalConfig{
			BasePath: dir,
			BaseURL:  "https://cdn.example.com",
		})
		url, err := store.URL(ctx, "images/photo.jpg")
		if err != nil {
			t.Fatalf("URL: %v", err)
		}
		if url != "https://cdn.example.com/images/photo.jpg" {
			t.Errorf("URL = %q", url)
		}
	})
}

func TestLocalStorage_SignedURL(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewLocalStorage(LocalConfig{
		BasePath: dir,
		BaseURL:  "https://cdn.example.com",
	})

	url, err := store.SignedURL(context.Background(), "file.txt")
	if err != nil {
		t.Fatalf("SignedURL: %v", err)
	}
	// Local storage delegates to URL
	if url != "https://cdn.example.com/file.txt" {
		t.Errorf("SignedURL = %q", url)
	}
}

func TestLocalStorage_PathTraversal(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	// Attempt path traversal
	_, err := store.Put(ctx, "../../../etc/passwd", strings.NewReader("evil"))
	if !errors.Is(err, ErrInvalidPath) {
		t.Errorf("expected ErrInvalidPath for path traversal, got %v", err)
	}

	_, err = store.Get(ctx, "../../etc/passwd")
	if !errors.Is(err, ErrInvalidPath) {
		t.Errorf("expected ErrInvalidPath for path traversal, got %v", err)
	}
}

func TestLocalStorage_DeleteDir(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	store.Put(ctx, "dir/a.txt", strings.NewReader("a"))
	store.Put(ctx, "dir/b.txt", strings.NewReader("b"))
	store.Put(ctx, "dir/sub/c.txt", strings.NewReader("c"))

	err := store.DeleteDir(ctx, "dir")
	if err != nil {
		t.Fatalf("DeleteDir: %v", err)
	}

	exists, _ := store.Exists(ctx, "dir/a.txt")
	if exists {
		t.Error("files still exist after DeleteDir")
	}
}

func TestLocalStorage_Close(t *testing.T) {
	store := newTestLocalStorage(t)
	if err := store.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestLocalConfig_Builders(t *testing.T) {
	config := DefaultLocalConfig().
		WithBasePath("/tmp/test").
		WithBaseURL("https://example.com").
		WithCreateDirs(false).
		WithPermissions(0600, 0700)

	if config.BasePath != "/tmp/test" {
		t.Errorf("BasePath = %q", config.BasePath)
	}
	if config.BaseURL != "https://example.com" {
		t.Errorf("BaseURL = %q", config.BaseURL)
	}
	if config.CreateDirs {
		t.Error("expected CreateDirs = false")
	}
	if config.FilePermissions != 0600 {
		t.Errorf("FilePermissions = %o", config.FilePermissions)
	}
	if config.DirPermissions != 0700 {
		t.Errorf("DirPermissions = %o", config.DirPermissions)
	}
}

func TestNewLocalStorage_AbsolutePath(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalStorage(LocalConfig{
		BasePath:   filepath.Join(dir, "relative"),
		CreateDirs: true,
	})
	if err != nil {
		t.Fatalf("NewLocalStorage: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(store.config.BasePath)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("not a directory")
	}
	if !filepath.IsAbs(store.config.BasePath) {
		t.Error("BasePath not converted to absolute")
	}
}

// Verify LocalStorage implements Storage interface at compile time.
var _ Storage = (*LocalStorage)(nil)

package objstore

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
)

func newTestMemoryStorage() *MemoryStorage {
	return NewMemoryStorage()
}

func TestMemoryStorage_PutAndGet(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	content := "hello world"
	info, err := store.Put(ctx, "test.txt", strings.NewReader(content))
	if err != nil {
		t.Fatalf("Put: %v", err)
	}

	if info.Path != "test.txt" {
		t.Errorf("Path = %q, want %q", info.Path, "test.txt")
	}
	if info.Name != "test.txt" {
		t.Errorf("Name = %q, want %q", info.Name, "test.txt")
	}
	if info.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", info.Size, len(content))
	}
	if info.ContentType != "text/plain" {
		t.Errorf("ContentType = %q, want %q", info.ContentType, "text/plain")
	}

	// Get
	reader, err := store.Get(ctx, "test.txt")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(data) != content {
		t.Errorf("content = %q, want %q", string(data), content)
	}
}

func TestMemoryStorage_PutWithOptions(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	meta := map[string]string{"author": "test"}
	info, err := store.Put(ctx, "doc.bin", strings.NewReader("data"),
		WithContentType("application/custom"),
		WithMetadata(meta),
	)
	if err != nil {
		t.Fatalf("Put: %v", err)
	}

	if info.ContentType != "application/custom" {
		t.Errorf("ContentType = %q, want %q", info.ContentType, "application/custom")
	}

	// Verify via Stat
	stat, err := store.Stat(ctx, "doc.bin")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if stat.ContentType != "application/custom" {
		t.Errorf("Stat ContentType = %q, want %q", stat.ContentType, "application/custom")
	}
	if stat.Metadata["author"] != "test" {
		t.Error("Metadata not persisted")
	}
}

func TestMemoryStorage_PutNoOverwrite(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	_, err := store.Put(ctx, "file.txt", strings.NewReader("first"))
	if err != nil {
		t.Fatalf("first Put: %v", err)
	}

	_, err = store.Put(ctx, "file.txt", strings.NewReader("second"), WithOverwrite(false))
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}

	// Verify original content unchanged
	data, err := GetBytes(ctx, store, "file.txt")
	if err != nil {
		t.Fatalf("GetBytes: %v", err)
	}
	if string(data) != "first" {
		t.Errorf("content = %q, want %q", string(data), "first")
	}
}

func TestMemoryStorage_PutOverwrite(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	store.Put(ctx, "file.txt", strings.NewReader("first"))
	store.Put(ctx, "file.txt", strings.NewReader("second"))

	data, _ := GetBytes(ctx, store, "file.txt")
	if string(data) != "second" {
		t.Errorf("content = %q, want %q", string(data), "second")
	}
}

func TestMemoryStorage_GetNotFound(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	_, err := store.Get(ctx, "missing.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStorage_Delete(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	store.Put(ctx, "file.txt", strings.NewReader("data"))

	err := store.Delete(ctx, "file.txt")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	exists, _ := store.Exists(ctx, "file.txt")
	if exists {
		t.Error("file still exists after delete")
	}
}

func TestMemoryStorage_DeleteNotFound(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	err := store.Delete(ctx, "missing.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStorage_Exists(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	exists, err := store.Exists(ctx, "file.txt")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if exists {
		t.Error("expected false for non-existent file")
	}

	store.Put(ctx, "file.txt", strings.NewReader("data"))

	exists, err = store.Exists(ctx, "file.txt")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !exists {
		t.Error("expected true for existing file")
	}
}

func TestMemoryStorage_Stat(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	store.Put(ctx, "images/photo.jpg", strings.NewReader("jpeg-data"))

	info, err := store.Stat(ctx, "images/photo.jpg")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	if info.Path != "images/photo.jpg" {
		t.Errorf("Path = %q", info.Path)
	}
	if info.Name != "photo.jpg" {
		t.Errorf("Name = %q", info.Name)
	}
	if info.Size != 9 {
		t.Errorf("Size = %d, want 9", info.Size)
	}
	if info.ContentType != "image/jpeg" {
		t.Errorf("ContentType = %q", info.ContentType)
	}
}

func TestMemoryStorage_StatNotFound(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	_, err := store.Stat(ctx, "missing.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStorage_Copy(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	store.Put(ctx, "src.txt", strings.NewReader("hello"))

	err := store.Copy(ctx, "src.txt", "dst.txt")
	if err != nil {
		t.Fatalf("Copy: %v", err)
	}

	// Both should exist
	srcData, _ := GetBytes(ctx, store, "src.txt")
	dstData, _ := GetBytes(ctx, store, "dst.txt")

	if string(srcData) != "hello" {
		t.Errorf("source content changed: %q", string(srcData))
	}
	if string(dstData) != "hello" {
		t.Errorf("dst content = %q, want %q", string(dstData), "hello")
	}
}

func TestMemoryStorage_CopyDeepCopy(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	store.Put(ctx, "src.txt", strings.NewReader("hello"))
	store.Copy(ctx, "src.txt", "dst.txt")

	// Modify source - dst should not change
	store.Put(ctx, "src.txt", strings.NewReader("modified"))

	dstData, _ := GetBytes(ctx, store, "dst.txt")
	if string(dstData) != "hello" {
		t.Errorf("copy was not independent: dst = %q", string(dstData))
	}
}

func TestMemoryStorage_CopyNotFound(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	err := store.Copy(ctx, "missing.txt", "dst.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStorage_Move(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	store.Put(ctx, "src.txt", strings.NewReader("hello"))

	err := store.Move(ctx, "src.txt", "dst.txt")
	if err != nil {
		t.Fatalf("Move: %v", err)
	}

	// Source should be gone
	exists, _ := store.Exists(ctx, "src.txt")
	if exists {
		t.Error("source still exists after move")
	}

	// Destination should have content
	data, _ := GetBytes(ctx, store, "dst.txt")
	if string(data) != "hello" {
		t.Errorf("dst content = %q, want %q", string(data), "hello")
	}
}

func TestMemoryStorage_MoveNotFound(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	err := store.Move(ctx, "missing.txt", "dst.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStorage_List(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	store.Put(ctx, "images/a.jpg", strings.NewReader("a"))
	store.Put(ctx, "images/b.jpg", strings.NewReader("b"))
	store.Put(ctx, "images/sub/c.jpg", strings.NewReader("c"))
	store.Put(ctx, "docs/d.txt", strings.NewReader("d"))

	t.Run("non-recursive with delimiter", func(t *testing.T) {
		result, err := store.List(ctx, "images/")
		if err != nil {
			t.Fatalf("List: %v", err)
		}

		if len(result.Files) != 2 {
			t.Errorf("got %d files, want 2", len(result.Files))
		}
		if len(result.Prefixes) != 1 {
			t.Errorf("got %d prefixes, want 1", len(result.Prefixes))
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
			t.Error("expected IsTruncated to be true")
		}
	})

	t.Run("top-level prefix lists matching", func(t *testing.T) {
		result, err := store.List(ctx, "images/", WithRecursive(true))
		if err != nil {
			t.Fatalf("List: %v", err)
		}

		if len(result.Files) != 3 {
			t.Errorf("got %d files, want 3", len(result.Files))
		}
	})

	t.Run("no match", func(t *testing.T) {
		result, err := store.List(ctx, "nonexistent/", WithRecursive(true))
		if err != nil {
			t.Fatalf("List: %v", err)
		}

		if len(result.Files) != 0 {
			t.Errorf("got %d files, want 0", len(result.Files))
		}
	})
}

func TestMemoryStorage_URL(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	t.Run("no base URL", func(t *testing.T) {
		_, err := store.URL(ctx, "file.txt")
		if !errors.Is(err, ErrNotImplemented) {
			t.Errorf("expected ErrNotImplemented, got %v", err)
		}
	})

	t.Run("with base URL", func(t *testing.T) {
		store.WithBaseURL("https://cdn.example.com")
		url, err := store.URL(ctx, "images/photo.jpg")
		if err != nil {
			t.Fatalf("URL: %v", err)
		}
		if url != "https://cdn.example.com/images/photo.jpg" {
			t.Errorf("URL = %q", url)
		}
	})

	t.Run("base URL trailing slash", func(t *testing.T) {
		store.WithBaseURL("https://cdn.example.com/")
		url, _ := store.URL(ctx, "file.txt")
		if url != "https://cdn.example.com/file.txt" {
			t.Errorf("URL = %q", url)
		}
	})
}

func TestMemoryStorage_SignedURL(t *testing.T) {
	store := NewMemoryStorage().WithBaseURL("https://cdn.example.com")
	ctx := context.Background()

	url, err := store.SignedURL(ctx, "file.txt")
	if err != nil {
		t.Fatalf("SignedURL: %v", err)
	}
	// Memory storage delegates to URL
	if url != "https://cdn.example.com/file.txt" {
		t.Errorf("SignedURL = %q", url)
	}
}

func TestMemoryStorage_Clear(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	store.Put(ctx, "a.txt", strings.NewReader("a"))
	store.Put(ctx, "b.txt", strings.NewReader("b"))

	if store.Size() != 2 {
		t.Errorf("Size = %d, want 2", store.Size())
	}

	store.Clear()

	if store.Size() != 0 {
		t.Errorf("Size after Clear = %d, want 0", store.Size())
	}
}

func TestMemoryStorage_TotalBytes(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	store.Put(ctx, "a.txt", strings.NewReader("hello"))  // 5 bytes
	store.Put(ctx, "b.txt", strings.NewReader("world!")) // 6 bytes

	if store.TotalBytes() != 11 {
		t.Errorf("TotalBytes = %d, want 11", store.TotalBytes())
	}
}

func TestMemoryStorage_GetBytes(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	store.Put(ctx, "file.txt", strings.NewReader("original"))

	data, err := store.GetBytes("file.txt")
	if err != nil {
		t.Fatalf("GetBytes: %v", err)
	}
	if string(data) != "original" {
		t.Errorf("data = %q", string(data))
	}

	// Verify returned slice is a copy (mutating it shouldn't affect stored data)
	data[0] = 'X'
	data2, _ := store.GetBytes("file.txt")
	if string(data2) != "original" {
		t.Error("GetBytes returned a reference to internal data instead of a copy")
	}
}

func TestMemoryStorage_GetBytesNotFound(t *testing.T) {
	store := newTestMemoryStorage()

	_, err := store.GetBytes("missing.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStorage_PathNormalization(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	// Put with leading slash
	store.Put(ctx, "/foo/bar.txt", strings.NewReader("data"))

	// Get without leading slash
	_, err := store.Get(ctx, "foo/bar.txt")
	if err != nil {
		t.Errorf("path normalization failed: %v", err)
	}
}

func TestMemoryStorage_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	var wg sync.WaitGroup
	errs := make(chan error, 100)

	// Concurrent writes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "file" + string(rune('A'+i%26)) + ".txt"
			_, err := store.Put(ctx, key, strings.NewReader("data"))
			if err != nil {
				errs <- err
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.List(ctx, "", WithRecursive(true))
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent error: %v", err)
	}
}

// Verify MemoryStorage implements Storage interface at compile time.
var _ Storage = (*MemoryStorage)(nil)

func TestMemoryStorage_PutEmptyFile(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryStorage()

	info, err := store.Put(ctx, "empty.txt", bytes.NewReader(nil))
	if err != nil {
		t.Fatalf("Put empty: %v", err)
	}
	if info.Size != 0 {
		t.Errorf("Size = %d, want 0", info.Size)
	}

	data, _ := GetBytes(ctx, store, "empty.txt")
	if len(data) != 0 {
		t.Errorf("got %d bytes, want 0", len(data))
	}
}

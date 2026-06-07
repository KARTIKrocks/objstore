package objstore

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPutBytes(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()

	info, err := PutBytes(ctx, store, "data.bin", []byte{1, 2, 3})
	if err != nil {
		t.Fatalf("PutBytes: %v", err)
	}
	if info.Size != 3 {
		t.Errorf("Size = %d, want 3", info.Size)
	}

	data, _ := GetBytes(ctx, store, "data.bin")
	if len(data) != 3 || data[0] != 1 || data[1] != 2 || data[2] != 3 {
		t.Errorf("data = %v", data)
	}
}

func TestPutString(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()

	_, err := PutString(ctx, store, "hello.txt", "Hello, World!")
	if err != nil {
		t.Fatalf("PutString: %v", err)
	}

	text, _ := GetString(ctx, store, "hello.txt")
	if text != "Hello, World!" {
		t.Errorf("text = %q", text)
	}
}

func TestGetBytes(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()

	store.Put(ctx, "file.txt", strings.NewReader("hello"))

	data, err := GetBytes(ctx, store, "file.txt")
	if err != nil {
		t.Fatalf("GetBytes: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("data = %q", string(data))
	}
}

func TestGetBytesNotFound(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()

	_, err := GetBytes(ctx, store, "missing.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetString(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()

	store.Put(ctx, "file.txt", strings.NewReader("world"))

	text, err := GetString(ctx, store, "file.txt")
	if err != nil {
		t.Fatalf("GetString: %v", err)
	}
	if text != "world" {
		t.Errorf("text = %q", text)
	}
}

func TestCopyTo(t *testing.T) {
	ctx := context.Background()
	src := NewMemoryStorage()
	dst := NewMemoryStorage()

	PutString(ctx, src, "file.txt", "cross-storage")

	err := CopyTo(ctx, src, "file.txt", dst, "copy.txt")
	if err != nil {
		t.Fatalf("CopyTo: %v", err)
	}

	// Source still exists
	text, _ := GetString(ctx, src, "file.txt")
	if text != "cross-storage" {
		t.Errorf("source changed: %q", text)
	}

	// Destination has copy
	text, _ = GetString(ctx, dst, "copy.txt")
	if text != "cross-storage" {
		t.Errorf("dst = %q", text)
	}
}

func TestCopyToNotFound(t *testing.T) {
	ctx := context.Background()
	src := NewMemoryStorage()
	dst := NewMemoryStorage()

	err := CopyTo(ctx, src, "missing.txt", dst, "copy.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMoveTo(t *testing.T) {
	ctx := context.Background()
	src := NewMemoryStorage()
	dst := NewMemoryStorage()

	PutString(ctx, src, "file.txt", "move-me")

	err := MoveTo(ctx, src, "file.txt", dst, "moved.txt")
	if err != nil {
		t.Fatalf("MoveTo: %v", err)
	}

	// Source should be deleted
	exists, _ := src.Exists(ctx, "file.txt")
	if exists {
		t.Error("source still exists")
	}

	// Destination should have content
	text, _ := GetString(ctx, dst, "moved.txt")
	if text != "move-me" {
		t.Errorf("dst = %q", text)
	}
}

func TestGenerateFileName(t *testing.T) {
	name := GenerateFileName("photo.jpg")
	if !strings.HasSuffix(name, ".jpg") {
		t.Errorf("missing extension: %q", name)
	}
	// UUID + .jpg = 36 + 4 = 40 chars
	if len(name) != 40 {
		t.Errorf("unexpected length %d: %q", len(name), name)
	}

	// Uniqueness
	name2 := GenerateFileName("photo.jpg")
	if name == name2 {
		t.Error("expected unique filenames")
	}
}

func TestGenerateFileName_NoExtension(t *testing.T) {
	name := GenerateFileName("Makefile")
	if strings.Contains(name, ".") {
		t.Errorf("unexpected dot in %q for extensionless file", name)
	}
}

func TestGeneratePath(t *testing.T) {
	path := GeneratePath("photo.jpg", "uploads")

	if !strings.HasPrefix(path, "uploads/") {
		t.Errorf("missing prefix: %q", path)
	}
	if !strings.HasSuffix(path, ".jpg") {
		t.Errorf("missing extension: %q", path)
	}

	// Should have date components: uploads/YYYY/MM/DD/uuid.jpg
	parts := strings.Split(path, "/")
	if len(parts) != 5 {
		t.Errorf("expected 5 path parts, got %d: %q", len(parts), path)
	}

	// Verify forward slashes (not backslashes)
	if strings.Contains(path, "\\") {
		t.Errorf("path contains backslashes: %q", path)
	}
}

func TestGenerateHashedPath(t *testing.T) {
	path := GenerateHashedPath("photo.jpg", "uploads", 2)

	if !strings.HasPrefix(path, "uploads/") {
		t.Errorf("missing prefix: %q", path)
	}
	if !strings.HasSuffix(path, ".jpg") {
		t.Errorf("missing extension: %q", path)
	}

	// Should have: uploads/XX/XX/uuid.jpg
	parts := strings.Split(path, "/")
	if len(parts) != 4 {
		t.Errorf("expected 4 path parts, got %d: %q", len(parts), path)
	}

	// Hash directories should be 2 chars each
	if len(parts[1]) != 2 || len(parts[2]) != 2 {
		t.Errorf("hash dirs wrong length: %q, %q", parts[1], parts[2])
	}

	// Verify forward slashes
	if strings.Contains(path, "\\") {
		t.Errorf("path contains backslashes: %q", path)
	}
}

func TestGenerateHashedPath_ZeroLevels(t *testing.T) {
	path := GenerateHashedPath("file.txt", "prefix", 0)

	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		t.Errorf("expected 2 path parts, got %d: %q", len(parts), path)
	}
}

func TestDeletePrefix(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()

	PutString(ctx, store, "images/a.jpg", "a")
	PutString(ctx, store, "images/b.jpg", "b")
	PutString(ctx, store, "images/sub/c.jpg", "c")
	PutString(ctx, store, "docs/d.txt", "d")

	err := DeletePrefix(ctx, store, "images/")
	if err != nil {
		t.Fatalf("DeletePrefix: %v", err)
	}

	// images/ files should be gone
	exists, _ := store.Exists(ctx, "images/a.jpg")
	if exists {
		t.Error("images/a.jpg still exists")
	}
	exists, _ = store.Exists(ctx, "images/sub/c.jpg")
	if exists {
		t.Error("images/sub/c.jpg still exists")
	}

	// docs/ file should remain
	exists, _ = store.Exists(ctx, "docs/d.txt")
	if !exists {
		t.Error("docs/d.txt was deleted")
	}
}

func TestDeletePrefix_Empty(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()

	err := DeletePrefix(ctx, store, "nonexistent/")
	if err != nil {
		t.Fatalf("DeletePrefix empty: %v", err)
	}
}

func TestSyncDir(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()

	// Create temp directory structure
	dir := t.TempDir()
	writeTestFile(t, dir, "a.txt", "alpha")
	writeTestFile(t, dir, "sub/b.txt", "beta")

	err := SyncDir(ctx, store, dir, "remote")
	if err != nil {
		t.Fatalf("SyncDir: %v", err)
	}

	text, err := GetString(ctx, store, "remote/a.txt")
	if err != nil {
		t.Fatalf("Get a.txt: %v", err)
	}
	if text != "alpha" {
		t.Errorf("a.txt = %q", text)
	}

	text, err = GetString(ctx, store, "remote/sub/b.txt")
	if err != nil {
		t.Fatalf("Get sub/b.txt: %v", err)
	}
	if text != "beta" {
		t.Errorf("sub/b.txt = %q", text)
	}
}

func TestSyncDir_ContextCancellation(t *testing.T) {
	store := NewMemoryStorage()

	dir := t.TempDir()
	writeTestFile(t, dir, "a.txt", "data")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := SyncDir(ctx, store, dir, "remote")
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}

func TestParseDataURI(t *testing.T) {
	t.Run("base64 image", func(t *testing.T) {
		// "hello" in base64
		data, mime, err := ParseDataURI("data:image/png;base64,aGVsbG8=")
		if err != nil {
			t.Fatalf("ParseDataURI: %v", err)
		}
		if string(data) != "hello" {
			t.Errorf("data = %q", string(data))
		}
		if mime != "image/png" {
			t.Errorf("mime = %q", mime)
		}
	})

	t.Run("plain text", func(t *testing.T) {
		data, mime, err := ParseDataURI("data:text/plain,Hello%20World")
		if err != nil {
			t.Fatalf("ParseDataURI: %v", err)
		}
		if string(data) != "Hello World" {
			t.Errorf("data = %q", string(data))
		}
		if mime != "text/plain" {
			t.Errorf("mime = %q", mime)
		}
	})

	t.Run("default mime type", func(t *testing.T) {
		data, mime, err := ParseDataURI("data:,simple")
		if err != nil {
			t.Fatalf("ParseDataURI: %v", err)
		}
		if string(data) != "simple" {
			t.Errorf("data = %q", string(data))
		}
		if mime != "text/plain" {
			t.Errorf("mime = %q, want text/plain", mime)
		}
	})

	t.Run("invalid prefix", func(t *testing.T) {
		_, _, err := ParseDataURI("http://example.com")
		if err == nil {
			t.Error("expected error for invalid data URI")
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		_, _, err := ParseDataURI("data:nope")
		if err == nil {
			t.Error("expected error for missing comma")
		}
	})
}

func TestPutDataURI(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()

	info, err := PutDataURI(ctx, store, "image.png", "data:image/png;base64,aGVsbG8=")
	if err != nil {
		t.Fatalf("PutDataURI: %v", err)
	}

	if info.ContentType != "image/png" {
		t.Errorf("ContentType = %q", info.ContentType)
	}

	data, _ := GetBytes(ctx, store, "image.png")
	if string(data) != "hello" {
		t.Errorf("data = %q", string(data))
	}
}

func TestIsImage(t *testing.T) {
	tests := []struct {
		contentType string
		want        bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"image/gif", true},
		{"text/plain", false},
		{"video/mp4", false},
	}
	for _, tt := range tests {
		info := &FileInfo{ContentType: tt.contentType}
		if got := IsImage(info); got != tt.want {
			t.Errorf("IsImage(%q) = %v, want %v", tt.contentType, got, tt.want)
		}
	}
}

func TestIsVideo(t *testing.T) {
	if !IsVideo(&FileInfo{ContentType: "video/mp4"}) {
		t.Error("expected true for video/mp4")
	}
	if IsVideo(&FileInfo{ContentType: "image/png"}) {
		t.Error("expected false for image/png")
	}
}

func TestIsAudio(t *testing.T) {
	if !IsAudio(&FileInfo{ContentType: "audio/mpeg"}) {
		t.Error("expected true for audio/mpeg")
	}
	if IsAudio(&FileInfo{ContentType: "text/plain"}) {
		t.Error("expected false for text/plain")
	}
}

func TestIsDocument(t *testing.T) {
	tests := []struct {
		contentType string
		want        bool
	}{
		{"application/pdf", true},
		{"application/msword", true},
		{"application/vnd.openxmlformats-officedocument.wordprocessingml.document", true},
		{"text/plain", true},
		{"text/csv", true},
		{"image/png", false},
		{"video/mp4", false},
	}
	for _, tt := range tests {
		info := &FileInfo{ContentType: tt.contentType}
		if got := IsDocument(info); got != tt.want {
			t.Errorf("IsDocument(%q) = %v, want %v", tt.contentType, got, tt.want)
		}
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		size int64
		want string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}
	for _, tt := range tests {
		got := FormatSize(tt.size)
		if got != tt.want {
			t.Errorf("FormatSize(%d) = %q, want %q", tt.size, got, tt.want)
		}
	}
}

// writeTestFile creates a file in the given directory with content.
func writeTestFile(t *testing.T, dir, path, content string) {
	t.Helper()
	full := filepath.Join(dir, path)
	os.MkdirAll(filepath.Dir(full), 0755)
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

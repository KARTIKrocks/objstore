package objstore

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PutBytes uploads bytes to storage.
func PutBytes(ctx context.Context, s Storage, path string, data []byte, opts ...PutOption) (*FileInfo, error) {
	return s.Put(ctx, path, bytes.NewReader(data), opts...)
}

// PutString uploads a string to storage.
func PutString(ctx context.Context, s Storage, path string, data string, opts ...PutOption) (*FileInfo, error) {
	return s.Put(ctx, path, strings.NewReader(data), opts...)
}

// GetBytes retrieves content as bytes.
func GetBytes(ctx context.Context, s Storage, path string) ([]byte, error) {
	reader, err := s.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()

	return io.ReadAll(reader)
}

// GetString retrieves content as a string.
func GetString(ctx context.Context, s Storage, path string) (string, error) {
	data, err := GetBytes(ctx, s, path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CopyTo copies a file from one storage to another.
func CopyTo(ctx context.Context, src Storage, srcPath string, dst Storage, dstPath string, opts ...PutOption) error {
	reader, err := src.Get(ctx, srcPath)
	if err != nil {
		return err
	}
	defer func() { _ = reader.Close() }()

	// Get source file info for content type
	info, err := src.Stat(ctx, srcPath)
	if err == nil && info.ContentType != "" {
		opts = append([]PutOption{WithContentType(info.ContentType)}, opts...)
	}

	_, err = dst.Put(ctx, dstPath, reader, opts...)
	return err
}

// MoveTo moves a file from one storage to another.
func MoveTo(ctx context.Context, src Storage, srcPath string, dst Storage, dstPath string, opts ...PutOption) error {
	if err := CopyTo(ctx, src, srcPath, dst, dstPath, opts...); err != nil {
		return err
	}
	return src.Delete(ctx, srcPath)
}

// GenerateFileName generates a unique filename with the original extension.
func GenerateFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	return uuid.New().String() + ext
}

// GeneratePath generates a unique path with date-based organization.
func GeneratePath(originalName string, prefix string) string {
	now := time.Now()
	filename := GenerateFileName(originalName)

	return strings.Join([]string{
		prefix,
		now.Format("2006"),
		now.Format("01"),
		now.Format("02"),
		filename,
	}, "/")
}

// GenerateHashedPath generates a path using hash-based distribution.
// This helps distribute files across directories when you have many files.
func GenerateHashedPath(originalName string, prefix string, levels int) string {
	id := uuid.New().String()
	ext := filepath.Ext(originalName)

	// Use first characters of UUID for directory structure
	parts := []string{prefix}
	for i := 0; i < levels && i*2 < len(id); i++ {
		parts = append(parts, id[i*2:i*2+2])
	}
	parts = append(parts, id+ext)

	return strings.Join(parts, "/")
}

// DeletePrefix deletes all files matching a prefix.
// If the storage backend implements BatchDeleter, it uses batch deletion for efficiency.
func DeletePrefix(ctx context.Context, s Storage, prefix string) error {
	result, err := s.List(ctx, prefix, WithRecursive(true), WithMaxKeys(1000))
	if err != nil {
		return err
	}

	if err := deleteBatch(ctx, s, result.Files); err != nil {
		return err
	}

	// Handle truncated results
	for result.IsTruncated {
		result, err = s.List(ctx, prefix, WithRecursive(true), WithMaxKeys(1000), WithToken(result.NextToken))
		if err != nil {
			return err
		}

		if err := deleteBatch(ctx, s, result.Files); err != nil {
			return err
		}
	}

	return nil
}

// deleteBatch deletes files using batch deletion if available, otherwise one-by-one.
func deleteBatch(ctx context.Context, s Storage, files []*FileInfo) error {
	if len(files) == 0 {
		return nil
	}

	// Use batch deletion if the backend supports it
	if bd, ok := s.(BatchDeleter); ok {
		paths := make([]string, len(files))
		for i, f := range files {
			paths[i] = f.Path
		}
		return bd.DeleteMultiple(ctx, paths)
	}

	// Fall back to one-by-one deletion
	for _, file := range files {
		if err := s.Delete(ctx, file.Path); err != nil && !errors.Is(err, ErrNotFound) {
			return err
		}
	}
	return nil
}

// SyncDir synchronizes a local directory to storage.
func SyncDir(ctx context.Context, s Storage, localPath, remotePath string) error {
	return filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return err
		}

		dstPath := strings.TrimSuffix(remotePath, "/") + "/" + filepath.ToSlash(relPath)

		// Open local file
		file, err := os.Open(path)
		if err != nil {
			return err
		}

		_, putErr := s.Put(ctx, dstPath, file)
		_ = file.Close()
		return putErr
	})
}

// ParseDataURI parses a data URI and returns the content and MIME type.
func ParseDataURI(dataURI string) ([]byte, string, error) {
	if !strings.HasPrefix(dataURI, "data:") {
		return nil, "", fmt.Errorf("invalid data URI")
	}

	// Remove "data:" prefix
	dataURI = strings.TrimPrefix(dataURI, "data:")

	// Split by comma
	parts := strings.SplitN(dataURI, ",", 2)
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid data URI format")
	}

	// Parse mime type and encoding
	header := parts[0]
	content := parts[1]

	mimeType := "text/plain"
	isBase64 := false

	headerParts := strings.Split(header, ";")
	if len(headerParts) > 0 && headerParts[0] != "" {
		mimeType = headerParts[0]
	}
	for _, p := range headerParts[1:] {
		if p == "base64" {
			isBase64 = true
		}
	}

	// Decode content
	var data []byte
	var decErr error

	if isBase64 {
		data, decErr = base64.StdEncoding.DecodeString(content)
	} else {
		// Handle percent-encoded data URIs
		decoded, err := url.QueryUnescape(content)
		if err != nil {
			data = []byte(content)
		} else {
			data = []byte(decoded)
		}
	}

	if decErr != nil {
		return nil, "", decErr
	}

	return data, mimeType, nil
}

// PutDataURI uploads content from a data URI.
func PutDataURI(ctx context.Context, s Storage, path string, dataURI string, opts ...PutOption) (*FileInfo, error) {
	data, mimeType, err := ParseDataURI(dataURI)
	if err != nil {
		return nil, err
	}

	// Prepend content type option
	opts = append([]PutOption{WithContentType(mimeType)}, opts...)

	return PutBytes(ctx, s, path, data, opts...)
}

// IsImage checks if a file is an image based on its content type.
func IsImage(info *FileInfo) bool {
	return strings.HasPrefix(info.ContentType, "image/")
}

// IsVideo checks if a file is a video based on its content type.
func IsVideo(info *FileInfo) bool {
	return strings.HasPrefix(info.ContentType, "video/")
}

// IsAudio checks if a file is audio based on its content type.
func IsAudio(info *FileInfo) bool {
	return strings.HasPrefix(info.ContentType, "audio/")
}

// IsDocument checks if a file is a document (PDF, Word, Excel, etc.).
func IsDocument(info *FileInfo) bool {
	docTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument",
		"application/vnd.ms-excel",
		"application/vnd.ms-powerpoint",
		"text/plain",
		"text/csv",
	}

	for _, t := range docTypes {
		if strings.HasPrefix(info.ContentType, t) {
			return true
		}
	}
	return false
}

// FormatSize formats a file size in human-readable format.
func FormatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

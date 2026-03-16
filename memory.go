package objstore

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// MemoryStorage implements Storage with in-memory storage for testing.
type MemoryStorage struct {
	mu      sync.RWMutex
	files   map[string]*memoryFile
	baseURL string
}

type memoryFile struct {
	data        []byte
	contentType string
	metadata    map[string]string
	modTime     time.Time
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		files: make(map[string]*memoryFile),
	}
}

// WithBaseURL sets the base URL for the memory storage.
func (s *MemoryStorage) WithBaseURL(url string) *MemoryStorage {
	s.baseURL = url
	return s
}

// Close is a no-op for memory storage.
func (s *MemoryStorage) Close() error {
	return nil
}

// Put uploads content to memory.
func (s *MemoryStorage) Put(ctx context.Context, path string, reader io.Reader, opts ...PutOption) (*FileInfo, error) {
	options := ApplyPutOptions(opts)

	path = NormalizePath(path)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if file exists
	if !options.Overwrite {
		if _, exists := s.files[path]; exists {
			return nil, ErrAlreadyExists
		}
	}

	// Read content
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// Detect content type
	contentType := options.ContentType
	if contentType == "" {
		contentType = DetectContentType(path)
	}

	// Store file
	s.files[path] = &memoryFile{
		data:        data,
		contentType: contentType,
		metadata:    options.Metadata,
		modTime:     time.Now(),
	}

	return &FileInfo{
		Path:         path,
		Name:         filepath.Base(path),
		Size:         int64(len(data)),
		ContentType:  contentType,
		LastModified: s.files[path].modTime,
		Metadata:     options.Metadata,
	}, nil
}

// Get retrieves content from memory.
func (s *MemoryStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	path = NormalizePath(path)

	s.mu.RLock()
	defer s.mu.RUnlock()

	file, exists := s.files[path]
	if !exists {
		return nil, ErrNotFound
	}

	return io.NopCloser(bytes.NewReader(file.data)), nil
}

// Delete removes a file from memory.
func (s *MemoryStorage) Delete(ctx context.Context, path string) error {
	path = NormalizePath(path)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.files[path]; !exists {
		return ErrNotFound
	}

	delete(s.files, path)
	return nil
}

// Exists checks if a file exists in memory.
func (s *MemoryStorage) Exists(ctx context.Context, path string) (bool, error) {
	path = NormalizePath(path)

	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.files[path]
	return exists, nil
}

// Stat returns file information.
func (s *MemoryStorage) Stat(ctx context.Context, path string) (*FileInfo, error) {
	path = NormalizePath(path)

	s.mu.RLock()
	defer s.mu.RUnlock()

	file, exists := s.files[path]
	if !exists {
		return nil, ErrNotFound
	}

	return &FileInfo{
		Path:         path,
		Name:         filepath.Base(path),
		Size:         int64(len(file.data)),
		ContentType:  file.contentType,
		LastModified: file.modTime,
		Metadata:     file.metadata,
	}, nil
}

// List returns files matching the prefix.
func (s *MemoryStorage) List(ctx context.Context, prefix string, opts ...ListOption) (*ListResult, error) {
	options := ApplyListOptions(opts)

	prefix = NormalizePath(prefix)

	s.mu.RLock()
	defer s.mu.RUnlock()

	result := &ListResult{
		Files:    make([]*FileInfo, 0),
		Prefixes: make([]string, 0),
	}

	prefixSet := make(map[string]bool)

	// Get sorted keys
	keys := make([]string, 0, len(s.files))
	for k := range s.files {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	count := 0
	for _, path := range keys {
		if !strings.HasPrefix(path, prefix) {
			continue
		}

		// Handle non-recursive listing
		if !options.Recursive && options.Delimiter != "" {
			relPath := strings.TrimPrefix(path, prefix)
			relPath = strings.TrimPrefix(relPath, options.Delimiter)

			if idx := strings.Index(relPath, options.Delimiter); idx >= 0 {
				// This is a nested file, add prefix
				prefixPath := prefix + relPath[:idx+1]
				if !prefixSet[prefixPath] {
					prefixSet[prefixPath] = true
					result.Prefixes = append(result.Prefixes, prefixPath)
				}
				continue
			}
		}

		if options.MaxKeys > 0 && count >= options.MaxKeys {
			result.IsTruncated = true
			break
		}

		file := s.files[path]
		result.Files = append(result.Files, &FileInfo{
			Path:         path,
			Name:         filepath.Base(path),
			Size:         int64(len(file.data)),
			ContentType:  file.contentType,
			LastModified: file.modTime,
			Metadata:     file.metadata,
		})
		count++
	}

	return result, nil
}

// Copy copies a file in memory.
func (s *MemoryStorage) Copy(ctx context.Context, src, dst string) error {
	src = NormalizePath(src)
	dst = NormalizePath(dst)

	s.mu.Lock()
	defer s.mu.Unlock()

	file, exists := s.files[src]
	if !exists {
		return ErrNotFound
	}

	// Deep copy
	newData := make([]byte, len(file.data))
	copy(newData, file.data)

	var newMetadata map[string]string
	if file.metadata != nil {
		newMetadata = make(map[string]string, len(file.metadata))
		for k, v := range file.metadata {
			newMetadata[k] = v
		}
	}

	s.files[dst] = &memoryFile{
		data:        newData,
		contentType: file.contentType,
		metadata:    newMetadata,
		modTime:     time.Now(),
	}

	return nil
}

// Move moves a file in memory atomically.
func (s *MemoryStorage) Move(ctx context.Context, src, dst string) error {
	src = NormalizePath(src)
	dst = NormalizePath(dst)

	s.mu.Lock()
	defer s.mu.Unlock()

	file, exists := s.files[src]
	if !exists {
		return ErrNotFound
	}

	// Move the file directly — no deep copy needed since we're deleting the source.
	file.modTime = time.Now()
	s.files[dst] = file
	delete(s.files, src)

	return nil
}

// URL returns a URL for the file.
func (s *MemoryStorage) URL(ctx context.Context, path string) (string, error) {
	if s.baseURL == "" {
		return "", ErrNotImplemented
	}
	return strings.TrimSuffix(s.baseURL, "/") + "/" + strings.TrimPrefix(path, "/"), nil
}

// SignedURL returns a signed URL (not supported for memory storage).
func (s *MemoryStorage) SignedURL(ctx context.Context, path string, opts ...SignedURLOption) (string, error) {
	return s.URL(ctx, path)
}

// Clear removes all files from memory.
func (s *MemoryStorage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.files = make(map[string]*memoryFile)
}

// Size returns the number of files in memory.
func (s *MemoryStorage) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.files)
}

// TotalBytes returns the total size of all files in memory.
func (s *MemoryStorage) TotalBytes() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var total int64
	for _, f := range s.files {
		total += int64(len(f.data))
	}
	return total
}

// GetBytes returns the raw bytes of a file (useful for testing).
func (s *MemoryStorage) GetBytes(path string) ([]byte, error) {
	path = NormalizePath(path)

	s.mu.RLock()
	defer s.mu.RUnlock()

	file, exists := s.files[path]
	if !exists {
		return nil, ErrNotFound
	}

	result := make([]byte, len(file.data))
	copy(result, file.data)
	return result, nil
}

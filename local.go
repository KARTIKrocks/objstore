package objstore

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalConfig holds configuration for local filesystem storage.
type LocalConfig struct {
	// BasePath is the root directory for storage.
	BasePath string

	// BaseURL is the public URL prefix for serving files.
	BaseURL string

	// CreateDirs automatically creates directories when uploading.
	CreateDirs bool

	// Permissions for created files (default: 0644).
	FilePermissions os.FileMode

	// Permissions for created directories (default: 0755).
	DirPermissions os.FileMode
}

// DefaultLocalConfig returns a default local storage configuration.
func DefaultLocalConfig() LocalConfig {
	return LocalConfig{
		BasePath:        "./storage",
		BaseURL:         "",
		CreateDirs:      true,
		FilePermissions: 0644,
		DirPermissions:  0755,
	}
}

// WithBasePath returns a new config with the specified base path.
func (c LocalConfig) WithBasePath(path string) LocalConfig {
	c.BasePath = path
	return c
}

// WithBaseURL returns a new config with the specified base URL.
func (c LocalConfig) WithBaseURL(url string) LocalConfig {
	c.BaseURL = url
	return c
}

// WithCreateDirs returns a new config with auto directory creation.
func (c LocalConfig) WithCreateDirs(create bool) LocalConfig {
	c.CreateDirs = create
	return c
}

// WithPermissions returns a new config with the specified permissions.
func (c LocalConfig) WithPermissions(file, dir os.FileMode) LocalConfig {
	c.FilePermissions = file
	c.DirPermissions = dir
	return c
}

// LocalStorage implements Storage for local filesystem.
type LocalStorage struct {
	config LocalConfig
}

// NewLocalStorage creates a new local filesystem storage.
func NewLocalStorage(config LocalConfig) (*LocalStorage, error) {
	// Ensure base path is absolute
	absPath, err := filepath.Abs(config.BasePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}
	config.BasePath = absPath

	// Create base directory if it doesn't exist
	if config.CreateDirs {
		if err := os.MkdirAll(config.BasePath, config.DirPermissions); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrPermission, err)
		}
	}

	return &LocalStorage{config: config}, nil
}

// Close is a no-op for local storage.
func (s *LocalStorage) Close() error {
	return nil
}

// Put uploads content to the local filesystem.
func (s *LocalStorage) Put(ctx context.Context, path string, reader io.Reader, opts ...PutOption) (*FileInfo, error) {
	options := ApplyPutOptions(opts)

	fullPath, err := s.fullPath(path)
	if err != nil {
		return nil, err
	}

	// Check if file exists
	if !options.Overwrite {
		if _, err := os.Stat(fullPath); err == nil {
			return nil, ErrAlreadyExists
		}
	}

	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if s.config.CreateDirs {
		if err := os.MkdirAll(dir, s.config.DirPermissions); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrPermission, err)
		}
	}

	// Create file
	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, s.config.FilePermissions)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPermission, err)
	}

	// Copy content
	size, err := io.Copy(file, reader)
	if err != nil {
		_ = file.Close()
		_ = os.Remove(fullPath)
		return nil, err
	}

	if err := file.Close(); err != nil {
		_ = os.Remove(fullPath)
		return nil, err
	}

	// Detect content type
	contentType := options.ContentType
	if contentType == "" {
		contentType = DetectContentType(path)
	}

	return &FileInfo{
		Path:         path,
		Name:         filepath.Base(path),
		Size:         size,
		ContentType:  contentType,
		LastModified: time.Now(),
		Metadata:     options.Metadata,
	}, nil
}

// Get retrieves content from the local filesystem.
func (s *LocalStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath, err := s.fullPath(path)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrPermission, err)
	}

	return file, nil
}

// Delete removes a file from the local filesystem.
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath, err := s.fullPath(path)
	if err != nil {
		return err
	}

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return fmt.Errorf("%w: %v", ErrPermission, err)
	}

	return nil
}

// Exists checks if a file exists on the local filesystem.
func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath, err := s.fullPath(path)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Stat returns file information.
func (s *LocalStorage) Stat(ctx context.Context, path string) (*FileInfo, error) {
	fullPath, err := s.fullPath(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &FileInfo{
		Path:         path,
		Name:         info.Name(),
		Size:         info.Size(),
		ContentType:  DetectContentType(path),
		LastModified: info.ModTime(),
		IsDir:        info.IsDir(),
	}, nil
}

// List returns files matching the prefix.
func (s *LocalStorage) List(ctx context.Context, prefix string, opts ...ListOption) (*ListResult, error) {
	options := ApplyListOptions(opts)

	searchPath, err := s.fullPath(prefix)
	if err != nil {
		return nil, err
	}

	searchPath, prefix = s.resolveSearchPath(searchPath, prefix)

	result := &ListResult{
		Files:    make([]*FileInfo, 0),
		Prefixes: make([]string, 0),
	}

	prefixMap := make(map[string]bool)

	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk error at %s: %w", path, walkErr)
		}

		relPath, ok := s.listRelPath(path, searchPath, prefix)
		if !ok {
			return nil
		}

		if skip := s.handleNestedEntry(relPath, prefix, info, options, prefixMap, result); skip {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		// Check max keys
		if options.MaxKeys > 0 && len(result.Files) >= options.MaxKeys {
			result.IsTruncated = true
			return filepath.SkipAll
		}

		result.Files = append(result.Files, &FileInfo{
			Path:         relPath,
			Name:         info.Name(),
			Size:         info.Size(),
			ContentType:  DetectContentType(relPath),
			LastModified: info.ModTime(),
			IsDir:        info.IsDir(),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// resolveSearchPath ensures the search path points to an existing directory.
func (s *LocalStorage) resolveSearchPath(searchPath, prefix string) (string, string) {
	info, err := os.Stat(searchPath)
	if err != nil {
		if os.IsNotExist(err) {
			return filepath.Dir(searchPath), filepath.Dir(prefix)
		}
		return searchPath, prefix
	}
	if !info.IsDir() {
		return filepath.Dir(searchPath), filepath.Dir(prefix)
	}
	return searchPath, prefix
}

// listRelPath returns the relative path for a walk entry, or false if it should be skipped.
func (s *LocalStorage) listRelPath(path, searchPath, prefix string) (string, bool) {
	if path == searchPath {
		return "", false
	}

	relPath, err := filepath.Rel(s.config.BasePath, path)
	if err != nil {
		return "", false
	}
	relPath = filepath.ToSlash(relPath)

	if !strings.HasPrefix(relPath, strings.TrimPrefix(prefix, "/")) && relPath != "." {
		return "", false
	}

	return relPath, true
}

// handleNestedEntry checks if an entry is nested beyond the delimiter and adds it as a prefix.
// Returns true if the entry was handled (should be skipped by the caller).
func (s *LocalStorage) handleNestedEntry(relPath, prefix string, info os.FileInfo, options *ListOptions, prefixMap map[string]bool, result *ListResult) bool {
	if options.Recursive || options.Delimiter == "" {
		return false
	}

	relToPrefix := strings.TrimPrefix(relPath, strings.TrimPrefix(prefix, "/"))
	relToPrefix = strings.TrimPrefix(relToPrefix, "/")

	if !strings.Contains(relToPrefix, options.Delimiter) {
		return false
	}

	parts := strings.SplitN(relToPrefix, options.Delimiter, 2)
	prefixPath := filepath.Join(prefix, parts[0]) + "/"
	if !prefixMap[prefixPath] {
		prefixMap[prefixPath] = true
		result.Prefixes = append(result.Prefixes, prefixPath)
	}
	return true
}

// Copy copies a file.
func (s *LocalStorage) Copy(ctx context.Context, src, dst string) error {
	srcPath, err := s.fullPath(src)
	if err != nil {
		return err
	}

	dstPath, err := s.fullPath(dst)
	if err != nil {
		return err
	}

	// Open source
	srcFile, err := os.Open(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return err
	}
	defer func() { _ = srcFile.Close() }()

	// Create destination directory
	if s.config.CreateDirs {
		if err := os.MkdirAll(filepath.Dir(dstPath), s.config.DirPermissions); err != nil {
			return fmt.Errorf("%w: %v", ErrPermission, err)
		}
	}

	// Create destination
	dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, s.config.FilePermissions)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPermission, err)
	}

	// Copy
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		_ = dstFile.Close()
		return err
	}

	return dstFile.Close()
}

// Move moves a file.
func (s *LocalStorage) Move(ctx context.Context, src, dst string) error {
	srcPath, err := s.fullPath(src)
	if err != nil {
		return err
	}

	dstPath, err := s.fullPath(dst)
	if err != nil {
		return err
	}

	// Create destination directory
	if s.config.CreateDirs {
		if err := os.MkdirAll(filepath.Dir(dstPath), s.config.DirPermissions); err != nil {
			return fmt.Errorf("%w: %v", ErrPermission, err)
		}
	}

	// Try rename first (faster if same filesystem)
	if err := os.Rename(srcPath, dstPath); err != nil {
		// If rename fails (cross-device), copy and delete
		if err := s.Copy(ctx, src, dst); err != nil {
			return err
		}
		return s.Delete(ctx, src)
	}

	return nil
}

// URL returns a public URL for the file.
func (s *LocalStorage) URL(ctx context.Context, path string) (string, error) {
	if s.config.BaseURL == "" {
		return "", ErrNotImplemented
	}

	return strings.TrimSuffix(s.config.BaseURL, "/") + "/" + strings.TrimPrefix(path, "/"), nil
}

// SignedURL returns a signed URL (not supported for local storage).
func (s *LocalStorage) SignedURL(ctx context.Context, path string, opts ...SignedURLOption) (string, error) {
	return s.URL(ctx, path)
}

// fullPath returns the full filesystem path, validating against path traversal.
func (s *LocalStorage) fullPath(path string) (string, error) {
	// Clean and normalize path
	path = filepath.Clean(path)
	path = strings.TrimPrefix(path, "/")

	// Build full path
	fullPath := filepath.Join(s.config.BasePath, path)

	// Ensure path is within base directory (prevent path traversal).
	// Use separator suffix to prevent /tmp/store matching /tmp/storevil.
	if fullPath != s.config.BasePath && !strings.HasPrefix(fullPath, s.config.BasePath+string(os.PathSeparator)) {
		return "", ErrInvalidPath
	}

	return fullPath, nil
}

// DeleteDir removes a directory and all its contents.
func (s *LocalStorage) DeleteDir(ctx context.Context, path string) error {
	fullPath, err := s.fullPath(path)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(fullPath); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return fmt.Errorf("%w: %v", ErrPermission, err)
	}

	return nil
}

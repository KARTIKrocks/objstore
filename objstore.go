// Package objstore provides a unified interface for file storage operations
// across different backends including local filesystem, AWS S3, Google Cloud Storage, and Azure Blob Storage.
package objstore

import (
	"context"
	"errors"
	"io"
	"time"
)

// Sentinel errors for storage operations.
var (
	ErrNotFound       = errors.New("objstore: file not found")
	ErrAlreadyExists  = errors.New("objstore: file already exists")
	ErrInvalidPath    = errors.New("objstore: invalid path")
	ErrPermission     = errors.New("objstore: permission denied")
	ErrNotImplemented = errors.New("objstore: operation not implemented")
	ErrInvalidConfig  = errors.New("objstore: invalid configuration")
)

// Storage defines the interface for storage backends.
type Storage interface {
	// Put uploads content to the specified path.
	Put(ctx context.Context, path string, reader io.Reader, opts ...PutOption) (*FileInfo, error)

	// Get retrieves content from the specified path.
	Get(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes the file at the specified path.
	Delete(ctx context.Context, path string) error

	// Exists checks if a file exists at the specified path.
	Exists(ctx context.Context, path string) (bool, error)

	// Stat returns file information without downloading content.
	Stat(ctx context.Context, path string) (*FileInfo, error)

	// List returns files matching the prefix.
	List(ctx context.Context, prefix string, opts ...ListOption) (*ListResult, error)

	// Copy copies a file from src to dst.
	Copy(ctx context.Context, src, dst string) error

	// Move moves a file from src to dst.
	Move(ctx context.Context, src, dst string) error

	// URL returns a public URL for the file (if supported).
	URL(ctx context.Context, path string) (string, error)

	// SignedURL returns a pre-signed URL for temporary access.
	SignedURL(ctx context.Context, path string, opts ...SignedURLOption) (string, error)

	// Close releases any resources held by the storage backend.
	Close() error
}

// BatchDeleter is an optional interface for backends that support batch deletion.
type BatchDeleter interface {
	DeleteMultiple(ctx context.Context, paths []string) error
}

// FileInfo contains information about a stored file.
type FileInfo struct {
	Path         string            // Full path including filename
	Name         string            // Filename only
	Size         int64             // Size in bytes
	ContentType  string            // MIME type
	ETag         string            // Entity tag for versioning
	LastModified time.Time         // Last modification time
	Metadata     map[string]string // Custom metadata
	IsDir        bool              // True if this is a directory/prefix
}

// ListResult contains the result of a list operation.
type ListResult struct {
	Files       []*FileInfo // Files matching the prefix
	Prefixes    []string    // Common prefixes (directories)
	NextToken   string      // Token for pagination
	IsTruncated bool        // True if there are more results
}

// PutOptions configures upload behavior.
type PutOptions struct {
	ContentType  string            // MIME type (auto-detected if empty)
	Metadata     map[string]string // Custom metadata
	CacheControl string            // Cache-Control header
	ACL          string            // Access control (e.g., "public-read")
	Overwrite    bool              // Allow overwriting existing files
}

// PutOption is a function that modifies PutOptions.
type PutOption func(*PutOptions)

// WithContentType sets the content type.
func WithContentType(contentType string) PutOption {
	return func(o *PutOptions) {
		o.ContentType = contentType
	}
}

// WithMetadata sets custom metadata.
func WithMetadata(metadata map[string]string) PutOption {
	return func(o *PutOptions) {
		o.Metadata = metadata
	}
}

// WithCacheControl sets the Cache-Control header.
func WithCacheControl(cacheControl string) PutOption {
	return func(o *PutOptions) {
		o.CacheControl = cacheControl
	}
}

// WithACL sets the access control.
func WithACL(acl string) PutOption {
	return func(o *PutOptions) {
		o.ACL = acl
	}
}

// WithOverwrite allows overwriting existing files.
func WithOverwrite(overwrite bool) PutOption {
	return func(o *PutOptions) {
		o.Overwrite = overwrite
	}
}

// ListOptions configures list behavior.
type ListOptions struct {
	MaxKeys   int    // Maximum number of results
	Delimiter string // Delimiter for grouping (e.g., "/")
	Token     string // Continuation token for pagination
	Recursive bool   // List recursively (ignore delimiter)
}

// ListOption is a function that modifies ListOptions.
type ListOption func(*ListOptions)

// WithMaxKeys sets the maximum number of results.
func WithMaxKeys(maxKeys int) ListOption {
	return func(o *ListOptions) {
		o.MaxKeys = maxKeys
	}
}

// WithDelimiter sets the delimiter for grouping.
func WithDelimiter(delimiter string) ListOption {
	return func(o *ListOptions) {
		o.Delimiter = delimiter
	}
}

// WithToken sets the continuation token.
func WithToken(token string) ListOption {
	return func(o *ListOptions) {
		o.Token = token
	}
}

// WithRecursive enables recursive listing.
func WithRecursive(recursive bool) ListOption {
	return func(o *ListOptions) {
		o.Recursive = recursive
	}
}

// SignedURLOptions configures signed URL generation.
type SignedURLOptions struct {
	Expires     time.Duration     // URL expiration time
	Method      string            // HTTP method (GET, PUT)
	ContentType string            // Content-Type for PUT requests
	Headers     map[string]string // Additional headers
}

// SignedURLOption is a function that modifies SignedURLOptions.
type SignedURLOption func(*SignedURLOptions)

// WithExpires sets the URL expiration time.
func WithExpires(expires time.Duration) SignedURLOption {
	return func(o *SignedURLOptions) {
		o.Expires = expires
	}
}

// WithMethod sets the HTTP method for the signed URL.
func WithMethod(method string) SignedURLOption {
	return func(o *SignedURLOptions) {
		o.Method = method
	}
}

// WithSignedContentType sets the content type for PUT signed URLs.
func WithSignedContentType(contentType string) SignedURLOption {
	return func(o *SignedURLOptions) {
		o.ContentType = contentType
	}
}

// WithHeaders sets additional headers for signed URLs.
func WithHeaders(headers map[string]string) SignedURLOption {
	return func(o *SignedURLOptions) {
		o.Headers = headers
	}
}

// ApplyPutOptions applies PutOption functions to PutOptions.
func ApplyPutOptions(opts []PutOption) *PutOptions {
	options := &PutOptions{
		Overwrite: true, // Default to overwrite
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// ApplyListOptions applies ListOption functions to ListOptions.
func ApplyListOptions(opts []ListOption) *ListOptions {
	options := &ListOptions{
		MaxKeys:   1000,
		Delimiter: "/",
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// ApplySignedURLOptions applies SignedURLOption functions to SignedURLOptions.
func ApplySignedURLOptions(opts []SignedURLOption) *SignedURLOptions {
	options := &SignedURLOptions{
		Expires: 15 * time.Minute,
		Method:  "GET",
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

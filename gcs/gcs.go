// Package gcs provides a Google Cloud Storage backend for objstore.
package gcs

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/KARTIKrocks/objstore"

	storage "cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Config holds configuration for Google Cloud Storage.
type Config struct {
	// Bucket is the GCS bucket name.
	Bucket string

	// CredentialsFile is the path to the credentials JSON file.
	CredentialsFile string

	// CredentialsJSON is the credentials JSON content.
	CredentialsJSON []byte

	// CredentialsType specifies the type of credentials being provided
	// (e.g., option.ServiceAccount, option.AuthorizedUser).
	// Defaults to option.ServiceAccount if not set.
	CredentialsType option.CredentialsType

	// Prefix is a path prefix for all operations.
	Prefix string

	// BaseURL is the public URL prefix for serving files.
	BaseURL string

	// DefaultACL is the default ACL for uploaded files.
	DefaultACL string
}

// DefaultConfig returns a default GCS configuration.
func DefaultConfig() Config {
	return Config{
		DefaultACL:      "private",
		CredentialsType: option.ServiceAccount,
	}
}

// WithBucket returns a new config with the specified bucket.
func (c Config) WithBucket(bucket string) Config {
	c.Bucket = bucket
	return c
}

// WithCredentialsFile returns a new config with the specified credentials file.
func (c Config) WithCredentialsFile(path string) Config {
	c.CredentialsFile = path
	return c
}

// WithCredentialsJSON returns a new config with the specified credentials JSON.
func (c Config) WithCredentialsJSON(json []byte) Config {
	c.CredentialsJSON = json
	return c
}

// WithCredentialsType returns a new config with the specified credentials type
// (e.g., option.ServiceAccount, option.AuthorizedUser).
func (c Config) WithCredentialsType(credType option.CredentialsType) Config {
	c.CredentialsType = credType
	return c
}

// WithPrefix returns a new config with the specified prefix.
func (c Config) WithPrefix(prefix string) Config {
	c.Prefix = prefix
	return c
}

// WithBaseURL returns a new config with the specified base URL.
func (c Config) WithBaseURL(url string) Config {
	c.BaseURL = url
	return c
}

// WithDefaultACL returns a new config with the specified default ACL.
func (c Config) WithDefaultACL(acl string) Config {
	c.DefaultACL = acl
	return c
}

// Storage implements objstore.Storage for Google Cloud Storage.
type Storage struct {
	client *storage.Client
	bucket *storage.BucketHandle
	config Config
}

// New creates a new GCS storage.
func New(ctx context.Context, cfg Config) (*Storage, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("%w: bucket is required", objstore.ErrInvalidConfig)
	}

	credType := cfg.CredentialsType
	if credType == "" {
		credType = option.ServiceAccount
	}

	var clientOpts []option.ClientOption

	if cfg.CredentialsFile != "" {
		clientOpts = append(clientOpts, option.WithAuthCredentialsFile(credType, cfg.CredentialsFile))
	} else if len(cfg.CredentialsJSON) > 0 {
		clientOpts = append(clientOpts, option.WithAuthCredentialsJSON(credType, cfg.CredentialsJSON))
	}

	client, err := storage.NewClient(ctx, clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", objstore.ErrInvalidConfig, err)
	}

	return &Storage{
		client: client,
		bucket: client.Bucket(cfg.Bucket),
		config: cfg,
	}, nil
}

// Close closes the GCS client.
func (st *Storage) Close() error {
	return st.client.Close()
}

// Put uploads content to GCS.
func (st *Storage) Put(ctx context.Context, path string, reader io.Reader, opts ...objstore.PutOption) (*objstore.FileInfo, error) {
	options := objstore.ApplyPutOptions(opts)

	objectName := st.objectName(path)

	// Check if file exists
	if !options.Overwrite {
		exists, err := st.Exists(ctx, path)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, objstore.ErrAlreadyExists
		}
	}

	obj := st.bucket.Object(objectName)
	writer := obj.NewWriter(ctx)

	// Set content type
	if options.ContentType != "" {
		writer.ContentType = options.ContentType
	} else {
		writer.ContentType = objstore.DetectContentType(path)
	}

	// Set cache control
	if options.CacheControl != "" {
		writer.CacheControl = options.CacheControl
	}

	// Set metadata
	if len(options.Metadata) > 0 {
		writer.Metadata = options.Metadata
	}

	// Set ACL
	if options.ACL != "" {
		writer.PredefinedACL = options.ACL
	} else if st.config.DefaultACL != "" {
		writer.PredefinedACL = st.config.DefaultACL
	}

	// Upload
	size, err := io.Copy(writer, reader)
	if err != nil {
		_ = writer.Close()
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return &objstore.FileInfo{
		Path:         path,
		Name:         filepath.Base(path),
		Size:         size,
		ContentType:  writer.ContentType,
		LastModified: time.Now(),
		Metadata:     options.Metadata,
	}, nil
}

// Get retrieves content from GCS.
func (st *Storage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	objectName := st.objectName(path)

	reader, err := st.bucket.Object(objectName).NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, objstore.ErrNotFound
		}
		return nil, err
	}

	return reader, nil
}

// Delete removes a file from GCS.
func (st *Storage) Delete(ctx context.Context, path string) error {
	objectName := st.objectName(path)

	if err := st.bucket.Object(objectName).Delete(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			return objstore.ErrNotFound
		}
		return err
	}

	return nil
}

// Exists checks if a file exists in GCS.
func (st *Storage) Exists(ctx context.Context, path string) (bool, error) {
	objectName := st.objectName(path)

	_, err := st.bucket.Object(objectName).Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Stat returns file information from GCS.
func (st *Storage) Stat(ctx context.Context, path string) (*objstore.FileInfo, error) {
	objectName := st.objectName(path)

	attrs, err := st.bucket.Object(objectName).Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, objstore.ErrNotFound
		}
		return nil, err
	}

	return &objstore.FileInfo{
		Path:         path,
		Name:         filepath.Base(path),
		Size:         attrs.Size,
		ContentType:  attrs.ContentType,
		ETag:         attrs.Etag,
		LastModified: attrs.Updated,
		Metadata:     attrs.Metadata,
	}, nil
}

// List returns files matching the prefix in GCS.
func (st *Storage) List(ctx context.Context, prefix string, opts ...objstore.ListOption) (*objstore.ListResult, error) {
	options := objstore.ApplyListOptions(opts)

	query := &storage.Query{
		Prefix: st.objectName(prefix),
	}

	if !options.Recursive && options.Delimiter != "" {
		query.Delimiter = options.Delimiter
	}

	result := &objstore.ListResult{
		Files:    make([]*objstore.FileInfo, 0),
		Prefixes: make([]string, 0),
	}

	it := st.bucket.Objects(ctx, query)
	count := 0

	for {
		if options.MaxKeys > 0 && count >= options.MaxKeys {
			result.IsTruncated = true
			break
		}

		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		// Handle common prefixes (directories)
		if attrs.Prefix != "" {
			result.Prefixes = append(result.Prefixes, st.stripPrefix(attrs.Prefix))
			continue
		}

		path := st.stripPrefix(attrs.Name)

		result.Files = append(result.Files, &objstore.FileInfo{
			Path:         path,
			Name:         filepath.Base(path),
			Size:         attrs.Size,
			ContentType:  attrs.ContentType,
			ETag:         attrs.Etag,
			LastModified: attrs.Updated,
			Metadata:     attrs.Metadata,
		})

		count++
	}

	return result, nil
}

// Copy copies a file in GCS.
func (st *Storage) Copy(ctx context.Context, src, dst string) error {
	srcObj := st.bucket.Object(st.objectName(src))
	dstObj := st.bucket.Object(st.objectName(dst))

	if _, err := dstObj.CopierFrom(srcObj).Run(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			return objstore.ErrNotFound
		}
		return err
	}

	return nil
}

// Move moves a file in GCS.
func (st *Storage) Move(ctx context.Context, src, dst string) error {
	if err := st.Copy(ctx, src, dst); err != nil {
		return err
	}
	return st.Delete(ctx, src)
}

// URL returns a public URL for the file.
func (st *Storage) URL(ctx context.Context, path string) (string, error) {
	if st.config.BaseURL != "" {
		return strings.TrimSuffix(st.config.BaseURL, "/") + "/" + strings.TrimPrefix(path, "/"), nil
	}

	objectName := st.objectName(path)
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", st.config.Bucket, objectName), nil
}

// SignedURL returns a pre-signed URL for temporary access.
func (st *Storage) SignedURL(ctx context.Context, path string, opts ...objstore.SignedURLOption) (string, error) {
	options := objstore.ApplySignedURLOptions(opts)

	objectName := st.objectName(path)

	signedURLOpts := &storage.SignedURLOptions{
		Method:  options.Method,
		Expires: time.Now().Add(options.Expires),
	}

	if options.ContentType != "" {
		signedURLOpts.ContentType = options.ContentType
	}

	return st.bucket.SignedURL(objectName, signedURLOpts)
}

// objectName returns the full object name with prefix.
func (st *Storage) objectName(path string) string {
	path = strings.TrimPrefix(path, "/")
	if st.config.Prefix != "" {
		return strings.TrimSuffix(st.config.Prefix, "/") + "/" + path
	}
	return path
}

// stripPrefix removes the configured prefix from an object name.
func (st *Storage) stripPrefix(name string) string {
	if st.config.Prefix != "" {
		prefix := strings.TrimSuffix(st.config.Prefix, "/") + "/"
		return strings.TrimPrefix(name, prefix)
	}
	return name
}

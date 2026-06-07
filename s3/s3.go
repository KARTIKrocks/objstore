// Package s3 provides an AWS S3 storage backend for objstore.
package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/KARTIKrocks/objstore"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithy "github.com/aws/smithy-go"
)

// Config holds configuration for S3 storage.
type Config struct {
	// Bucket is the S3 bucket name.
	Bucket string

	// Region is the AWS region.
	Region string

	// AccessKeyID is the AWS access key (optional if using IAM roles).
	AccessKeyID string

	// SecretAccessKey is the AWS secret key (optional if using IAM roles).
	SecretAccessKey string

	// Endpoint is a custom endpoint for S3-compatible services (e.g., MinIO).
	Endpoint string

	// Prefix is a path prefix for all operations.
	Prefix string

	// UsePathStyle uses path-style URLs instead of virtual-hosted-style.
	UsePathStyle bool

	// BaseURL is the public URL prefix for serving files.
	BaseURL string

	// DefaultACL is the default ACL for uploaded files.
	DefaultACL string
}

// DefaultConfig returns a default S3 configuration.
func DefaultConfig() Config {
	return Config{
		Region:     "us-east-1",
		DefaultACL: "private",
	}
}

// WithBucket returns a new config with the specified bucket.
func (c Config) WithBucket(bucket string) Config {
	c.Bucket = bucket
	return c
}

// WithRegion returns a new config with the specified region.
func (c Config) WithRegion(region string) Config {
	c.Region = region
	return c
}

// WithCredentials returns a new config with the specified credentials.
func (c Config) WithCredentials(accessKeyID, secretAccessKey string) Config {
	c.AccessKeyID = accessKeyID
	c.SecretAccessKey = secretAccessKey
	return c
}

// WithEndpoint returns a new config with a custom endpoint.
func (c Config) WithEndpoint(endpoint string) Config {
	c.Endpoint = endpoint
	return c
}

// WithPrefix returns a new config with the specified prefix.
func (c Config) WithPrefix(prefix string) Config {
	c.Prefix = prefix
	return c
}

// WithPathStyle returns a new config with path-style URLs enabled.
func (c Config) WithPathStyle(pathStyle bool) Config {
	c.UsePathStyle = pathStyle
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

// Storage implements objstore.Storage for AWS S3.
type Storage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	config        Config
}

// New creates a new S3 storage.
func New(ctx context.Context, cfg Config) (*Storage, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("%w: bucket is required", objstore.ErrInvalidConfig)
	}

	// Build AWS config options
	var awsOpts []func(*config.LoadOptions) error

	awsOpts = append(awsOpts, config.WithRegion(cfg.Region))

	// Add credentials if provided
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		awsOpts = append(awsOpts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(ctx, awsOpts...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", objstore.ErrInvalidConfig, err)
	}

	// Build S3 client options
	var s3Opts []func(*s3.Options)

	if cfg.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
	}

	if cfg.UsePathStyle {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	}

	// Create S3 client
	client := s3.NewFromConfig(awsCfg, s3Opts...)

	return &Storage{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		config:        cfg,
	}, nil
}

// Close is a no-op for S3 storage (the HTTP client is managed by the SDK).
func (st *Storage) Close() error {
	return nil
}

// Put uploads content to S3.
func (st *Storage) Put(ctx context.Context, path string, reader io.Reader, opts ...objstore.PutOption) (*objstore.FileInfo, error) {
	options := objstore.ApplyPutOptions(opts)

	key := st.key(path)

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

	// Detect content type
	contentType := options.ContentType
	if contentType == "" {
		contentType = objstore.DetectContentType(path)
	}

	// Build input
	input := &s3.PutObjectInput{
		Bucket:      aws.String(st.config.Bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
	}

	// Set ACL
	acl := options.ACL
	if acl == "" {
		acl = st.config.DefaultACL
	}
	if acl != "" {
		input.ACL = types.ObjectCannedACL(acl)
	}

	// Set cache control
	if options.CacheControl != "" {
		input.CacheControl = aws.String(options.CacheControl)
	}

	// Set metadata
	if len(options.Metadata) > 0 {
		input.Metadata = options.Metadata
	}

	// Upload
	result, err := st.client.PutObject(ctx, input)
	if err != nil {
		return nil, err
	}

	etag := ""
	if result.ETag != nil {
		etag = strings.Trim(*result.ETag, "\"")
	}

	return &objstore.FileInfo{
		Path:         path,
		Name:         filepath.Base(path),
		ContentType:  contentType,
		ETag:         etag,
		LastModified: time.Now(),
		Metadata:     options.Metadata,
	}, nil
}

// Get retrieves content from S3.
func (st *Storage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	key := st.key(path)

	result, err := st.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(st.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFoundError(err) {
			return nil, objstore.ErrNotFound
		}
		return nil, err
	}

	return result.Body, nil
}

// Delete removes a file from S3.
func (st *Storage) Delete(ctx context.Context, path string) error {
	key := st.key(path)

	_, err := st.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(st.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFoundError(err) {
			return objstore.ErrNotFound
		}
		return err
	}

	return nil
}

// Exists checks if a file exists in S3.
func (st *Storage) Exists(ctx context.Context, path string) (bool, error) {
	key := st.key(path)

	_, err := st.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(st.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Stat returns file information from S3.
func (st *Storage) Stat(ctx context.Context, path string) (*objstore.FileInfo, error) {
	key := st.key(path)

	result, err := st.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(st.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFoundError(err) {
			return nil, objstore.ErrNotFound
		}
		return nil, err
	}

	etag := ""
	if result.ETag != nil {
		etag = strings.Trim(*result.ETag, "\"")
	}

	contentType := "application/octet-stream"
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	var lastModified time.Time
	if result.LastModified != nil {
		lastModified = *result.LastModified
	}

	var size int64
	if result.ContentLength != nil {
		size = *result.ContentLength
	}

	return &objstore.FileInfo{
		Path:         path,
		Name:         filepath.Base(path),
		Size:         size,
		ContentType:  contentType,
		ETag:         etag,
		LastModified: lastModified,
		Metadata:     result.Metadata,
	}, nil
}

// List returns files matching the prefix in S3.
func (st *Storage) List(ctx context.Context, prefix string, opts ...objstore.ListOption) (*objstore.ListResult, error) {
	options := objstore.ApplyListOptions(opts)

	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(st.config.Bucket),
		Prefix:  aws.String(st.key(prefix)),
		MaxKeys: aws.Int32(int32(options.MaxKeys)),
	}

	if !options.Recursive && options.Delimiter != "" {
		input.Delimiter = aws.String(options.Delimiter)
	}

	if options.Token != "" {
		input.ContinuationToken = aws.String(options.Token)
	}

	result, err := st.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, err
	}

	listResult := &objstore.ListResult{
		Files:    make([]*objstore.FileInfo, 0, len(result.Contents)),
		Prefixes: make([]string, 0, len(result.CommonPrefixes)),
	}

	// Process files
	for _, obj := range result.Contents {
		if obj.Key == nil {
			continue
		}

		path := st.stripPrefix(*obj.Key)

		var lastModified time.Time
		if obj.LastModified != nil {
			lastModified = *obj.LastModified
		}

		var size int64
		if obj.Size != nil {
			size = *obj.Size
		}

		etag := ""
		if obj.ETag != nil {
			etag = strings.Trim(*obj.ETag, "\"")
		}

		listResult.Files = append(listResult.Files, &objstore.FileInfo{
			Path:         path,
			Name:         filepath.Base(path),
			Size:         size,
			ContentType:  objstore.DetectContentType(path),
			ETag:         etag,
			LastModified: lastModified,
		})
	}

	// Process common prefixes
	for _, p := range result.CommonPrefixes {
		if p.Prefix != nil {
			listResult.Prefixes = append(listResult.Prefixes, st.stripPrefix(*p.Prefix))
		}
	}

	// Set pagination info
	if result.IsTruncated != nil {
		listResult.IsTruncated = *result.IsTruncated
	}
	if result.NextContinuationToken != nil {
		listResult.NextToken = *result.NextContinuationToken
	}

	return listResult, nil
}

// Copy copies a file in S3.
func (st *Storage) Copy(ctx context.Context, src, dst string) error {
	srcKey := st.key(src)
	dstKey := st.key(dst)

	copySource := fmt.Sprintf("%s/%s", st.config.Bucket, url.PathEscape(srcKey))

	_, err := st.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(st.config.Bucket),
		CopySource: aws.String(copySource),
		Key:        aws.String(dstKey),
	})
	if err != nil {
		if isNotFoundError(err) {
			return objstore.ErrNotFound
		}
		return err
	}

	return nil
}

// Move moves a file in S3.
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

	// Build default S3 URL
	key := st.key(path)
	if st.config.Endpoint != "" {
		return fmt.Sprintf("%s/%s/%s", strings.TrimSuffix(st.config.Endpoint, "/"), st.config.Bucket, key), nil
	}

	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", st.config.Bucket, st.config.Region, key), nil
}

// SignedURL returns a pre-signed URL for temporary access.
func (st *Storage) SignedURL(ctx context.Context, path string, opts ...objstore.SignedURLOption) (string, error) {
	options := objstore.ApplySignedURLOptions(opts)

	key := st.key(path)

	if options.Method == "PUT" {
		input := &s3.PutObjectInput{
			Bucket: aws.String(st.config.Bucket),
			Key:    aws.String(key),
		}
		if options.ContentType != "" {
			input.ContentType = aws.String(options.ContentType)
		}

		result, err := st.presignClient.PresignPutObject(ctx, input, s3.WithPresignExpires(options.Expires))
		if err != nil {
			return "", err
		}
		return result.URL, nil
	}

	// Default to GET
	result, err := st.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(st.config.Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(options.Expires))
	if err != nil {
		return "", err
	}

	return result.URL, nil
}

// DeleteMultiple deletes multiple files in S3.
// This implements the objstore.BatchDeleter interface.
func (st *Storage) DeleteMultiple(ctx context.Context, paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	objects := make([]types.ObjectIdentifier, len(paths))
	for i, path := range paths {
		key := st.key(path)
		objects[i] = types.ObjectIdentifier{Key: aws.String(key)}
	}

	_, err := st.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(st.config.Bucket),
		Delete: &types.Delete{Objects: objects},
	})

	return err
}

// key returns the full S3 key with prefix.
func (st *Storage) key(path string) string {
	path = strings.TrimPrefix(path, "/")
	if st.config.Prefix != "" {
		return strings.TrimSuffix(st.config.Prefix, "/") + "/" + path
	}
	return path
}

// stripPrefix removes the configured prefix from a key.
func (st *Storage) stripPrefix(key string) string {
	if st.config.Prefix != "" {
		prefix := strings.TrimSuffix(st.config.Prefix, "/") + "/"
		return strings.TrimPrefix(key, prefix)
	}
	return key
}

// isNotFoundError checks if an error is a not found error using proper AWS SDK error types.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Check for NoSuchKey
	if _, ok := errors.AsType[*types.NoSuchKey](err); ok {
		return true
	}

	// Check for NotFound (returned by HeadObject)
	if _, ok := errors.AsType[*types.NotFound](err); ok {
		return true
	}

	// Check for generic API error with 404 status code
	if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
		return apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "NoSuchKey"
	}

	return false
}

// Package azure provides an Azure Blob Storage backend for objstore.
package azure

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/KARTIKrocks/objstore"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
)

// Config holds configuration for Azure Blob Storage.
type Config struct {
	// AccountName is the Azure storage account name.
	AccountName string

	// AccountKey is the Azure storage account key (optional if using other auth).
	AccountKey string

	// ConnectionString is an Azure storage connection string (optional).
	ConnectionString string

	// ContainerName is the blob container name.
	ContainerName string

	// Prefix is a path prefix for all operations.
	Prefix string

	// BaseURL is the public URL prefix for serving files.
	BaseURL string
}

// DefaultConfig returns a default Azure configuration.
func DefaultConfig() Config {
	return Config{}
}

// WithAccountName returns a new config with the specified account name.
func (c Config) WithAccountName(name string) Config {
	c.AccountName = name
	return c
}

// WithAccountKey returns a new config with the specified account key.
func (c Config) WithAccountKey(key string) Config {
	c.AccountKey = key
	return c
}

// WithConnectionString returns a new config with the specified connection string.
func (c Config) WithConnectionString(connStr string) Config {
	c.ConnectionString = connStr
	return c
}

// WithContainerName returns a new config with the specified container name.
func (c Config) WithContainerName(name string) Config {
	c.ContainerName = name
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

// Storage implements objstore.Storage for Azure Blob Storage.
type Storage struct {
	client    *azblob.Client
	sharedKey *azblob.SharedKeyCredential
	config    Config
}

// New creates a new Azure Blob storage.
func New(ctx context.Context, cfg Config) (*Storage, error) {
	if cfg.ContainerName == "" {
		return nil, fmt.Errorf("%w: container name is required", objstore.ErrInvalidConfig)
	}

	st := &Storage{config: cfg}

	switch {
	case cfg.ConnectionString != "":
		client, err := azblob.NewClientFromConnectionString(cfg.ConnectionString, nil)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", objstore.ErrInvalidConfig, err)
		}
		st.client = client

	case cfg.AccountName != "" && cfg.AccountKey != "":
		cred, err := azblob.NewSharedKeyCredential(cfg.AccountName, cfg.AccountKey)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", objstore.ErrInvalidConfig, err)
		}
		st.sharedKey = cred

		serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", cfg.AccountName)
		client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", objstore.ErrInvalidConfig, err)
		}
		st.client = client

	case cfg.AccountName != "":
		// Use DefaultAzureCredential (managed identity, env vars, CLI, etc.)
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", objstore.ErrInvalidConfig, err)
		}

		serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", cfg.AccountName)
		client, err := azblob.NewClient(serviceURL, cred, nil)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", objstore.ErrInvalidConfig, err)
		}
		st.client = client

	default:
		return nil, fmt.Errorf("%w: account name or connection string is required", objstore.ErrInvalidConfig)
	}

	return st, nil
}

// Close is a no-op for Azure storage (the HTTP client is managed by the SDK).
func (st *Storage) Close() error {
	return nil
}

// Put uploads content to Azure Blob Storage.
func (st *Storage) Put(ctx context.Context, path string, reader io.Reader, opts ...objstore.PutOption) (*objstore.FileInfo, error) {
	options := objstore.ApplyPutOptions(opts)

	blobName := st.blobName(path)

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

	uploadOpts := &azblob.UploadStreamOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: &contentType,
		},
		Metadata: toAzureMetadata(options.Metadata),
	}

	if options.CacheControl != "" {
		uploadOpts.HTTPHeaders.BlobCacheControl = &options.CacheControl
	}

	_, err := st.client.UploadStream(ctx, st.config.ContainerName, blobName, reader, uploadOpts)
	if err != nil {
		return nil, err
	}

	return &objstore.FileInfo{
		Path:         path,
		Name:         filepath.Base(path),
		ContentType:  contentType,
		LastModified: time.Now(),
		Metadata:     options.Metadata,
	}, nil
}

// Get retrieves content from Azure Blob Storage.
func (st *Storage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	blobName := st.blobName(path)

	resp, err := st.client.DownloadStream(ctx, st.config.ContainerName, blobName, nil)
	if err != nil {
		if isNotFoundError(err) {
			return nil, objstore.ErrNotFound
		}
		return nil, err
	}

	return resp.Body, nil
}

// Delete removes a file from Azure Blob Storage.
func (st *Storage) Delete(ctx context.Context, path string) error {
	blobName := st.blobName(path)

	_, err := st.client.DeleteBlob(ctx, st.config.ContainerName, blobName, nil)
	if err != nil {
		if isNotFoundError(err) {
			return objstore.ErrNotFound
		}
		return err
	}

	return nil
}

// Exists checks if a file exists in Azure Blob Storage.
func (st *Storage) Exists(ctx context.Context, path string) (bool, error) {
	_, err := st.Stat(ctx, path)
	if err != nil {
		if errors.Is(err, objstore.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Stat returns file information from Azure Blob Storage.
func (st *Storage) Stat(ctx context.Context, path string) (*objstore.FileInfo, error) {
	blobName := st.blobName(path)

	blobClient := st.client.ServiceClient().NewContainerClient(st.config.ContainerName).NewBlobClient(blobName)
	props, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		if isNotFoundError(err) {
			return nil, objstore.ErrNotFound
		}
		return nil, err
	}

	info := &objstore.FileInfo{
		Path: path,
		Name: filepath.Base(path),
	}

	if props.ContentLength != nil {
		info.Size = *props.ContentLength
	}
	if props.ContentType != nil {
		info.ContentType = *props.ContentType
	}
	if props.ETag != nil {
		info.ETag = string(*props.ETag)
	}
	if props.LastModified != nil {
		info.LastModified = *props.LastModified
	}
	info.Metadata = fromAzureMetadata(props.Metadata)

	return info, nil
}

// List returns files matching the prefix in Azure Blob Storage.
func (st *Storage) List(ctx context.Context, prefix string, opts ...objstore.ListOption) (*objstore.ListResult, error) {
	options := objstore.ApplyListOptions(opts)

	// Use hierarchy listing when delimiter is set and not recursive
	if !options.Recursive && options.Delimiter != "" {
		return st.listHierarchy(ctx, prefix, options)
	}

	listPrefix := st.blobName(prefix)
	listOpts := &azblob.ListBlobsFlatOptions{
		Prefix:     &listPrefix,
		MaxResults: ptrInt32(int32(options.MaxKeys)),
	}

	if options.Token != "" {
		listOpts.Marker = &options.Token
	}

	result := &objstore.ListResult{
		Files:    make([]*objstore.FileInfo, 0),
		Prefixes: make([]string, 0),
	}

	pager := st.client.NewListBlobsFlatPager(st.config.ContainerName, listOpts)
	count := 0

	for pager.More() {
		if options.MaxKeys > 0 && count >= options.MaxKeys {
			result.IsTruncated = true
			break
		}

		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, item := range page.Segment.BlobItems {
			if options.MaxKeys > 0 && count >= options.MaxKeys {
				result.IsTruncated = true
				if page.NextMarker != nil {
					result.NextToken = *page.NextMarker
				}
				return result, nil
			}

			if item.Name == nil {
				continue
			}

			result.Files = append(result.Files, st.blobItemToFileInfo(item))
			count++
		}

		if page.NextMarker != nil && *page.NextMarker != "" {
			result.NextToken = *page.NextMarker
		}
	}

	return result, nil
}

// listHierarchy lists blobs using hierarchy (delimiter-based) listing.
func (st *Storage) listHierarchy(ctx context.Context, prefix string, options *objstore.ListOptions) (*objstore.ListResult, error) {
	listPrefix := st.blobName(prefix)
	listOpts := &container.ListBlobsHierarchyOptions{
		Prefix:     &listPrefix,
		MaxResults: ptrInt32(int32(options.MaxKeys)),
	}

	if options.Token != "" {
		listOpts.Marker = &options.Token
	}

	result := &objstore.ListResult{
		Files:    make([]*objstore.FileInfo, 0),
		Prefixes: make([]string, 0),
	}

	containerClient := st.client.ServiceClient().NewContainerClient(st.config.ContainerName)
	pager := containerClient.NewListBlobsHierarchyPager(options.Delimiter, listOpts)
	count := 0

	for pager.More() {
		if options.MaxKeys > 0 && count >= options.MaxKeys {
			result.IsTruncated = true
			break
		}

		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		// Process prefixes (directories)
		for _, p := range page.Segment.BlobPrefixes {
			if p.Name != nil {
				result.Prefixes = append(result.Prefixes, st.stripPrefix(*p.Name))
			}
		}

		// Process blobs
		for _, item := range page.Segment.BlobItems {
			if options.MaxKeys > 0 && count >= options.MaxKeys {
				result.IsTruncated = true
				break
			}

			if item.Name == nil {
				continue
			}

			result.Files = append(result.Files, st.blobItemToFileInfo(item))
			count++
		}

		if page.NextMarker != nil && *page.NextMarker != "" {
			result.NextToken = *page.NextMarker
		}
	}

	return result, nil
}

// Copy copies a file in Azure Blob Storage.
func (st *Storage) Copy(ctx context.Context, src, dst string) error {
	srcBlobName := st.blobName(src)
	dstBlobName := st.blobName(dst)

	srcURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s",
		st.config.AccountName, st.config.ContainerName, srcBlobName)

	dstBlobClient := st.client.ServiceClient().NewContainerClient(st.config.ContainerName).NewBlobClient(dstBlobName)
	_, err := dstBlobClient.CopyFromURL(ctx, srcURL, nil)
	if err != nil {
		if isNotFoundError(err) {
			return objstore.ErrNotFound
		}
		return err
	}

	return nil
}

// Move moves a file in Azure Blob Storage.
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

	blobName := st.blobName(path)
	return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s",
		st.config.AccountName, st.config.ContainerName, blobName), nil
}

// SignedURL returns a SAS URL for temporary access.
func (st *Storage) SignedURL(ctx context.Context, path string, opts ...objstore.SignedURLOption) (string, error) {
	options := objstore.ApplySignedURLOptions(opts)

	if st.sharedKey == nil {
		return "", fmt.Errorf("%w: signed URLs require shared key credentials", objstore.ErrNotImplemented)
	}

	blobName := st.blobName(path)

	perms := sas.BlobPermissions{Read: true}
	if options.Method == "PUT" {
		perms = sas.BlobPermissions{Write: true, Create: true}
	}

	now := time.Now().UTC()
	sasValues := sas.BlobSignatureValues{
		Protocol:      sas.ProtocolHTTPS,
		StartTime:     now,
		ExpiryTime:    now.Add(options.Expires),
		Permissions:   perms.String(),
		ContainerName: st.config.ContainerName,
		BlobName:      blobName,
	}

	if options.ContentType != "" {
		sasValues.ContentType = options.ContentType
	}

	queryParams, err := sasValues.SignWithSharedKey(st.sharedKey)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s?%s",
		st.config.AccountName, st.config.ContainerName, blobName, queryParams.Encode()), nil
}

// DeleteMultiple deletes multiple files in Azure Blob Storage.
// This implements the objstore.BatchDeleter interface.
func (st *Storage) DeleteMultiple(ctx context.Context, paths []string) error {
	for _, path := range paths {
		if err := st.Delete(ctx, path); err != nil && !errors.Is(err, objstore.ErrNotFound) {
			return err
		}
	}
	return nil
}

// blobItemToFileInfo converts an Azure blob item to FileInfo.
func (st *Storage) blobItemToFileInfo(item *container.BlobItem) *objstore.FileInfo {
	path := st.stripPrefix(*item.Name)
	info := &objstore.FileInfo{
		Path:        path,
		Name:        filepath.Base(path),
		ContentType: objstore.DetectContentType(path),
	}

	if item.Properties != nil {
		if item.Properties.ContentLength != nil {
			info.Size = *item.Properties.ContentLength
		}
		if item.Properties.ContentType != nil {
			info.ContentType = *item.Properties.ContentType
		}
		if item.Properties.ETag != nil {
			info.ETag = string(*item.Properties.ETag)
		}
		if item.Properties.LastModified != nil {
			info.LastModified = *item.Properties.LastModified
		}
	}

	return info
}

// blobName returns the full blob name with prefix.
func (st *Storage) blobName(path string) string {
	path = strings.TrimPrefix(path, "/")
	if st.config.Prefix != "" {
		return strings.TrimSuffix(st.config.Prefix, "/") + "/" + path
	}
	return path
}

// stripPrefix removes the configured prefix from a blob name.
func (st *Storage) stripPrefix(name string) string {
	if st.config.Prefix != "" {
		prefix := strings.TrimSuffix(st.config.Prefix, "/") + "/"
		return strings.TrimPrefix(name, prefix)
	}
	return name
}

// isNotFoundError checks if an error is a not found error.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	if bloberror.HasCode(err, bloberror.BlobNotFound) {
		return true
	}

	if bloberror.HasCode(err, bloberror.ContainerNotFound) {
		return true
	}

	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) && respErr.StatusCode == 404 {
		return true
	}

	return false
}

// toAzureMetadata converts map[string]string to map[string]*string for Azure SDK.
func toAzureMetadata(m map[string]string) map[string]*string {
	if m == nil {
		return nil
	}
	result := make(map[string]*string, len(m))
	for k, v := range m {
		result[k] = &v
	}
	return result
}

// fromAzureMetadata converts map[string]*string from Azure SDK to map[string]string.
func fromAzureMetadata(m map[string]*string) map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		if v != nil {
			result[k] = *v
		}
	}
	return result
}

func ptrInt32(v int32) *int32 {
	return &v
}

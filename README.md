# objstore

[![Go Reference](https://pkg.go.dev/badge/github.com/KARTIKrocks/objstore.svg)](https://pkg.go.dev/github.com/KARTIKrocks/objstore)
[![Go Report Card](https://goreportcard.com/badge/github.com/KARTIKrocks/objstore)](https://goreportcard.com/report/github.com/KARTIKrocks/objstore)
[![Go Version](https://img.shields.io/github/go-mod/go-version/KARTIKrocks/objstore)](go.mod)
[![CI](https://github.com/KARTIKrocks/objstore/actions/workflows/ci.yml/badge.svg)](https://github.com/KARTIKrocks/objstore/actions/workflows/ci.yml)
[![GitHub tag](https://img.shields.io/github/v/tag/KARTIKrocks/objstore)](https://github.com/KARTIKrocks/objstore/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![codecov](https://codecov.io/gh/KARTIKrocks/objstore/branch/main/graph/badge.svg)](https://codecov.io/gh/KARTIKrocks/objstore)

Unified file storage interface for Go, supporting local filesystem, AWS S3, Google Cloud Storage, Azure Blob Storage, and in-memory storage for testing.

## Installation

```bash
go get github.com/KARTIKrocks/objstore
```

## Quick Start

```go
import "github.com/KARTIKrocks/objstore"

// Local storage
store, _ := objstore.NewLocalStorage(
    objstore.DefaultLocalConfig().WithBasePath("./uploads"),
)

// Upload file
file, _ := os.Open("document.pdf")
info, _ := store.Put(ctx, "docs/document.pdf", file)

// Download file
reader, _ := store.Get(ctx, "docs/document.pdf")
defer reader.Close()

// Delete file
store.Delete(ctx, "docs/document.pdf")
```

## Storage Backends

### Local Storage

```go
config := objstore.LocalConfig{
    BasePath:        "./storage",
    BaseURL:         "https://example.com/files",
    CreateDirs:      true,
    FilePermissions: 0644,
    DirPermissions:  0755,
}

store, err := objstore.NewLocalStorage(config)

// Or with builder pattern
store, err := objstore.NewLocalStorage(
    objstore.DefaultLocalConfig().
        WithBasePath("/var/uploads").
        WithBaseURL("https://cdn.example.com"),
)
```

### AWS S3

```go
import "github.com/KARTIKrocks/objstore/s3"

store, err := s3.New(ctx,
    s3.DefaultConfig().
        WithBucket("my-bucket").
        WithRegion("us-west-2").
        WithCredentials("ACCESS_KEY", "SECRET_KEY"),
)

// With custom endpoint (MinIO, DigitalOcean Spaces, etc.)
store, err := s3.New(ctx,
    s3.DefaultConfig().
        WithBucket("my-bucket").
        WithEndpoint("https://nyc3.digitaloceanspaces.com").
        WithPathStyle(true),
)

// With IAM role (no credentials needed)
store, err := s3.New(ctx,
    s3.DefaultConfig().
        WithBucket("my-bucket").
        WithRegion("us-east-1"),
)

// With path prefix
store, err := s3.New(ctx,
    s3.DefaultConfig().
        WithBucket("my-bucket").
        WithPrefix("uploads/user-123"),
)
```

### Google Cloud Storage

```go
import "github.com/KARTIKrocks/objstore/gcs"

store, err := gcs.New(ctx,
    gcs.DefaultConfig().
        WithBucket("my-bucket").
        WithCredentialsFile("/path/to/service-account.json"),
)

// With credentials JSON
store, err := gcs.New(ctx,
    gcs.DefaultConfig().
        WithBucket("my-bucket").
        WithCredentialsJSON(jsonBytes),
)

// With authorized user credentials (non-default type)
store, err := gcs.New(ctx,
    gcs.DefaultConfig().
        WithBucket("my-bucket").
        WithCredentialsFile("/path/to/authorized-user.json").
        WithCredentialsType(option.AuthorizedUser),
)

// With default credentials (GCE, Cloud Run, etc.)
store, err := gcs.New(ctx,
    gcs.DefaultConfig().
        WithBucket("my-bucket"),
)

defer store.Close()
```

### Azure Blob Storage

```go
import "github.com/KARTIKrocks/objstore/azure"

// With account name and key
store, err := azure.New(ctx,
    azure.DefaultConfig().
        WithAccountName("myaccount").
        WithAccountKey("mykey").
        WithContainerName("mycontainer"),
)

// With connection string
store, err := azure.New(ctx,
    azure.DefaultConfig().
        WithConnectionString("DefaultEndpointsProtocol=https;AccountName=...").
        WithContainerName("mycontainer"),
)

// With default Azure credentials (managed identity, env vars, CLI)
store, err := azure.New(ctx,
    azure.DefaultConfig().
        WithAccountName("myaccount").
        WithContainerName("mycontainer"),
)
```

### In-Memory Storage (Testing)

```go
store := objstore.NewMemoryStorage()

// Upload
store.Put(ctx, "test.txt", strings.NewReader("hello"))

// Verify
data, _ := objstore.GetBytes(ctx, store, "test.txt")
fmt.Println(string(data)) // "hello"

// Clear all
store.Clear()
```

## Core Operations

### Upload

```go
// From io.Reader
file, _ := os.Open("photo.jpg")
info, err := store.Put(ctx, "images/photo.jpg", file)

// With options
info, err := store.Put(ctx, "images/photo.jpg", file,
    objstore.WithContentType("image/jpeg"),
    objstore.WithMetadata(map[string]string{"author": "john"}),
    objstore.WithCacheControl("max-age=31536000"),
    objstore.WithACL("public-read"),
)

// Prevent overwrite
info, err := store.Put(ctx, "images/photo.jpg", file,
    objstore.WithOverwrite(false),
)
if err == objstore.ErrAlreadyExists {
    // File already exists
}

// Helper functions
objstore.PutBytes(ctx, store, "data.bin", []byte{1, 2, 3})
objstore.PutString(ctx, store, "hello.txt", "Hello, World!")
objstore.PutDataURI(ctx, store, "image.png", "data:image/png;base64,...")
```

### Download

```go
// As io.ReadCloser
reader, err := store.Get(ctx, "docs/file.pdf")
if err == objstore.ErrNotFound {
    // File doesn't exist
}
defer reader.Close()
io.Copy(dst, reader)

// Helper functions
data, _ := objstore.GetBytes(ctx, store, "data.bin")
text, _ := objstore.GetString(ctx, store, "hello.txt")
```

### Delete

```go
err := store.Delete(ctx, "images/photo.jpg")

// Delete with prefix (all files in directory)
objstore.DeletePrefix(ctx, store, "images/user-123/")

// S3: Batch delete multiple files
s3Store.DeleteMultiple(ctx, []string{"file1.txt", "file2.txt"})
```

### Check Existence

```go
exists, err := store.Exists(ctx, "images/photo.jpg")
```

### File Information

```go
info, err := store.Stat(ctx, "images/photo.jpg")

fmt.Println(info.Path)         // "images/photo.jpg"
fmt.Println(info.Name)         // "photo.jpg"
fmt.Println(info.Size)         // 12345
fmt.Println(info.ContentType)  // "image/jpeg"
fmt.Println(info.LastModified) // 2026-03-16 10:30:00
fmt.Println(info.ETag)         // "abc123"
fmt.Println(info.Metadata)     // map[author:john]
```

### List Files

```go
// List files in directory
result, err := store.List(ctx, "images/")

for _, file := range result.Files {
    fmt.Println(file.Path, file.Size)
}

// List subdirectories
for _, prefix := range result.Prefixes {
    fmt.Println("Directory:", prefix)
}

// With options
result, err := store.List(ctx, "images/",
    objstore.WithMaxKeys(100),
    objstore.WithDelimiter("/"),
    objstore.WithRecursive(true),
)

// Pagination
for {
    result, _ := store.List(ctx, "images/",
        objstore.WithMaxKeys(100),
        objstore.WithToken(nextToken),
    )

    // Process files...

    if !result.IsTruncated {
        break
    }
    nextToken = result.NextToken
}
```

### Copy and Move

```go
// Copy
err := store.Copy(ctx, "images/original.jpg", "images/backup.jpg")

// Move (rename)
err := store.Move(ctx, "temp/upload.jpg", "images/photo.jpg")

// Copy between storages
objstore.CopyTo(ctx, srcStore, "file.txt", dstStore, "file.txt")

// Move between storages
objstore.MoveTo(ctx, srcStore, "file.txt", dstStore, "file.txt")
```

### URLs

```go
// Public URL
url, err := store.URL(ctx, "images/photo.jpg")
// "https://cdn.example.com/images/photo.jpg"

// Signed URL (temporary access)
url, err := store.SignedURL(ctx, "images/photo.jpg",
    objstore.WithExpires(15 * time.Minute),
)

// Signed URL for upload
url, err := store.SignedURL(ctx, "uploads/new-file.jpg",
    objstore.WithMethod("PUT"),
    objstore.WithExpires(5 * time.Minute),
    objstore.WithSignedContentType("image/jpeg"),
)
```

## Helper Functions

### Path Generation

```go
// Generate unique filename
filename := objstore.GenerateFileName("photo.jpg")
// "550e8400-e29b-41d4-a716-446655440000.jpg"

// Generate date-based path
path := objstore.GeneratePath("photo.jpg", "uploads")
// "uploads/2024/01/15/550e8400-e29b-41d4-a716-446655440000.jpg"

// Generate hash-distributed path
path := objstore.GenerateHashedPath("photo.jpg", "uploads", 2)
// "uploads/55/0e/550e8400-e29b-41d4-a716-446655440000.jpg"
```

### File Type Detection

```go
info, _ := store.Stat(ctx, "file.jpg")

objstore.IsImage(info)    // true
objstore.IsVideo(info)    // false
objstore.IsAudio(info)    // false
objstore.IsDocument(info) // false
```

### Size Formatting

```go
objstore.FormatSize(1024)       // "1.0 KB"
objstore.FormatSize(1048576)    // "1.0 MB"
objstore.FormatSize(1073741824) // "1.0 GB"
```

### Sync Directory

```go
// Upload entire directory to storage
objstore.SyncDir(ctx, store, "./local/files", "remote/files")
```

## Switching Backends

All backends implement the same `objstore.Storage` interface:

```go
var store objstore.Storage

switch env {
case "development":
    store = objstore.NewMemoryStorage()
case "local":
    store, _ = objstore.NewLocalStorage(objstore.DefaultLocalConfig())
case "production-s3":
    store, _ = s3.New(ctx, s3.DefaultConfig().
        WithBucket(os.Getenv("S3_BUCKET")))
case "production-gcs":
    store, _ = gcs.New(ctx, gcs.DefaultConfig().
        WithBucket(os.Getenv("GCS_BUCKET")))
case "production-azure":
    store, _ = azure.New(ctx, azure.DefaultConfig().
        WithAccountName(os.Getenv("AZURE_ACCOUNT")).
        WithContainerName(os.Getenv("AZURE_CONTAINER")))
}

// Same API for all backends
store.Put(ctx, "file.txt", reader)
store.Get(ctx, "file.txt")
store.Delete(ctx, "file.txt")
```

## Error Handling

```go
_, err := store.Get(ctx, "missing.txt")

switch {
case errors.Is(err, objstore.ErrNotFound):
    // File doesn't exist
case errors.Is(err, objstore.ErrAlreadyExists):
    // File already exists (when overwrite=false)
case errors.Is(err, objstore.ErrInvalidPath):
    // Invalid path (e.g., path traversal attempt)
case errors.Is(err, objstore.ErrPermission):
    // Permission denied
case errors.Is(err, objstore.ErrNotImplemented):
    // Operation not supported by this backend
default:
    // Other error
}
```

## ACL Values

Common ACL values for S3 and GCS:

| ACL                         | Description                   |
| --------------------------- | ----------------------------- |
| `private`                   | Owner-only access (default)   |
| `public-read`               | Public read access            |
| `public-read-write`         | Public read/write access      |
| `authenticated-read`        | Authenticated users can read  |
| `bucket-owner-full-control` | Bucket owner has full control |

<!-- S3-compatible services are already covered by your S3 backend (MinIO, DigitalOcean Spaces, Cloudflare R2, Backblaze B2, Wasabi) via the Endpoint + PathStyle config. So you get ~6 providers for free -->

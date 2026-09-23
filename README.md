<!-- The centred logo block opens the file, so there is no h1 on line 1. -->
<!-- markdownlint-disable-next-line MD041 -->
<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="website/static/img/logo-dark.svg">
    <img src="website/static/img/logo.svg" alt="objstore" width="104" height="104">
  </picture>
</p>

<h1 align="center">objstore</h1>

<p align="center">
  Unified file storage interface for Go, supporting local filesystem, AWS S3,
  Google Cloud Storage, Azure Blob Storage, and in-memory storage for testing.
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/KARTIKrocks/objstore"><img src="https://pkg.go.dev/badge/github.com/KARTIKrocks/objstore.svg" alt="Go Reference"></a>
  <a href="go.mod"><img src="https://img.shields.io/github/go-mod/go-version/KARTIKrocks/objstore" alt="Go version"></a>
  <a href="https://github.com/KARTIKrocks/objstore/actions/workflows/ci.yml"><img src="https://github.com/KARTIKrocks/objstore/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/KARTIKrocks/objstore/releases"><img src="https://img.shields.io/github/v/tag/KARTIKrocks/objstore" alt="GitHub tag"></a>
  <a href="https://codecov.io/gh/KARTIKrocks/objstore"><img src="https://codecov.io/gh/KARTIKrocks/objstore/branch/main/graph/badge.svg" alt="codecov"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"></a>
</p>

<p align="center">
  <b><a href="https://kartikrocks.github.io/objstore/">Documentation</a></b> ·
  <b><a href="https://pkg.go.dev/github.com/KARTIKrocks/objstore">API Reference</a></b> ·
  <b><a href="CHANGELOG.md">Changelog</a></b>
</p>

## Why objstore?

Every cloud storage SDK has its own shape — different option builders,
different error types, different pagination, different ways to sign a URL.
objstore normalizes all of that behind one interface:

| Capability                                     | objstore | Raw SDK per backend |
| ---------------------------------------------- | -------- | ------------------- |
| One interface across local/S3/GCS/Azure/memory | ✓        | You build it        |
| Sentinel errors matched with `errors.Is`       | ✓        | You build it        |
| Streaming multipart upload (S3)                | ✓        | You build it        |
| Ranged (partial) downloads on every backend    | ✓        | You build it        |
| Signed URLs, including for local/memory        | ✓        | You build it        |
| Path generation, file-type detection, sync     | ✓        | You build it        |
| Zero-dependency in-memory backend for tests    | ✓        | You build it        |

objstore isn't a replacement for the cloud SDKs — it's built on top of them
(`aws-sdk-go-v2`, `cloud.google.com/go/storage`, `azure-sdk-for-go`). It's the
abstraction layer most projects end up writing themselves the first time they
need to support more than one storage backend, packaged once and kept
consistent across all five.

## Features

- **Unified Interface**: `Put`, `Get`, `Delete`, `Exists`, `Stat`, `List`, `Copy`, `Move`, `URL`, `SignedURL` — identical on every backend
- **Five Backends**: Local filesystem, AWS S3, Google Cloud Storage, Azure Blob Storage, and in-memory
- **S3-Compatible Services**: MinIO, DigitalOcean Spaces, Cloudflare R2, Backblaze B2, Wasabi via endpoint + path-style config
- **Streaming Multipart Upload**: bodies past 5MiB split into bounded-memory parts through S3's multipart uploader automatically
- **Signed URLs Everywhere**: cloud backends presign natively; local and memory produce verifiable HMAC-signed URLs
- **Ranged Downloads**: resume downloads or seek into large files with byte-range requests
- **Conditional Writes**: atomic create-only writes (`WithOverwrite(false)`) and compare-and-swap updates (`WithIfMatch`), enforced natively by each cloud provider
- **Built-in Helpers**: unique/date/hash-distributed path generation, file-type detection, size formatting, directory sync
- **Sentinel Errors**: a small, consistent error set matched with `errors.Is` across every backend
- **Zero-Dependency Root Module**: the core package has no third-party dependencies; cloud SDKs live in their own submodules
- **Zero-Dependency Testing**: the in-memory backend needs no cloud credentials or network

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

## Documentation

Full guides live at **[kartikrocks.github.io/objstore](https://kartikrocks.github.io/objstore/)**:

| Guide                                                                                | Covers                                                   |
| ------------------------------------------------------------------------------------ | -------------------------------------------------------- |
| [Getting Started](https://kartikrocks.github.io/objstore/docs/getting-started)       | Install and run your first upload/download               |
| [Local Filesystem](https://kartikrocks.github.io/objstore/docs/backends/local)       | Disk-backed storage with a signable base URL             |
| [AWS S3](https://kartikrocks.github.io/objstore/docs/backends/s3)                    | Multipart uploads, S3-compatible endpoints, batch delete |
| [Google Cloud Storage](https://kartikrocks.github.io/objstore/docs/backends/gcs)     | Service account, JSON, and default credentials           |
| [Azure Blob Storage](https://kartikrocks.github.io/objstore/docs/backends/azure)     | Account key, connection string, managed identity         |
| [In-Memory](https://kartikrocks.github.io/objstore/docs/backends/memory)             | Zero-dependency backend for tests                        |
| [Core Operations](https://kartikrocks.github.io/objstore/docs/core-operations)       | Upload, download, list, copy/move, URLs, ACLs            |
| [Signed URLs](https://kartikrocks.github.io/objstore/docs/signed-urls)               | Cloud presigning vs. HMAC-signed local/memory URLs       |
| [Helper Functions](https://kartikrocks.github.io/objstore/docs/helpers)              | Path generation, file-type detection, directory sync     |
| [Switching Backends](https://kartikrocks.github.io/objstore/docs/switching-backends) | Writing storage code that's backend-agnostic             |
| [Error Handling](https://kartikrocks.github.io/objstore/docs/errors)                 | Sentinel errors and `errors.Is` matching                 |

Exact type signatures are generated from source on
[pkg.go.dev](https://pkg.go.dev/github.com/KARTIKrocks/objstore).

Runnable programs are in [`examples/`](examples/) — basic usage, helpers, and
switching backends.

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

`Put` uploads through the AWS SDK's multipart uploader, so it isn't capped at S3's 5GB single-`PutObject` limit — bodies larger than 5MiB are automatically split into parts and uploaded with bounded memory, regardless of whether the source `io.Reader` is seekable. With default settings the ceiling is 10,000 parts × 5MiB ≈ 48.8GB; for anything larger, raise the part size with `WithUploadPartSize` (S3 allows up to 5GiB per part, giving headroom well past S3's actual 5TB max object size). By default up to 5 parts (5MiB each, 25MiB total) are buffered in memory at once; tune part size and parallelism independently with `WithUploadPartSize` / `WithUploadConcurrency`:

```go
// Lower memory ceiling: one 8MiB part in flight at a time.
store, err := s3.New(ctx,
    s3.DefaultConfig().
        WithBucket("my-bucket").
        WithUploadPartSize(8*1024*1024).
        WithUploadConcurrency(1),
)
```

An `ETag` returned for an object uploaded via multipart is S3's composite multipart ETag (`"<hex>-<partcount>"`), not a plain content MD5 — this differs from small objects (<5MiB), which still get a plain-MD5 `ETag` from a direct `PutObject`. Treat `FileInfo.ETag` as an opaque version tag, not a content hash, if your objects may cross that size threshold.

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
if errors.Is(err, objstore.ErrAlreadyExists) {
    // File already exists
}

// Update only if nobody changed it since you read it (optimistic concurrency)
stat, _ := store.Stat(ctx, "config.json")
_, err = store.Put(ctx, "config.json", bytes.NewReader(updated),
    objstore.WithIfMatch(stat.ETag),
)
if errors.Is(err, objstore.ErrPreconditionFailed) {
    // Changed or deleted concurrently — re-read and retry
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
if errors.Is(err, objstore.ErrNotFound) {
    // File doesn't exist
}
defer reader.Close()
io.Copy(dst, reader)

// Helper functions
data, _ := objstore.GetBytes(ctx, store, "data.bin")
text, _ := objstore.GetString(ctx, store, "hello.txt")

// Partial (ranged) download — e.g. resuming a download or seeking into a video
reader, err := store.Get(ctx, "videos/movie.mp4", objstore.WithRange(1024, 4096)) // bytes [1024, 5120)
reader, err := store.Get(ctx, "videos/movie.mp4", objstore.WithRange(1024, 0))    // from byte 1024 to EOF
if errors.Is(err, objstore.ErrInvalidRange) {
    // offset is negative or at/beyond the end of the file
}
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
nextToken := ""
for {
    result, err := store.List(ctx, "images/",
        objstore.WithMaxKeys(100),
        objstore.WithToken(nextToken),
    )
    if err != nil {
        log.Fatal(err)
    }

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

#### Signed URLs on the local and memory backends

Cloud backends (`s3`, `gcs`, `azure`) presign natively. The `local` and `memory`
backends have no storage server of their own, so they produce **HMAC-signed
URLs** that your application serves and verifies. Configure a signing secret and
a base URL pointing at your own file-serving/upload handler:

```go
store, _ := objstore.NewLocalStorage(
    objstore.DefaultLocalConfig().
        WithBasePath("./uploads").
        WithBaseURL("https://api.example.com/files").
        WithSigningSecret(os.Getenv("OBJSTORE_SIGNING_SECRET")),
)
// MemoryStorage: NewMemoryStorage().WithBaseURL(...).WithSigningSecret(...)

uploadURL, _ := store.SignedURL(ctx, "uploads/doc.pdf",
    objstore.WithMethod("PUT"),
    objstore.WithSignedContentType("application/pdf"),
    objstore.WithExpires(15*time.Minute),
)
```

Your handler validates the URL it receives with `VerifySignedURL`:

```go
req, err := objstore.VerifySignedURL(r.URL.String(), secret)
switch {
case errors.Is(err, objstore.ErrSignatureExpired): // 403, expired
case errors.Is(err, objstore.ErrSignatureInvalid): // 403, bad signature
case err == nil:
    // req.Method / req.Path / req.ContentType are authorized; enforce them,
    // then store.Put(...) the uploaded body (PUT) or store.Get(...) (GET).
}
```

The signature covers only the URL path, so passing `r.URL.String()` works as-is
even though it carries no scheme/host. Note that `req.Path` includes any path
prefix from `BaseURL` (e.g. a `BaseURL` of `https://api.example.com/files` yields
a `req.Path` of `/files/<object>`); strip that prefix before passing the key to
`store`.

Without a signing secret, `SignedURL` returns the unsigned public URL for `GET`
and `ErrNotImplemented` for any other method (a local backend can't authorize a
write without a secret). Both require a base URL, else `ErrNotImplemented`.

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
case errors.Is(err, objstore.ErrPreconditionFailed):
    // WithIfMatch ETag no longer matches (object changed or was deleted)
case errors.Is(err, objstore.ErrInvalidPath):
    // Invalid path (e.g., path traversal attempt)
case errors.Is(err, objstore.ErrPermission):
    // Permission denied
case errors.Is(err, objstore.ErrInvalidRange):
    // Requested byte range (WithRange) is negative or beyond the file's end
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

## License

[MIT](LICENSE)

## Contributing

Contributions welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

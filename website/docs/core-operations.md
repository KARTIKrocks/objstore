---
title: Core Operations
---

# Core Operations

Every operation below works identically regardless of which backend `store`
is — the differences are all in construction, covered per-backend under
[Backends](/docs/backends/local).

## Upload

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

// Helper functions
objstore.PutBytes(ctx, store, "data.bin", []byte{1, 2, 3})
objstore.PutString(ctx, store, "hello.txt", "Hello, World!")
objstore.PutDataURI(ctx, store, "image.png", "data:image/png;base64,...")
```

## Conditional Writes

`WithOverwrite(false)` is a create-only write, and every backend checks it
atomically: of several concurrent writers, exactly one succeeds.

`WithIfMatch` makes a write depend on the object being unchanged. Pass the
`ETag` from an earlier `Put`, `Stat`, or `List`. If anyone wrote or deleted
the object since then, the write is rejected with `ErrPreconditionFailed`.
That gives you optimistic concurrency for read-modify-write updates:

```go
for {
    info, err := store.Stat(ctx, "config.json")
    if err != nil {
        return err
    }
    current, _ := objstore.GetBytes(ctx, store, "config.json")
    updated := modify(current)

    _, err = store.Put(ctx, "config.json", bytes.NewReader(updated),
        objstore.WithIfMatch(info.ETag),
    )
    if errors.Is(err, objstore.ErrPreconditionFailed) {
        continue // someone else won the race; re-read and retry
    }
    return err
}
```

If the object doesn't exist, the error also matches `ErrNotFound`.
`WithOverwrite` is ignored when `WithIfMatch` is set.

Treat `ETag` as an opaque string. Only pass it back to the same backend it
came from, because each backend produces its own format.

| Backend | How it is enforced |
| ------- | ------------------ |
| S3 | Native `If-Match` / `If-None-Match`, including multipart uploads |
| GCS | Generation preconditions; the ETag is resolved to the generation that has it |
| Azure | Native `If-Match` / `If-None-Match` access conditions |
| Memory | Under the storage lock |
| Local | `O_EXCL` for create-only, which also holds across processes. `WithIfMatch` is atomic only against other `Put`s on the same `LocalStorage`, not against other processes or `Copy`/`Move`/`Delete` |

Some S3-compatible services don't support conditional writes yet. Check your
provider before relying on them.

## Download

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

## Delete

```go
err := store.Delete(ctx, "images/photo.jpg")

// Delete with prefix (all files in directory)
objstore.DeletePrefix(ctx, store, "images/user-123/")

// S3: Batch delete multiple files
s3Store.DeleteMultiple(ctx, []string{"file1.txt", "file2.txt"})
```

## Check Existence

```go
exists, err := store.Exists(ctx, "images/photo.jpg")
```

## File Information

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

## List Files

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

## Copy and Move

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

## URLs

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

See [Signed URLs](/docs/signed-urls) for how signing differs between cloud
backends and the local/memory backends.

## ACL Values

Common ACL values for S3 and GCS:

| ACL                         | Description                   |
| --------------------------- | ----------------------------- |
| `private`                   | Owner-only access (default)   |
| `public-read`               | Public read access            |
| `public-read-write`         | Public read/write access      |
| `authenticated-read`        | Authenticated users can read  |
| `bucket-owner-full-control` | Bucket owner has full control |

import CodeBlock from '../components/CodeBlock';

export default function OperationsDocs() {
  return (
    <section id="operations" className="py-10 border-b border-border">
      <h2 className="text-2xl font-bold text-text-heading mb-2">Core Operations</h2>
      <p className="text-text-muted mb-8">
        Every backend exposes the same set of operations. The examples below use a
        generic <code className="font-mono">store</code> variable — they work the same way
        on Local, S3, GCS, Azure, and Memory backends.
      </p>

      <div id="ops-upload" className="mt-2">
        <h3 className="text-xl font-semibold text-text-heading mb-2">Upload</h3>
        <p className="text-text-muted mb-3">From an <code className="font-mono">io.Reader</code>:</p>
        <CodeBlock code={`file, _ := os.Open("photo.jpg")
info, err := store.Put(ctx, "images/photo.jpg", file)`} />

        <p className="text-text-muted mt-4 mb-3">With per-object options:</p>
        <CodeBlock code={`info, err := store.Put(ctx, "images/photo.jpg", file,
    objstore.WithContentType("image/jpeg"),
    objstore.WithMetadata(map[string]string{"author": "john"}),
    objstore.WithCacheControl("max-age=31536000"),
    objstore.WithACL("public-read"),
)`} />

        <p className="text-text-muted mt-4 mb-3">Prevent overwriting an existing object:</p>
        <CodeBlock code={`info, err := store.Put(ctx, "images/photo.jpg", file,
    objstore.WithOverwrite(false),
)
if err == objstore.ErrAlreadyExists {
    // File already exists
}`} />

        <p className="text-text-muted mt-4 mb-3">Convenience helpers for common payload types:</p>
        <CodeBlock code={`objstore.PutBytes(ctx, store, "data.bin", []byte{1, 2, 3})
objstore.PutString(ctx, store, "hello.txt", "Hello, World!")
objstore.PutDataURI(ctx, store, "image.png", "data:image/png;base64,...")`} />
      </div>

      <div id="ops-download" className="mt-10">
        <h3 className="text-xl font-semibold text-text-heading mb-2">Download</h3>
        <p className="text-text-muted mb-3">
          <code className="font-mono">Get</code> returns an <code className="font-mono">io.ReadCloser</code> —
          stream it directly to a destination writer:
        </p>
        <CodeBlock code={`reader, err := store.Get(ctx, "docs/file.pdf")
if err == objstore.ErrNotFound {
    // File doesn't exist
}
defer reader.Close()
io.Copy(dst, reader)`} />

        <p className="text-text-muted mt-4 mb-3">Or use the helpers when the file fits in memory:</p>
        <CodeBlock code={`data, _ := objstore.GetBytes(ctx, store, "data.bin")
text, _ := objstore.GetString(ctx, store, "hello.txt")`} />
      </div>

      <div id="ops-delete" className="mt-10">
        <h3 className="text-xl font-semibold text-text-heading mb-2">Delete</h3>
        <CodeBlock code={`err := store.Delete(ctx, "images/photo.jpg")

// Delete every object under a prefix (works on every backend)
objstore.DeletePrefix(ctx, store, "images/user-123/")

// S3-only: batch delete multiple keys in a single request
s3Store.DeleteMultiple(ctx, []string{"file1.txt", "file2.txt"})`} />
      </div>

      <div id="ops-exists" className="mt-10">
        <h3 className="text-xl font-semibold text-text-heading mb-2">Existence</h3>
        <CodeBlock code={`exists, err := store.Exists(ctx, "images/photo.jpg")`} />
      </div>

      <div id="ops-stat" className="mt-10">
        <h3 className="text-xl font-semibold text-text-heading mb-2">File Info</h3>
        <p className="text-text-muted mb-3">
          <code className="font-mono">Stat</code> returns metadata without downloading the body:
        </p>
        <CodeBlock code={`info, err := store.Stat(ctx, "images/photo.jpg")

fmt.Println(info.Path)         // "images/photo.jpg"
fmt.Println(info.Name)         // "photo.jpg"
fmt.Println(info.Size)         // 12345
fmt.Println(info.ContentType)  // "image/jpeg"
fmt.Println(info.LastModified) // 2026-03-16 10:30:00
fmt.Println(info.ETag)         // "abc123"
fmt.Println(info.Metadata)     // map[author:john]`} />
      </div>

      <div id="ops-list" className="mt-10">
        <h3 className="text-xl font-semibold text-text-heading mb-2">List Files</h3>
        <CodeBlock code={`result, err := store.List(ctx, "images/")

for _, file := range result.Files {
    fmt.Println(file.Path, file.Size)
}

// List subdirectories (when delimiter is "/")
for _, prefix := range result.Prefixes {
    fmt.Println("Directory:", prefix)
}`} />

        <p className="text-text-muted mt-4 mb-3">With listing options:</p>
        <CodeBlock code={`result, err := store.List(ctx, "images/",
    objstore.WithMaxKeys(100),
    objstore.WithDelimiter("/"),
    objstore.WithRecursive(true),
)`} />

        <p className="text-text-muted mt-4 mb-3">Token-based pagination:</p>
        <CodeBlock code={`var nextToken string
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
}`} />
      </div>

      <div id="ops-copy-move" className="mt-10">
        <h3 className="text-xl font-semibold text-text-heading mb-2">Copy & Move</h3>
        <CodeBlock code={`// Copy within the same store
err := store.Copy(ctx, "images/original.jpg", "images/backup.jpg")

// Move (rename) within the same store
err := store.Move(ctx, "temp/upload.jpg", "images/photo.jpg")

// Copy across backends
objstore.CopyTo(ctx, srcStore, "file.txt", dstStore, "file.txt")

// Move across backends
objstore.MoveTo(ctx, srcStore, "file.txt", dstStore, "file.txt")`} />
      </div>

      <div id="ops-urls" className="mt-10">
        <h3 className="text-xl font-semibold text-text-heading mb-2">URLs & Signed URLs</h3>
        <p className="text-text-muted mb-3">Public URL (uses <code className="font-mono">BaseURL</code> when configured):</p>
        <CodeBlock code={`url, err := store.URL(ctx, "images/photo.jpg")
// "https://cdn.example.com/images/photo.jpg"`} />

        <p className="text-text-muted mt-4 mb-3">Signed URL for temporary read access:</p>
        <CodeBlock code={`url, err := store.SignedURL(ctx, "images/photo.jpg",
    objstore.WithExpires(15 * time.Minute),
)`} />

        <p className="text-text-muted mt-4 mb-3">Signed URL for direct browser uploads:</p>
        <CodeBlock code={`url, err := store.SignedURL(ctx, "uploads/new-file.jpg",
    objstore.WithMethod("PUT"),
    objstore.WithExpires(5 * time.Minute),
    objstore.WithSignedContentType("image/jpeg"),
)`} />
      </div>
    </section>
  );
}

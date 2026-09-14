---
title: Signed URLs
---

# Signed URLs

```go
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

Cloud backends (`s3`, `gcs`, `azure`) presign natively against the provider's
own signing scheme.

## Local and memory

The `local` and `memory` backends have no storage server of their own, so
they produce **HMAC-signed URLs** that your application serves and verifies.
Configure a signing secret and a base URL pointing at your own
file-serving/upload handler:

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

The signature covers only the URL path, so passing `r.URL.String()` works
as-is even though it carries no scheme/host. Note that `req.Path` includes
any path prefix from `BaseURL` (e.g. a `BaseURL` of
`https://api.example.com/files` yields a `req.Path` of `/files/<object>`);
strip that prefix before passing the key to `store`.

Without a signing secret, `SignedURL` returns the unsigned public URL for
`GET` and `ErrNotImplemented` for any other method (a local backend can't
authorize a write without a secret). Both require a base URL, else
`ErrNotImplemented`.

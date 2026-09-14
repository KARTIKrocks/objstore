---
title: Error Handling
---

# Error Handling

objstore normalizes backend-specific errors into a small set of sentinel
errors, matched with `errors.Is`:

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
case errors.Is(err, objstore.ErrInvalidRange):
    // Requested byte range (WithRange) is negative or beyond the file's end
case errors.Is(err, objstore.ErrNotImplemented):
    // Operation not supported by this backend
default:
    // Other error
}
```

Signed-URL verification (`VerifySignedURL`, used by the [local and memory
backends](/docs/signed-urls#local-and-memory)) adds two more:

- `objstore.ErrSignatureInvalid` — the signature doesn't match
- `objstore.ErrSignatureExpired` — the URL's expiry has passed

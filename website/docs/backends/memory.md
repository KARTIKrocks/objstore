---
title: In-Memory (Testing)
---

# In-Memory Storage

`objstore.NewMemoryStorage` keeps everything in process memory — no disk, no
network, no cleanup between test runs beyond `Clear()`.

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

Like the [local backend](/docs/backends/local), memory storage has no storage
server of its own, so [`SignedURL`](/docs/signed-urls) produces HMAC-signed
URLs when configured with `WithBaseURL` and `WithSigningSecret`.

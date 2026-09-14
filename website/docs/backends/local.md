---
title: Local Filesystem
---

# Local Filesystem

`objstore.NewLocalStorage` stores files on disk under a base path, and can
serve them through a `BaseURL` you point at your own file-serving handler.

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

Local storage has no storage server of its own, so [`SignedURL`](/docs/signed-urls)
produces **HMAC-signed URLs** that your application verifies — see
[Signed URLs on the local and memory backends](/docs/signed-urls#local-and-memory).

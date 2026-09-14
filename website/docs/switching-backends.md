---
title: Switching Backends
---

# Switching Backends

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

See the runnable version in
[`examples/switching-backends`](https://github.com/KARTIKrocks/objstore/tree/main/examples/switching-backends).

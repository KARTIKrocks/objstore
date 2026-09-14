---
title: Google Cloud Storage
---

# Google Cloud Storage

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

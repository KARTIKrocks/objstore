---
title: Getting Started
---

# Getting Started

## Installation

```bash
go get github.com/KARTIKrocks/objstore
```

The root module provides the `Storage` interface, the `local` and `memory`
backends, and the shared helpers. Cloud backends live in their own modules —
`github.com/KARTIKrocks/objstore/s3`, `.../gcs`, `.../azure` — so a build that
only uses local storage doesn't pull in every cloud SDK.

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

Every backend implements the same `objstore.Storage` interface, so this
`Put` / `Get` / `Delete` shape is identical whether `store` is backed by the
local filesystem, S3, GCS, Azure, or memory. Pick your backend from
[Backends](/docs/backends/local), then read [Core Operations](/docs/core-operations)
for everything `Storage` can do.

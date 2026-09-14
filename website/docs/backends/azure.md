---
title: Azure Blob Storage
---

# Azure Blob Storage

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

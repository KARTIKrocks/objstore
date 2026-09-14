---
title: Helper Functions
---

# Helper Functions

## Path Generation

```go
// Generate unique filename
filename := objstore.GenerateFileName("photo.jpg")
// "550e8400-e29b-41d4-a716-446655440000.jpg"

// Generate date-based path
path := objstore.GeneratePath("photo.jpg", "uploads")
// "uploads/2024/01/15/550e8400-e29b-41d4-a716-446655440000.jpg"

// Generate hash-distributed path
path := objstore.GenerateHashedPath("photo.jpg", "uploads", 2)
// "uploads/55/0e/550e8400-e29b-41d4-a716-446655440000.jpg"
```

## File Type Detection

```go
info, _ := store.Stat(ctx, "file.jpg")

objstore.IsImage(info)    // true
objstore.IsVideo(info)    // false
objstore.IsAudio(info)    // false
objstore.IsDocument(info) // false
```

## Size Formatting

```go
objstore.FormatSize(1024)       // "1.0 KB"
objstore.FormatSize(1048576)    // "1.0 MB"
objstore.FormatSize(1073741824) // "1.0 GB"
```

## Sync Directory

```go
// Upload entire directory to storage
objstore.SyncDir(ctx, store, "./local/files", "remote/files")
```

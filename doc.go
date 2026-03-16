// Package objstore provides a unified interface for file storage operations
// across multiple cloud and local backends.
//
// Supported backends:
//   - Local filesystem (objstore.LocalStorage)
//   - AWS S3 and S3-compatible services (s3.Storage)
//   - Google Cloud Storage (gcs.Storage)
//   - Azure Blob Storage (azure.Storage)
//   - In-memory for testing (objstore.MemoryStorage)
//
// All backends implement the [Storage] interface, allowing seamless switching
// between providers without changing application code.
//
// # Quick Start
//
//	// Create a storage backend
//	store, err := objstore.NewLocalStorage(
//	    objstore.DefaultLocalConfig().WithBasePath("./uploads"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Upload
//	info, err := store.Put(ctx, "images/photo.jpg", file,
//	    objstore.WithContentType("image/jpeg"),
//	)
//
//	// Download
//	reader, err := store.Get(ctx, "images/photo.jpg")
//	defer reader.Close()
//
//	// Delete
//	err = store.Delete(ctx, "images/photo.jpg")
//
// # Backend Selection
//
// Use the same interface across environments:
//
//	var store objstore.Storage
//	switch env {
//	case "test":
//	    store = objstore.NewMemoryStorage()
//	case "local":
//	    store, _ = objstore.NewLocalStorage(objstore.DefaultLocalConfig())
//	case "production":
//	    store, _ = s3.New(ctx, s3.DefaultConfig().WithBucket("my-bucket"))
//	}
//
// # Upload Options
//
// Use functional options to configure uploads:
//
//	store.Put(ctx, path, reader,
//	    objstore.WithContentType("image/png"),
//	    objstore.WithMetadata(map[string]string{"owner": "alice"}),
//	    objstore.WithCacheControl("max-age=86400"),
//	    objstore.WithACL("public-read"),
//	    objstore.WithOverwrite(false),
//	)
package objstore

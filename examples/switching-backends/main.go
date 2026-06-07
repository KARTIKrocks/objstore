// Example: switching-backends demonstrates how to swap storage backends
// using the unified objstore.Storage interface.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/KARTIKrocks/objstore"
)

// uploadAndList shows that the same code works regardless of the backend.
func uploadAndList(ctx context.Context, store objstore.Storage, name string) {
	// Upload
	info, err := objstore.PutString(ctx, store, "greeting.txt", "Hello from "+name+"!")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("[%s] Uploaded: %s (%s)\n", name, info.Path, objstore.FormatSize(info.Size))

	// Read back
	text, err := objstore.GetString(ctx, store, "greeting.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("[%s] Content: %s\n", name, text)

	// Stat
	stat, err := store.Stat(ctx, "greeting.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("[%s] Type: %s, Size: %s\n", name, stat.ContentType, objstore.FormatSize(stat.Size))

	// Cleanup
	_ = store.Delete(ctx, "greeting.txt")
	fmt.Printf("[%s] Cleaned up\n\n", name)
}

func main() {
	ctx := context.Background()

	// This function creates the right storage backend based on the environment.
	// In a real app, you'd configure this once at startup.
	newStore := func() (objstore.Storage, string) {
		switch os.Getenv("STORAGE_BACKEND") {
		case "local":
			store, err := objstore.NewLocalStorage(
				objstore.DefaultLocalConfig().WithBasePath("./data"),
			)
			if err != nil {
				log.Fatal(err)
			}
			return store, "local"

		// Uncomment to use cloud backends:
		//
		// case "s3":
		//     store, err := s3.New(ctx, s3.DefaultConfig().
		//         WithBucket(os.Getenv("S3_BUCKET")).
		//         WithRegion(os.Getenv("AWS_REGION")),
		//     )
		//     return store, "s3"
		//
		// case "gcs":
		//     store, err := gcs.New(ctx, gcs.DefaultConfig().
		//         WithBucket(os.Getenv("GCS_BUCKET")),
		//     )
		//     return store, "gcs"
		//
		// case "azure":
		//     store, err := azure.New(ctx, azure.DefaultConfig().
		//         WithAccountName(os.Getenv("AZURE_ACCOUNT")).
		//         WithContainerName(os.Getenv("AZURE_CONTAINER")),
		//     )
		//     return store, "azure"

		default:
			// In-memory for testing/demo (no setup needed)
			return objstore.NewMemoryStorage(), "memory"
		}
	}

	store, name := newStore()
	defer store.Close()

	fmt.Printf("Using backend: %s\n\n", name)

	// Same function works with any backend
	uploadAndList(ctx, store, name)

	// Demonstrate the BatchDeleter optional interface
	if bd, ok := store.(objstore.BatchDeleter); ok {
		fmt.Printf("[%s] Supports batch deletion!\n", name)
		_ = bd.DeleteMultiple(ctx, []string{"a.txt", "b.txt"})
	} else {
		fmt.Printf("[%s] No batch deletion — DeletePrefix works as a fallback\n", name)
	}

	// Cleanup local storage dir if used
	if name == "local" {
		_ = os.RemoveAll("./data")
	}
}

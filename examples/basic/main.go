// Example: basic demonstrates core objstore operations using local filesystem storage.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/KARTIKrocks/objstore"
)

func main() {
	ctx := context.Background()

	// Create a local storage backend
	store, err := objstore.NewLocalStorage(
		objstore.DefaultLocalConfig().
			WithBasePath("./example-data").
			WithBaseURL("http://localhost:8080/files"),
	)
	if err != nil {
		log.Fatal(err)
	}

	uploadAndRead(ctx, store)
	copyMoveAndList(ctx, store)
	urlsAndTypes(ctx, store)
	deleteAndOverwrite(ctx, store)

	// Cleanup
	_ = os.RemoveAll("./example-data")
	fmt.Println("\nDone! Example data cleaned up.")

	// Error handling
	_, err = store.Get(ctx, "nonexistent.txt")
	switch err { //nolint:errorlint // example code uses direct comparison for clarity
	case objstore.ErrNotFound:
		fmt.Println("Correctly got ErrNotFound for missing file")
	default:
		fmt.Println("Unexpected error:", err)
	}
}

func uploadAndRead(ctx context.Context, store *objstore.LocalStorage) {
	// Upload a string
	info, err := objstore.PutString(ctx, store, "hello.txt", "Hello, objstore!")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Uploaded: %s (%s, %s)\n", info.Path, info.ContentType, objstore.FormatSize(info.Size))

	// Upload bytes with options
	info, err = objstore.PutBytes(ctx, store, "images/logo.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg"/>`),
		objstore.WithContentType("image/svg+xml"),
		objstore.WithMetadata(map[string]string{"author": "example"}),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Uploaded: %s (%s)\n", info.Path, info.ContentType)

	// Check existence
	exists, err := store.Exists(ctx, "hello.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("hello.txt exists: %v\n", exists)

	// Get file info
	stat, err := store.Stat(ctx, "hello.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Stat: name=%s size=%s type=%s\n", stat.Name, objstore.FormatSize(stat.Size), stat.ContentType)

	// Read content back
	text, err := objstore.GetString(ctx, store, "hello.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Content: %s\n", text)
}

func copyMoveAndList(ctx context.Context, store *objstore.LocalStorage) {
	// Copy a file
	if err := store.Copy(ctx, "hello.txt", "hello-copy.txt"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Copied hello.txt -> hello-copy.txt")

	// Move a file
	if err := store.Move(ctx, "hello-copy.txt", "archive/hello.txt"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Moved hello-copy.txt -> archive/hello.txt")

	// List files
	result, err := store.List(ctx, "", objstore.WithRecursive(true))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nAll files:")
	for _, f := range result.Files {
		fmt.Printf("  %s (%s)\n", f.Path, objstore.FormatSize(f.Size))
	}
}

func urlsAndTypes(ctx context.Context, store *objstore.LocalStorage) {
	// Get public URL
	url, err := store.URL(ctx, "hello.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nPublic URL: %s\n", url)

	// File type checks
	svgInfo, _ := store.Stat(ctx, "images/logo.svg")
	fmt.Printf("\nlogo.svg is image: %v\n", objstore.IsImage(svgInfo))
	fmt.Printf("logo.svg is document: %v\n", objstore.IsDocument(svgInfo))

	// Generate unique paths
	fmt.Printf("\nGenerated filename: %s\n", objstore.GenerateFileName("photo.jpg"))
	fmt.Printf("Generated path: %s\n", objstore.GeneratePath("photo.jpg", "uploads"))
	fmt.Printf("Generated hashed path: %s\n", objstore.GenerateHashedPath("photo.jpg", "uploads", 2))
}

func deleteAndOverwrite(ctx context.Context, store *objstore.LocalStorage) {
	// Delete files
	if err := store.Delete(ctx, "hello.txt"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nDeleted hello.txt")

	// Delete all remaining files using prefix
	if err := objstore.DeletePrefix(ctx, store, ""); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deleted all remaining files")

	// Prevent overwrite
	_, _ = objstore.PutString(ctx, store, "unique.txt", "first")
	_, err := objstore.PutString(ctx, store, "unique.txt", "second", objstore.WithOverwrite(false))
	if err == objstore.ErrAlreadyExists {
		fmt.Println("Correctly prevented overwrite!")
	}
}

// Example: helpers demonstrates objstore helper functions using in-memory storage.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/KARTIKrocks/objstore"
)

func main() {
	ctx := context.Background()

	// In-memory storage is perfect for testing and examples
	store := objstore.NewMemoryStorage()

	// --- Data URI support ---
	dataURI := "data:text/plain;base64,SGVsbG8gZnJvbSBhIGRhdGEgVVJJIQ=="
	info, err := objstore.PutDataURI(ctx, store, "from-data-uri.txt", dataURI)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("From data URI: %s (%s, %s)\n", info.Path, info.ContentType, objstore.FormatSize(info.Size))

	text, _ := objstore.GetString(ctx, store, "from-data-uri.txt")
	fmt.Printf("Content: %s\n\n", text)

	// --- Parse data URIs directly ---
	data, mimeType, err := objstore.ParseDataURI("data:image/png;base64,iVBORw0KGgo=")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Parsed data URI: mime=%s, bytes=%d\n\n", mimeType, len(data))

	// --- Cross-storage operations ---
	src := objstore.NewMemoryStorage()
	dst := objstore.NewMemoryStorage()

	_, _ = objstore.PutString(ctx, src, "config.json", `{"key": "value"}`,
		objstore.WithContentType("application/json"),
	)

	// Copy between storages
	if err := objstore.CopyTo(ctx, src, "config.json", dst, "backup/config.json"); err != nil {
		log.Fatal(err)
	}
	copied, _ := objstore.GetString(ctx, dst, "backup/config.json")
	fmt.Printf("Copied between stores: %s\n", copied)

	// Move between storages
	_, _ = objstore.PutString(ctx, src, "temp.txt", "temporary data")
	if err := objstore.MoveTo(ctx, src, "temp.txt", dst, "permanent.txt"); err != nil {
		log.Fatal(err)
	}
	exists, _ := src.Exists(ctx, "temp.txt")
	fmt.Printf("Source after move exists: %v\n", exists)
	moved, _ := objstore.GetString(ctx, dst, "permanent.txt")
	fmt.Printf("Moved content: %s\n\n", moved)

	// --- File type detection ---
	files := map[string]string{
		"photo.jpg":   "image data",
		"video.mp4":   "video data",
		"song.mp3":    "audio data",
		"report.pdf":  "pdf data",
		"data.bin":    "binary data",
		"styles.css":  "css data",
		"archive.tar": "tar data",
	}

	for name, content := range files {
		_, _ = objstore.PutString(ctx, store, name, content)
	}

	fmt.Println("File type detection:")
	for name := range files {
		finfo, _ := store.Stat(ctx, name)
		fmt.Printf("  %-15s type=%-30s image=%v video=%v audio=%v doc=%v\n",
			name, finfo.ContentType,
			objstore.IsImage(finfo), objstore.IsVideo(finfo),
			objstore.IsAudio(finfo), objstore.IsDocument(finfo))
	}

	// --- Size formatting ---
	fmt.Printf("\nSize formatting:\n")
	sizes := []int64{0, 512, 1024, 1536, 1048576, 1073741824, 1099511627776}
	for _, s := range sizes {
		fmt.Printf("  %15d bytes = %s\n", s, objstore.FormatSize(s))
	}

	// --- Path generation ---
	fmt.Printf("\nPath generation:\n")
	fmt.Printf("  GenerateFileName:    %s\n", objstore.GenerateFileName("report.pdf"))
	fmt.Printf("  GeneratePath:        %s\n", objstore.GeneratePath("report.pdf", "documents"))
	fmt.Printf("  GenerateHashedPath:  %s\n", objstore.GenerateHashedPath("report.pdf", "documents", 2))
	fmt.Printf("  GenerateHashedPath:  %s\n", objstore.GenerateHashedPath("report.pdf", "documents", 3))

	// --- Memory storage testing helpers ---
	fmt.Printf("\nMemory storage stats:\n")
	fmt.Printf("  Files: %d\n", store.Size())
	fmt.Printf("  Total bytes: %s\n", objstore.FormatSize(store.TotalBytes()))

	// Direct byte access (useful in tests)
	raw, _ := store.GetBytes("photo.jpg")
	fmt.Printf("  Direct bytes for photo.jpg: %q\n", string(raw))

	// Clear all
	store.Clear()
	fmt.Printf("  After clear: %d files\n", store.Size())
}

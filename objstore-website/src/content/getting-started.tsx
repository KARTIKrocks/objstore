import CodeBlock from '../components/CodeBlock';

export default function GettingStarted() {
  return (
    <section id="getting-started" className="py-10 border-b border-border">
      <h2 className="text-2xl font-bold text-text-heading mb-4">Getting Started</h2>

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Installation</h3>
      <p className="text-text-muted mb-3">
        objstore requires <strong>Go 1.22+</strong>.
      </p>
      <CodeBlock lang="bash" code="go get github.com/KARTIKrocks/objstore" />

      <h3 className="text-lg font-semibold text-text-heading mt-8 mb-2">Quick Start</h3>
      <p className="text-text-muted mb-3">
        Upload, download, and delete a file using the local filesystem backend:
      </p>
      <CodeBlock code={`package main

import (
    "context"
    "log"
    "os"

    "github.com/KARTIKrocks/objstore"
)

func main() {
    ctx := context.Background()

    // Local storage
    store, err := objstore.NewLocalStorage(
        objstore.DefaultLocalConfig().WithBasePath("./uploads"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Upload file
    file, _ := os.Open("document.pdf")
    defer file.Close()
    info, _ := store.Put(ctx, "docs/document.pdf", file)
    log.Printf("Uploaded %s (%d bytes)", info.Path, info.Size)

    // Download file
    reader, _ := store.Get(ctx, "docs/document.pdf")
    defer reader.Close()

    // Delete file
    store.Delete(ctx, "docs/document.pdf")
}`} />
    </section>
  );
}

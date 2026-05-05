import CodeBlock from '../components/CodeBlock';

export default function MemoryDocs() {
  return (
    <section id="backends-memory" className="py-10 border-b border-border">
      <h2 className="text-2xl font-bold text-text-heading mb-2">In-Memory (Testing)</h2>
      <p className="text-text-muted mb-4">
        A drop-in implementation of <code className="font-mono">objstore.Storage</code> that keeps
        everything in process memory. Same API as the cloud backends — perfect for unit tests
        without spinning up MinIO or hitting a real bucket.
      </p>

      <CodeBlock code={`store := objstore.NewMemoryStorage()

// Upload
store.Put(ctx, "test.txt", strings.NewReader("hello"))

// Verify
data, _ := objstore.GetBytes(ctx, store, "test.txt")
fmt.Println(string(data)) // "hello"

// Clear all
store.Clear()`} />
    </section>
  );
}

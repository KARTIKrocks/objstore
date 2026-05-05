import CodeBlock from '../components/CodeBlock';

export default function HelpersDocs() {
  return (
    <section id="helpers" className="py-10 border-b border-border">
      <h2 className="text-2xl font-bold text-text-heading mb-2">Helpers</h2>
      <p className="text-text-muted mb-8">
        Utilities that complement the core <code className="font-mono">Storage</code> interface.
        All live in the root <code className="font-mono">objstore</code> package.
      </p>

      <div id="helpers-paths">
        <h3 className="text-xl font-semibold text-text-heading mb-2">Path Generation</h3>
        <p className="text-text-muted mb-3">
          Build collision-resistant object keys without rolling your own UUID/date/hash logic:
        </p>
        <CodeBlock code={`// Unique filename, original extension preserved
filename := objstore.GenerateFileName("photo.jpg")
// "550e8400-e29b-41d4-a716-446655440000.jpg"

// Date-bucketed path
path := objstore.GeneratePath("photo.jpg", "uploads")
// "uploads/2024/01/15/550e8400-e29b-41d4-a716-446655440000.jpg"

// Hash-distributed path (better object distribution on S3)
path := objstore.GenerateHashedPath("photo.jpg", "uploads", 2)
// "uploads/55/0e/550e8400-e29b-41d4-a716-446655440000.jpg"`} />
      </div>

      <div id="helpers-types" className="mt-10">
        <h3 className="text-xl font-semibold text-text-heading mb-2">File Type Detection</h3>
        <p className="text-text-muted mb-3">
          Group files by media category using the <code className="font-mono">ObjectInfo</code> from
          <code className="font-mono"> Stat</code>:
        </p>
        <CodeBlock code={`info, _ := store.Stat(ctx, "file.jpg")

objstore.IsImage(info)    // true
objstore.IsVideo(info)    // false
objstore.IsAudio(info)    // false
objstore.IsDocument(info) // false`} />
      </div>

      <div id="helpers-size" className="mt-10">
        <h3 className="text-xl font-semibold text-text-heading mb-2">Size Formatting</h3>
        <CodeBlock code={`objstore.FormatSize(1024)       // "1.0 KB"
objstore.FormatSize(1048576)    // "1.0 MB"
objstore.FormatSize(1073741824) // "1.0 GB"`} />
      </div>

      <div id="helpers-sync" className="mt-10">
        <h3 className="text-xl font-semibold text-text-heading mb-2">Sync Directory</h3>
        <p className="text-text-muted mb-3">
          Recursively upload an entire local directory tree to a remote prefix:
        </p>
        <CodeBlock code={`objstore.SyncDir(ctx, store, "./local/files", "remote/files")`} />
      </div>
    </section>
  );
}

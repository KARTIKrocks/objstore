import CodeBlock from '../components/CodeBlock';

export default function ErrorsDocs() {
  return (
    <section id="errors" className="py-10 border-b border-border">
      <h2 className="text-2xl font-bold text-text-heading mb-2">Error Handling</h2>
      <p className="text-text-muted mb-4">
        Backend-specific errors are normalized into a small set of sentinel values you can match
        against with <code className="font-mono">errors.Is</code>:
      </p>

      <CodeBlock code={`_, err := store.Get(ctx, "missing.txt")

switch {
case errors.Is(err, objstore.ErrNotFound):
    // File doesn't exist
case errors.Is(err, objstore.ErrAlreadyExists):
    // File already exists (when overwrite=false)
case errors.Is(err, objstore.ErrInvalidPath):
    // Invalid path (e.g., path traversal attempt)
case errors.Is(err, objstore.ErrPermission):
    // Permission denied
case errors.Is(err, objstore.ErrNotImplemented):
    // Operation not supported by this backend
default:
    // Other error
}`} />
    </section>
  );
}

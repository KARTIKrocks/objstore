import CodeBlock from '../components/CodeBlock';

export default function SwitchingDocs() {
  return (
    <section id="switching" className="py-10 border-b border-border">
      <h2 className="text-2xl font-bold text-text-heading mb-2">Switching Backends</h2>
      <p className="text-text-muted mb-4">
        Because every backend implements the same <code className="font-mono">objstore.Storage</code> interface,
        you can pick the one to use at startup based on environment, configuration, or feature flags —
        without changing the calling code.
      </p>

      <CodeBlock code={`var store objstore.Storage

switch env {
case "development":
    store = objstore.NewMemoryStorage()
case "local":
    store, _ = objstore.NewLocalStorage(objstore.DefaultLocalConfig())
case "production-s3":
    store, _ = s3.New(ctx, s3.DefaultConfig().
        WithBucket(os.Getenv("S3_BUCKET")))
case "production-gcs":
    store, _ = gcs.New(ctx, gcs.DefaultConfig().
        WithBucket(os.Getenv("GCS_BUCKET")))
case "production-azure":
    store, _ = azure.New(ctx, azure.DefaultConfig().
        WithAccountName(os.Getenv("AZURE_ACCOUNT")).
        WithContainerName(os.Getenv("AZURE_CONTAINER")))
}

// Same API regardless of which branch ran
store.Put(ctx, "file.txt", reader)
store.Get(ctx, "file.txt")
store.Delete(ctx, "file.txt")`} />
    </section>
  );
}

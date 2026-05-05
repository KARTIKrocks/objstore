import CodeBlock from '../components/CodeBlock';

export default function GcsDocs() {
  return (
    <section id="backends-gcs" className="py-10 border-b border-border">
      <h2 className="text-2xl font-bold text-text-heading mb-2">Google Cloud Storage</h2>
      <p className="text-text-muted mb-3">
        GCS backend lives in the <code className="text-accent font-mono">/gcs</code> subpackage.
        Supports service-account files, JSON credentials, authorized-user credentials, and
        Application Default Credentials (GCE, Cloud Run, GKE, etc.).
      </p>
      <code className="text-sm bg-bg-card px-2 py-1 rounded text-accent font-mono">
        import "github.com/KARTIKrocks/objstore/gcs"
      </code>

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Service-account file</h3>
      <CodeBlock code={`store, err := gcs.New(ctx,
    gcs.DefaultConfig().
        WithBucket("my-bucket").
        WithCredentialsFile("/path/to/service-account.json"),
)
defer store.Close()`} />

      <h3 className="text-lg font-semibold text-text-heading mt-8 mb-2">Inline JSON credentials</h3>
      <CodeBlock code={`store, err := gcs.New(ctx,
    gcs.DefaultConfig().
        WithBucket("my-bucket").
        WithCredentialsJSON(jsonBytes),
)`} />

      <h3 className="text-lg font-semibold text-text-heading mt-8 mb-2">Authorized-user credentials</h3>
      <CodeBlock code={`store, err := gcs.New(ctx,
    gcs.DefaultConfig().
        WithBucket("my-bucket").
        WithCredentialsFile("/path/to/authorized-user.json").
        WithCredentialsType(option.AuthorizedUser),
)`} />

      <h3 className="text-lg font-semibold text-text-heading mt-8 mb-2">Application Default Credentials</h3>
      <p className="text-text-muted mb-3">
        Use the ambient credentials provided by GCE, Cloud Run, GKE, or
        <code className="font-mono"> gcloud auth application-default login</code>:
      </p>
      <CodeBlock code={`store, err := gcs.New(ctx,
    gcs.DefaultConfig().
        WithBucket("my-bucket"),
)`} />
    </section>
  );
}

import CodeBlock from '../components/CodeBlock';

export default function S3Docs() {
  return (
    <section id="backends-s3" className="py-10 border-b border-border">
      <h2 className="text-2xl font-bold text-text-heading mb-2">AWS S3</h2>
      <p className="text-text-muted mb-3">
        S3 backend lives in the <code className="text-accent font-mono">/s3</code> subpackage.
        It also works with any S3-compatible service (MinIO, DigitalOcean Spaces,
        Cloudflare R2, Backblaze B2, Wasabi) via the <code className="font-mono">Endpoint</code> + <code className="font-mono">PathStyle</code> options.
      </p>
      <code className="text-sm bg-bg-card px-2 py-1 rounded text-accent font-mono">
        import "github.com/KARTIKrocks/objstore/s3"
      </code>

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">With static credentials</h3>
      <CodeBlock code={`store, err := s3.New(ctx,
    s3.DefaultConfig().
        WithBucket("my-bucket").
        WithRegion("us-west-2").
        WithCredentials("ACCESS_KEY", "SECRET_KEY"),
)`} />

      <h3 className="text-lg font-semibold text-text-heading mt-8 mb-2">S3-compatible endpoints</h3>
      <CodeBlock code={`store, err := s3.New(ctx,
    s3.DefaultConfig().
        WithBucket("my-bucket").
        WithEndpoint("https://nyc3.digitaloceanspaces.com").
        WithPathStyle(true),
)`} />

      <h3 className="text-lg font-semibold text-text-heading mt-8 mb-2">IAM role (no credentials)</h3>
      <CodeBlock code={`store, err := s3.New(ctx,
    s3.DefaultConfig().
        WithBucket("my-bucket").
        WithRegion("us-east-1"),
)`} />

      <h3 className="text-lg font-semibold text-text-heading mt-8 mb-2">Path prefix</h3>
      <p className="text-text-muted mb-3">
        Scope every operation to a sub-prefix inside the bucket:
      </p>
      <CodeBlock code={`store, err := s3.New(ctx,
    s3.DefaultConfig().
        WithBucket("my-bucket").
        WithPrefix("uploads/user-123"),
)`} />
    </section>
  );
}

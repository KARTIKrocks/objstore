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

      <h3 className="text-lg font-semibold text-text-heading mt-8 mb-2">Multipart uploads</h3>
      <p className="text-text-muted mb-3">
        <code className="font-mono">Put</code> uploads through the AWS SDK's multipart uploader, so
        it isn't capped at S3's 5GB single-<code className="font-mono">PutObject</code> limit — bodies
        larger than 5MiB are automatically split into parts and uploaded with bounded memory,
        regardless of whether the source <code className="font-mono">io.Reader</code> is seekable.
        With default settings the ceiling is 10,000 parts × 5MiB ≈ 48.8GB; for anything larger,
        raise the part size with <code className="font-mono">WithUploadPartSize</code> (S3 allows up
        to 5GiB per part, giving headroom well past S3's actual 5TB max object size). By default up
        to 5 parts (5MiB each, 25MiB total) are buffered in memory at once — tune part size and
        parallelism independently:
      </p>
      <CodeBlock code={`// Lower memory ceiling: one 8MiB part in flight at a time.
store, err := s3.New(ctx,
    s3.DefaultConfig().
        WithBucket("my-bucket").
        WithUploadPartSize(8*1024*1024).
        WithUploadConcurrency(1),
)`} />
      <p className="text-text-muted mt-4 mb-3">
        An <code className="font-mono">ETag</code> returned for an object uploaded via multipart is
        S3's composite multipart ETag (<code className="font-mono">"&lt;hex&gt;-&lt;partcount&gt;"</code>),
        not a plain content MD5 — this differs from small objects (&lt;5MiB), which still get a
        plain-MD5 <code className="font-mono">ETag</code> from a direct <code className="font-mono">PutObject</code>.
        Treat <code className="font-mono">FileInfo.ETag</code> as an opaque version tag, not a
        content hash, if your objects may cross that size threshold.
      </p>
    </section>
  );
}

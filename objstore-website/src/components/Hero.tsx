import { useState } from 'react';

interface Feature {
  title: string;
  desc: string;
}

const features: Feature[] = [
  { title: 'Unified Interface', desc: 'One Storage API works the same across every backend' },
  { title: '5 Backends', desc: 'Local filesystem, AWS S3, Google Cloud Storage, Azure Blob, In-Memory' },
  { title: 'S3-Compatible', desc: 'MinIO, DigitalOcean Spaces, Cloudflare R2, Backblaze B2, Wasabi via Endpoint + PathStyle' },
  { title: 'Signed URLs', desc: 'Generate temporary GET/PUT URLs for direct uploads and downloads' },
  { title: 'Pagination & Listing', desc: 'Token-based pagination, prefix delimiters, and recursive listing' },
  { title: 'Helpers Included', desc: 'UUID/date/hash path generation, MIME detection, size formatting, directory sync' },
  { title: 'Test-Friendly', desc: 'Drop-in in-memory backend with the same interface for fast tests' },
  { title: 'Builder Config', desc: 'Fluent WithX(...) builders on every backend’s config' },
];

const installCmd = 'go get github.com/KARTIKrocks/objstore';

export default function Hero() {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(installCmd);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <section id="top" className="py-16 border-b border-border">
      <h1 className="text-4xl md:text-5xl font-bold text-text-heading mb-4">
        Unified Go file storage interface
      </h1>
      <p className="text-lg text-text-muted max-w-2xl mb-8">
        A single Go API for local filesystem, AWS S3, Google Cloud Storage, Azure Blob,
        and in-memory storage. Switch backends with one line, get signed URLs, listing,
        copy/move, and helper utilities out of the box.
      </p>

      <div className="flex items-center gap-2 bg-bg-card border border-border rounded-lg px-4 py-3 max-w-lg mb-10">
        <span className="text-text-muted select-none">$</span>
        <code className="flex-1 text-sm font-mono text-accent">{installCmd}</code>
        <button
          onClick={handleCopy}
          className="text-xs text-text-muted hover:text-text px-2 py-1 rounded bg-overlay hover:bg-overlay-hover transition-colors"
        >
          {copied ? 'Copied!' : 'Copy'}
        </button>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {features.map((f) => (
          <div key={f.title} className="bg-bg-card border border-border rounded-lg p-4">
            <h3 className="text-sm font-semibold text-text-heading mb-1">{f.title}</h3>
            <p className="text-xs text-text-muted">{f.desc}</p>
          </div>
        ))}
      </div>
    </section>
  );
}

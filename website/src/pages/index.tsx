import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import type { ReactNode } from 'react';

import styles from './index.module.css';

type Feature = {
  readonly title: string;
  readonly description: string;
};

type Backend = {
  readonly name: string;
  readonly description: string;
};

type Operation = {
  readonly name: string;
};

const FEATURES = [
  {
    title: 'One Interface',
    description:
      'Put, Get, Delete, Exists, Stat, List, Copy, Move, URL, SignedURL — identical on every backend',
  },
  {
    title: 'Five Backends',
    description:
      'Local filesystem, AWS S3, Google Cloud Storage, Azure Blob, and in-memory',
  },
  {
    title: 'S3-Compatible Services',
    description:
      'MinIO, DigitalOcean Spaces, Cloudflare R2, Backblaze B2, Wasabi via endpoint + path-style config',
  },
  {
    title: 'Streaming Multipart Upload',
    description:
      "Bodies past 5MiB split into bounded-memory parts through S3's multipart uploader automatically",
  },
  {
    title: 'Signed URLs Everywhere',
    description:
      'Cloud backends presign natively; local and memory produce verifiable HMAC-signed URLs',
  },
  {
    title: 'Ranged Downloads',
    description:
      'Resume downloads or seek into large files with byte-range requests',
  },
  {
    title: 'Conditional Writes',
    description:
      'Atomic create-only writes and ETag compare-and-swap updates, enforced natively by each provider',
  },
  {
    title: 'Built-in Helpers',
    description:
      'Unique/date/hash-distributed path generation, file-type detection, size formatting, directory sync',
  },
  {
    title: 'Zero-Dependency Testing',
    description:
      'In-memory backend for unit tests — no cloud credentials, no network',
  },
  {
    title: 'Sentinel Errors',
    description:
      'A small, consistent error set matched with errors.Is across every backend',
  },
] as const satisfies readonly Feature[];

// Kept in sync with the "Storage Backends" section of the repository README.
const BACKENDS = [
  {
    name: 'Local Filesystem',
    description: 'Disk-backed storage with an HMAC-signable base URL',
  },
  {
    name: 'AWS S3',
    description:
      'Multipart uploads, IAM roles, path prefixes, custom endpoints',
  },
  {
    name: 'Google Cloud Storage',
    description: 'Service account, JSON, or default credentials',
  },
  {
    name: 'Azure Blob Storage',
    description: 'Account key, connection string, or managed identity',
  },
  { name: 'In-Memory', description: 'Zero-dependency backend built for tests' },
] as const satisfies readonly Backend[];

// Kept in sync with the "Core Operations" section of the repository README.
const OPERATIONS = [
  { name: 'Put' },
  { name: 'Get' },
  { name: 'Delete' },
  { name: 'Exists' },
  { name: 'Stat' },
  { name: 'List' },
  { name: 'Copy' },
  { name: 'Move' },
  { name: 'URL' },
  { name: 'SignedURL' },
] as const satisfies readonly Operation[];

const INSTALL_COMMAND = 'go get github.com/KARTIKrocks/objstore';

function Hero(): ReactNode {
  return (
    <header className={styles.hero}>
      <div className="container">
        <h1 className={styles.title}>Unified file storage interface for Go</h1>
        <p className={styles.subtitle}>
          One <code>Storage</code> interface for local filesystem, AWS S3,
          Google Cloud Storage, Azure Blob Storage, and in-memory testing. Write
          your storage layer once — swap the backend with configuration, not
          code.
        </p>

        <div className={styles.buttons}>
          <Link
            className="button button--primary button--lg"
            to="/docs/getting-started">
            Get Started
          </Link>
          <Link
            className="button button--secondary button--lg"
            to="https://pkg.go.dev/github.com/KARTIKrocks/objstore">
            API Reference
          </Link>
        </div>

        <div className={styles.install}>
          <span className={styles.prompt} aria-hidden="true">
            $
          </span>
          <code>{INSTALL_COMMAND}</code>
        </div>
      </div>
    </header>
  );
}

function Features(): ReactNode {
  return (
    <section className="container" aria-label="Features">
      <div className={styles.features}>
        {FEATURES.map((feature) => (
          <article key={feature.title} className={styles.card}>
            <h2>{feature.title}</h2>
            <p>{feature.description}</p>
          </article>
        ))}
      </div>
    </section>
  );
}

function Backends(): ReactNode {
  return (
    <section className={styles.section}>
      <div className="container">
        <h2 className={styles.sectionTitle}>One interface, five backends</h2>
        <p className={styles.sectionLead}>
          Every backend below implements the same <code>objstore.Storage</code>{' '}
          interface. Develop against memory or local disk, then point the same
          code at S3, GCS, or Azure in production — see{' '}
          <Link to="/docs/switching-backends">Switching Backends</Link>.
        </p>

        <div className={styles.backends}>
          {BACKENDS.map((backend) => (
            <article key={backend.name} className={styles.backendCard}>
              <h3>{backend.name}</h3>
              <p>{backend.description}</p>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}

function Operations(): ReactNode {
  return (
    <section className={styles.section}>
      <div className="container">
        <h2 className={styles.sectionTitle}>The same operations, everywhere</h2>
        <p className={styles.sectionLead}>
          No backend-specific escape hatches to learn — these ten operations are
          all you need, and they behave the same on every backend.
        </p>

        <div className={styles.operations}>
          {OPERATIONS.map((op) => (
            <code key={op.name} className={styles.operation}>
              {op.name}
            </code>
          ))}
        </div>
      </div>
    </section>
  );
}

export default function Home(): ReactNode {
  const { siteConfig } = useDocusaurusContext();

  return (
    <Layout
      title={siteConfig.tagline}
      description="A unified file storage interface for Go, supporting local filesystem, AWS S3, Google Cloud Storage, Azure Blob Storage, and in-memory storage for testing.">
      <Hero />
      <main>
        <Features />
        <Backends />
        <Operations />
      </main>
    </Layout>
  );
}

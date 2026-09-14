---
title: Introduction
slug: /
---

# objstore

**objstore** is a unified file storage interface for Go. It gives you one
`Storage` interface — `Put`, `Get`, `Delete`, `Exists`, `Stat`, `List`,
`Copy`, `Move`, `URL`, `SignedURL` — backed by whichever backend your
environment needs:

- **Local filesystem** — for development, or single-node deployments
- **AWS S3** — including S3-compatible services (MinIO, DigitalOcean Spaces,
  Cloudflare R2, Backblaze B2, Wasabi) via endpoint + path-style config
- **Google Cloud Storage**
- **Azure Blob Storage**
- **In-memory** — for tests, with zero external dependencies

Write your application against `objstore.Storage` once, and swap the backend
by changing configuration — not code. See [Switching Backends](/docs/switching-backends)
for the pattern.

## Why objstore?

Every cloud storage SDK has its own shape: different option builders,
different error types, different pagination, different ways to sign a URL.
objstore normalizes all of that behind one interface, so:

- Local development and CI can run against the `memory` or `local` backend
  with no cloud credentials.
- Moving from S3 to GCS (or supporting both) doesn't mean rewriting your
  storage layer.
- Common needs — generating unique paths, detecting file types, formatting
  sizes, syncing a directory — are built in as [helpers](/docs/helpers), not
  re-implemented per project.

## Next steps

- [Getting Started](/docs/getting-started) — install and run your first
  upload/download.
- [Backends](/docs/backends/local) — configuration for each storage backend.
- [Core Operations](/docs/core-operations) — the operations every backend
  supports the same way.

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`objstore` is a Go library providing a unified file-storage interface across
local filesystem, AWS S3, Google Cloud Storage, Azure Blob Storage, and an
in-memory backend for testing. Every backend implements the single
`objstore.Storage` interface so application code can switch providers without
changing call sites.

## Multi-module layout

This is a **multi-module repository**. The cloud backends are separate Go
modules so that applications only pull in the SDK dependencies they actually
use:

- **Root** `github.com/KARTIKrocks/objstore` — the `Storage` interface, shared
  types/options, sentinel errors, and the `LocalStorage` + `MemoryStorage`
  backends, plus standalone helpers. Only dependency is `github.com/google/uuid`.
- **`s3/`** `…/objstore/s3` — AWS S3 / S3-compatible backend (aws-sdk-go-v2).
- **`gcs/`** `…/objstore/gcs` — Google Cloud Storage backend.
- **`azure/`** `…/objstore/azure` — Azure Blob Storage backend.

Each sub-module's `go.mod` **requires a published version** of the root module
(e.g. `github.com/KARTIKrocks/objstore v0.1.3`), not a `replace` directive.
This means: when you change root code that a sub-module depends on, the
sub-module will **not** see your local changes until you wire up replace
directives. Run `make release-local` to add `replace …/objstore => ../` to every
sub-module for local development. Run `make release-prep VERSION=vX.Y.Z` to strip
those replace directives and pin a real version before tagging a release.

## Common commands

All test/build/lint targets loop over every module (`.`, `./s3`, `./gcs`,
`./azure`) — `go test ./...` from the root only covers the root module.

```bash
make all          # tidy, fmt, vet, lint, build, test across all modules
make ci           # what CI runs: tidy, fmt-check, vet, lint, test-race
make test         # go test ./... in each module
make test-race    # go test -race -count=1 ./... in each module
make lint         # golangci-lint run (installs golangci-lint v2.12.2 if missing)
make fix          # fmt + golangci-lint --fix
make coverage     # per-module coverage profiles
make bench        # go test -bench=. -benchmem ./...
```

### Running a single test

`make` targets don't take test filters; invoke `go test` directly in the
relevant module directory:

```bash
go test -run TestLocalStorage_Put ./...          # root module
cd s3 && go test -run TestStorage_SignedURL ./... # a sub-module
```

`make setup` installs `golangci-lint` and `goimports` if absent. Go 1.26+ is
required.

## Architecture

### The Storage interface (`objstore.go`)

`Storage` is the contract every backend implements: `Put`, `Get`, `Delete`,
`Exists`, `Stat`, `List`, `Copy`, `Move`, `URL`, `SignedURL`, `Close`. Shared
across all backends:

- **Functional options.** `Put`/`List`/`SignedURL` take variadic option funcs
  (`WithContentType`, `WithMetadata`, `WithOverwrite`, `WithMaxKeys`,
  `WithExpires`, …). Backends materialize them via the exported
  `ApplyPutOptions` / `ApplyListOptions` / `ApplySignedURLOptions` helpers,
  which also set the defaults (Put defaults to `Overwrite: true`; List defaults
  to `MaxKeys: 1000`, `Delimiter: "/"`; SignedURL defaults to 15-minute GET).
- **Sentinel errors.** Backends must map provider-specific failures to the
  shared sentinels (`ErrNotFound`, `ErrAlreadyExists`, `ErrInvalidPath`, …) so
  callers can use `errors.Is`. This is the main consistency requirement when
  adding or modifying a backend.
- **`FileInfo` / `ListResult`** are the common return shapes; backends translate
  their native metadata into these.
- **`BatchDeleter`** is an _optional_ interface (`DeleteMultiple`). Helpers
  type-assert for it and fall back to one-by-one deletion when a backend doesn't
  implement it.

### Backend-agnostic helpers (`helpers.go`)

Free functions that operate on any `Storage` (e.g. `PutBytes`, `GetString`,
`CopyTo`/`MoveTo` for cross-backend transfers, `DeletePrefix`, `SyncDir`,
`GeneratePath`/`GenerateHashedPath`, `ParseDataURI`/`PutDataURI`, content-type
predicates). These live in the root module and work uniformly because they only
touch the `Storage` interface.

### Signed URLs for local/memory (`signing.go`)

The cloud backends presign natively; `local` and `memory` have no storage
server, so `SignedURL` produces HMAC-SHA256 signed URLs (when a signing secret
is configured via `LocalConfig.WithSigningSecret` / `MemoryStorage.WithSigningSecret`)
that the application serves and verifies with the standalone
`VerifySignedURL(rawURL, secret)`, which returns a `SignedRequest`. Honors the
`WithMethod` / `WithExpires` / `WithSignedContentType` options; maps failures to
the `ErrSignatureInvalid` / `ErrSignatureExpired` sentinels. Without a secret,
`SignedURL` returns the unsigned public URL for GET and `ErrNotImplemented` for
any other method; both paths require a base URL.

### Config builders

Each backend has a `DefaultConfig()` returning a value plus chainable
`WithX()` methods that **return a copy** (value receiver, not pointer) — config
objects are immutable builders. Root local storage uses `LocalConfig` /
`DefaultLocalConfig()`; sub-modules expose their own `Config` with backend-
specific fields (bucket, region, credentials, endpoint, path-style, prefix).

### Adding behavior to a backend

When implementing a backend method, match the established patterns: apply
options through the `Apply*Options` helpers, translate native errors into the
objstore sentinels, populate `FileInfo` consistently, and honor `context`
cancellation (see `ctxReader` in `local.go` for the local-fs approach).

## Documentation website

The project's docs site lives on a **separate `website` branch** — not on
`main`. On `main` the `objstore-website/` directory is **gitignored** (it's only
a local working copy); the canonical source is `website:objstore-website/`, and
the built output deploys to the `gh-pages` branch. When asked to work on the
docs, `git checkout website` first.

It is a standalone frontend project with a completely different toolchain from
the Go library (npm/Vite, not `make`):

- **Stack:** React 19 + TypeScript, Vite, Tailwind CSS 4, and Shiki for code
  highlighting.
- **Layout (`objstore-website/`):** `src/components/` (`Navbar`, `Sidebar`,
  `Hero`, `CodeBlock`, `ThemeToggle`, `ThemeProvider`, `ModuleSection`) and
  `src/content/` — one page per backend/topic: `getting-started`, `local`,
  `memory`, `s3`, `gcs`, `azure`, `operations`, `helpers`, `errors`,
  `switching`, and `acl`. When library behavior changes, the matching
  `src/content/*.tsx` page is what needs updating.
- **Commands (run inside `objstore-website/`):** `npm run dev` (Vite dev
  server), `npm run build` (`tsc -b && vite build`), `npm run lint` (eslint),
  `npm run deploy` (`npm run build && gh-pages -d dist` → publishes to
  `gh-pages`).
- **Base path:** `vite.config.ts` sets `base: '/objstore/'` for GitHub Pages;
  keep it in sync with the repo name if it ever changes.

## Conventions

- All exported types and functions need doc comments (enforced by review/lint).
- Run `make fmt` (gofmt -s + goimports) before committing; CI fails on
  unformatted or unordered-import files via `make fmt-check`.
- Keep PRs focused on a single change and include tests for new behavior.

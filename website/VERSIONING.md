# Documentation versioning

This site is built with `docusaurus-plugin-content-docs`' versioning, but
objstore is still pre-1.0 (`v0.1.x`) — under semver, *any* `0.x` release can
change documented behaviour, so snapshotting every minor would mean
snapshotting almost every release. That defeats the point, so versioning
stays off until `v1.0.0`. Until then, `docs/` is simply "current" and serves
at `/docs/` directly — there is no `/docs/next/` split.

The scaffolding below (`versions.config.json`, `scripts/cut-version.mjs`,
this file) is in place now so turning versioning on at `v1.0.0` is a config
change, not a migration.

## The rules, from v1.0.0 onward

### 1. Snapshot when a release changes documented behaviour — not on every release

A snapshot exists to answer one question: *what was true before this release
broke it?* If nothing broke, the snapshot is a byte-identical copy that has
to be maintained forever.

**The test:** does this release make an existing page **wrong for someone on
the previous version**? A changed default, changed semantics, a rename, a
removal, a deprecation — those get a snapshot. Purely additive releases do
not — see rule 2 for what they get instead.

Go makes this rule stronger than it would be elsewhere. Semver plus the Go
compatibility promise means additive minors leave the documented contract
intact, and a `v2` is a different import path
(`github.com/KARTIKrocks/objstore/v2`) with its own pkg.go.dev — a genuinely
different package. **Always snapshot at a major.**

### 2. Additive changes get a version marker, not a snapshot

The one real problem a reader on an older version has is the inverse of a
breaking change: they read about something that does not exist in their
version yet, and it does not compile. Snapshots are an expensive fix for
that. A marker is a cheap one:

| Situation                        | Write                                                   |
| -------------------------------- | ------------------------------------------------------- |
| New row in an API table          | append `_1.6+_` to the description cell                 |
| New option or behaviour in prose | open the paragraph with `_Added in 1.6._`               |
| New member inside a code block   | trailing `// 1.6+` comment                              |
| Behaviour that changed           | `_Changed in 1.7._` plus one line on what it was before |

Markers use `MAJOR.MINOR` — `1.6`, not `v1.6.0` — so they match snapshot
names and stay greppable. Drop a marker once it names a version older than
the oldest live snapshot; by then everyone reading has it.

### 3. Snapshots are `MAJOR.MINOR`, never patch

Versions are `1.7`, `1.8`, `2.0`. Never `1.7.0`, never `v1.7`.

A patch release (`1.7.1`) that changes documented behaviour is **edited into
the existing `versioned_docs/version-1.7/` in place**. It does not get its
own snapshot. A patch by definition does not change the contract; if it did,
the docs were already wrong, and the fix belongs in the snapshot that is
wrong.

`npm run cut-version` enforces the format; it rejects anything that isn't
`MAJOR.MINOR`, already exists, or is older than the current newest.

### 4. Only the newest 4 versions are built

`MAX_LIVE_VERSIONS` in `versions.config.json` caps how many snapshots get
built and indexed. Older ones stay in git — readable at their tag,
restorable by bumping the constant — but they don't cost build time or
search index size.

### 5. `docs/` is the future, not the present

Once versioning is switched on:

| Directory                     | Serves                               | URL           |
| ----------------------------- | ------------------------------------ | ------------- |
| `docs/`                       | **Next** — unreleased, tracks `main` | `/docs/next/` |
| `versioned_docs/version-1.7/` | the current release                  | `/docs/`      |

A reader who lands on `/docs/` sees released behaviour. Someone who wants
what's on `main` opens `/docs/next/`, which carries an "unreleased" banner.

This means **a PR that changes documented behaviour edits `docs/`**, not the
snapshot. The snapshot is frozen history.

## Turning versioning on at v1.0.0

1. In `docusaurus.config.ts`, set the `docs.versions.current` override to
   `{ label: 'Next (unreleased)', path: 'next', banner: 'unreleased' }` — the
   config already does this automatically once `versions.json` has at least
   one entry, so this step may already be done for you.
2. Cut the first snapshot:

   ```bash
   cd website
   npm run cut-version -- 1.0
   npm run check
   ```

3. From then on, follow the release runbook below.

## Release runbook (post-1.0)

### Every release

Make sure `docs/` describes the release accurately, and that new APIs carry
their `_1.1+_` markers.

### If the release only adds

Nothing else to do. `docs/` becomes the new truth on the next deploy, the
markers tell readers on older versions what they need to upgrade to, and the
existing snapshot keeps serving `/docs/`.

### If the release changes documented behaviour

```bash
cd website
npm run cut-version -- 1.1
npm run check
```

`versions.json`, `versioned_docs/version-1.1/`, and
`versioned_sidebars/version-1.1-sidebars.json` are created for you, `/docs/`
starts serving 1.1, and the previous version moves into the version
dropdown.

If the cut pushes a version out of the live window, the script tells you
which one and prints the `git rm` to drop it for good.

### Patch releases

Never a snapshot. Edit `versioned_docs/version-<minor>/` directly, and
mirror the change into `docs/` if it still applies to `main`.

## Adding a page (any time, versioned or not)

1. Create `docs/<id>.md` with `title` frontmatter.
2. Add the id to `sidebars.ts` under the right category.

The sidebar is explicit rather than autogenerated so ordering is a
deliberate choice and a new file can't silently rearrange the nav.

## What does not belong here

Type signatures, method sets, and struct fields. Those are generated from
source on [pkg.go.dev](https://pkg.go.dev/github.com/KARTIKrocks/objstore)
and are always correct; hand-copying them here creates a second source of
truth that drifts.

These guides cover concepts, grouped overviews, configuration, and worked
examples — and link out for the exact signatures.

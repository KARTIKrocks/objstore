# Style and pattern rationale

Context for the scoped rules in `config.json`. This file is freeform prose read
alongside the diff; `config.json` is what actually gates comment scope and
severity.

## Offset/length overflow — worked example

The `offset-length-overflow` rule exists because of two real bugs found and
fixed in the same review pass that added ranged reads (`WithRange` / `Get`):

- `memory.go`'s `Get` computed `options.Offset+options.Length < end` to decide
  where to stop a slice. For a `Length` near `math.MaxInt64`, the sum wrapped
  negative, the comparison came out true, and `file.data[offset:end]` panicked
  with a negative slice bound instead of returning `ErrInvalidRange` or simply
  clamping to EOF.
- `s3/s3.go`'s `formatRange` computed `options.Offset+options.Length-1` to
  build the HTTP `Range` header. The same overflow produced a malformed header
  like `bytes=100--9223372036854775709` sent straight to S3, surfacing as an
  opaque 400 instead of the library's own `ErrInvalidRange`.

Both were fixed the same way: never add two caller-controlled `int64` values
and compare the sum against a bound. Instead, subtract first — compare
`length` against `size - offset` (memory.go), or check
`length - 1 <= math.MaxInt64 - offset` before forming the sum (s3.go). Apply
the same shape of fix anywhere else offset/length pairs meet arithmetic:
pagination tokens, buffer pre-sizing, hashed-path byte slicing.

## Backend parity — worked example

The same review pass caught a second issue with `Get`'s new range support: an
`offset == size` request returned an empty, successful read on `local` and
`memory`, but S3 rejects that same request with a 416/`InvalidRange` error
(per RFC 7233, a range is unsatisfiable once its first byte position is not
less than the resource length — GCS and Azure follow the same rule). A caller
who wrote code against the `Storage` interface — the library's whole reason to
exist — would see success on two backends and `ErrInvalidRange` on the other
three for an identical call. The fix was to make `local`/`memory` match the
HTTP-range-derived behavior (`offset >= size` is invalid) rather than leave
cloud backends to match the filesystem's more lenient one. When a boundary
condition differs across backends, the cloud providers' shared behavior is
usually the one to converge on, since they can't be changed to match this
library.

## Sentinel errors are the contract

`ErrNotFound`, `ErrAlreadyExists`, `ErrInvalidPath`, `ErrPermission`,
`ErrNotImplemented`, `ErrInvalidRange`, `ErrPreconditionFailed`,
`ErrSignatureInvalid`, and `ErrSignatureExpired` (`objstore.go`) are what
callers write `errors.Is` against. Some errors deliberately match two
sentinels: a `WithIfMatch` Put against a missing object wraps both
`ErrPreconditionFailed` and `ErrNotFound`, and every backend must do the same. A backend that returns a raw SDK error where a sentinel applies
silently breaks every caller's error-handling switch statement — this is a
correctness bug, not a nitpick, even though it will compile and often even
pass a happy-path test.

## Config builders are immutable

`DefaultConfig()` / `WithX()` methods use a value receiver and return a
modified copy — config objects are builders, not mutable structs. A `WithX`
written with a pointer receiver, or one that mutates a shared map/slice field
in place instead of copying it, breaks the "call chain in any order, discard
intermediates freely" contract every existing config type relies on.

## Multi-module boundaries

This is four Go modules (root, `s3/`, `gcs/`, `azure/`), each with its own
`go.mod`, tied together at dev time by the committed `go.work`. A change to
`objstore.go` (the `Storage` interface, shared option types, or a sentinel
error) needs the corresponding change in whichever of `s3/`, `gcs/`, `azure/`
implement the affected surface, in the same PR — the root module's own
`go test ./...` will not catch a mismatch, since each sub-module only pulls in
the root via its pinned `require` (overridden locally by `go.work`, not by a
`replace`).

## Conditional writes must be atomic — worked example

`WithOverwrite(false)` on `local`, `gcs`, and `azure` used to be implemented as
check-then-act: call `Exists`, and write if it returned false. Under
concurrency every writer could pass the check, so all of them "won" a
create-only write. The fix was to hand the condition to whoever performs the
write: `O_EXCL` on the local filesystem, `storage.Conditions{DoesNotExist: true}`
on GCS, `If-None-Match: *` on Azure (S3 already sent `IfNoneMatch`).
`WithIfMatch` follows the same rule: S3 and Azure send `If-Match`. GCS resolves
the ETag to a generation/metageneration and writes with `GenerationMatch`, so
the server still rejects a stale write. `local` serializes Puts per path.
`memory` checks under its write lock.

Flag any precondition (`Overwrite`, `IfMatch`, or a future one) that is
evaluated by a separate read before the write rather than by the write itself.
Also check that every write path yields a fresh ETag, since a rewrite that
leaves the ETag unchanged makes a stale `WithIfMatch` succeed. That is why
local `Put` forces mtime forward (`bumpModTime`).

## Cloud backend tests use fake HTTP servers

There are no live-provider tests. Instead each cloud backend has a small
`httptest` fake of its wire API: `fakeS3` in `s3/multipart_test.go`,
`fakeGCS` in `gcs/conditional_test.go` (JSON API via `option.WithEndpoint`),
and `fakeAzure` in `azure/conditional_test.go`
(`azblob.NewClientWithNoCredential`). A PR that changes request headers,
query parameters, or error mapping in a backend should extend the matching
fake rather than add an integration test. Pure logic (range formatting, path
building) is still best pulled into a small function with a table-driven
test, the way `formatRange` and `isPlaintextEndpoint` are.

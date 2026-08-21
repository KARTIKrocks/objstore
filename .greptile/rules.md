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
`ErrNotImplemented`, `ErrInvalidRange`, `ErrSignatureInvalid`, and
`ErrSignatureExpired` (`objstore.go`) are what callers write `errors.Is`
against. A backend that returns a raw SDK error where a sentinel applies
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

## Cloud backends have thin test coverage today

`gcs/` and `azure/` currently have no test files, and `s3/s3_test.go` only
covers pure-function helpers (`isPlaintextEndpoint`, `formatRange`) — there is
no mocked-SDK harness for `GetObject`/`Object.NewReader`/`DownloadStream` yet.
Don't expect a PR touching those backends to add integration-style tests
against a live provider; do expect any new pure logic (error mapping, header/
range formatting, path building) to be pulled into a small testable function
the way `formatRange` and `isPlaintextEndpoint` already are, with table-driven
tests alongside it.

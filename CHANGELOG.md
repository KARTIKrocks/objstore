# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.2] - 2026-06-07

### Fixed

- **S3**: `Put` with `WithOverwrite(false)` now performs an atomic conditional write (`If-None-Match: *`) instead of a separate existence check, eliminating a check-then-write race condition (requires an S3-compatible backend that supports conditional writes)
- **S3**: `DeleteMultiple` now surfaces per-object errors returned by the batch delete API instead of silently ignoring them
- **GCS**: `List` now honors pagination (`WithMaxKeys`/`WithToken`) and reports `NextToken`/`IsTruncated` based on the iterator's continuation token rather than the buffered-item count; previously these options were ignored and truncation could be misreported
- **GCS**: sentinel comparisons use `errors.Is` for `storage.ErrObjectNotExist`, and `Put` no longer drops the writer's close error
- **Azure**: `List` no longer requests zero results when no `MaxKeys` limit is set
- **Azure**: `AccountName` is now derived from the connection string so `Copy`, `URL`, and `SignedURL` work with connection-string auth, and these methods return a clear error when the account name is unavailable
- **Local**: filesystem operations (`Get`, `Delete`, `Exists`, `Stat`, `List`, `Copy`, `Move`, `DeleteDir`, and `Put` streaming) now honor `context` cancellation, including an early check in `Put` and mid-stream cancellation in `Copy`/`Move`
- **Memory**: `Move` no longer mutates a stored entry in place, fixing a data race with lock-free `List` snapshots

### Changed

- **Memory**: `Put` reads the body before acquiring the write lock and `List` operates on a snapshot, reducing lock contention
- Config initialization errors across backends are now wrapped with `%w` to preserve the error chain
- **CI**: CodeQL now builds all modules (root, `s3`, `gcs`, `azure`) so every backend is analyzed, and the coverage job runs per-module
- `make coverage` now runs per-module; added a concurrency test for the local backend

## [0.1.1] - 2026-04-01

### Changed

- Bumped AWS SDK dependencies for the S3 backend (`aws-sdk-go-v2`, `config`, `credentials`, `service/s3`)
- Bumped `google.golang.org/api` for the GCS backend
- CI now runs on Go 1.26 only
- Bumped `codecov/codecov-action` from 5 to 6 and added `codecov.yml`

### Fixed

- Fixed Go version badge link in README

## [0.1.0] - 2026-03-16

### Added

- Unified `Storage` interface with `Put`, `Get`, `Delete`, `Exists`, `Stat`, `List`, `Copy`, `Move`, `URL`, `SignedURL`, and `Close` methods
- Local filesystem backend (`objstore.LocalStorage`) with path traversal protection
- AWS S3 backend (`s3.Storage`) with support for S3-compatible services (MinIO, DigitalOcean Spaces, Cloudflare R2, Backblaze B2, Wasabi)
- Google Cloud Storage backend (`gcs.Storage`)
- Azure Blob Storage backend (`azure.Storage`) with shared key, connection string, and default credential auth
- In-memory backend (`objstore.MemoryStorage`) for testing
- `BatchDeleter` interface with implementations for S3 and Azure
- Functional options for `Put`, `List`, and `SignedURL` operations
- Helper functions: `PutBytes`, `PutString`, `GetBytes`, `GetString`, `PutDataURI`, `CopyTo`, `MoveTo`, `DeletePrefix`, `SyncDir`
- Path generation utilities: `GenerateFileName`, `GeneratePath`, `GenerateHashedPath`
- File type detection: `DetectContentType`, `IsImage`, `IsVideo`, `IsAudio`, `IsDocument`
- `FormatSize` for human-readable file sizes
- `ParseDataURI` for data URI handling
- Sentinel errors: `ErrNotFound`, `ErrAlreadyExists`, `ErrInvalidPath`, `ErrPermission`, `ErrNotImplemented`, `ErrInvalidConfig`
- CI with GitHub Actions (test, lint, coverage, CodeQL)
- Dependabot configuration for automated dependency updates

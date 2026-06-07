# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

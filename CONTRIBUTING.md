# Contributing to objstore

Thanks for your interest in contributing!

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/<username>/objstore.git`
3. Create a branch: `git checkout -b my-feature`
4. Make your changes
5. Run checks: `make all`
6. Push and open a pull request

## Development

### Prerequisites

- Go 1.26+
- golangci-lint v2

### Project Structure

This is a multi-module repository:

- **Root** (`github.com/KARTIKrocks/objstore`) — core interface, types, local/memory backends, helpers
- **`s3/`** (`github.com/KARTIKrocks/objstore/s3`) — AWS S3 backend
- **`gcs/`** (`github.com/KARTIKrocks/objstore/gcs`) — Google Cloud Storage backend
- **`azure/`** (`github.com/KARTIKrocks/objstore/azure`) — Azure Blob Storage backend

### Running Tests

```bash
make test        # run tests with race detector
make bench       # run benchmarks
make lint        # run linter
make ci          # run all checks (tidy, fmt, lint, test across all modules)
```

### Code Style

- Follow standard Go conventions
- Run `gofmt` and `goimports` before committing
- All exported types and functions must have doc comments
- Keep test coverage high for new code

## Pull Requests

- Keep PRs focused on a single change
- Include tests for new functionality
- Update documentation if the public API changes
- Ensure `make all` passes before requesting review

## Reporting Issues

- Use GitHub Issues
- Include Go version, OS, storage backend, and a minimal reproduction

## License

By contributing you agree that your contributions will be licensed under the MIT License.

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
- golangci-lint v2.12.2

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

## Module layout and `go.work`

Each backend (`s3/`, `gcs/`, `azure/`) is its own Go module, so importing
`objstore` never drags in the AWS, Google Cloud, or Azure SDKs. Each one
`require`s a published version of the root module.

The committed `go.work` at the repo root overrides that requirement with the
working tree. Without it the backends would compile against the *published*
root even here, and a breaking change to the core would pass CI while silently
breaking every backend. No `go.mod` in this repo carries a `replace`
directive — `go.work` is the only place local resolution is configured, and it
is ignored entirely when someone depends on these modules.

To reproduce a consumer's build, set `GOWORK=off`.

## Releasing

Tag the root module first, then point each backend at that tag. Everything here
is safe to commit to `main` — `go.work` keeps local builds on the working tree
no matter which root version the backends require.

```bash
git tag vX.Y.Z && git push origin vX.Y.Z    # root module first

for mod in s3 gcs azure; do
  (cd "$mod" && go mod edit -require github.com/KARTIKrocks/objstore@vX.Y.Z)
done
make tidy && make test

GOWORK=off make test    # what a consumer actually compiles

git commit -am 'Pin sub-modules to vX.Y.Z'
for mod in s3 gcs azure; do git tag "$mod/vX.Y.Z"; done
git push origin main --tags
```

The sub-module bump has to be its own commit: a module tag resolves to a commit,
and the proxy reads *that commit's* `go.mod`. Tagging before the bump would
publish a backend still requiring the old root.

## Reporting Issues

- Use GitHub Issues
- Include Go version, OS, storage backend, and a minimal reproduction

## License

By contributing you agree that your contributions will be licensed under the MIT License.

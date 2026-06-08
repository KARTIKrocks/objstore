.PHONY: all setup test test-race coverage lint lint-fix fix fmt fmt-check vet tidy build bench clean ci release-prep release-local

GOLANGCI_LINT_VERSION := v2.12.2
GOIMPORTS_VERSION := v0.45.0

MODULES = . ./s3 ./gcs ./azure
SUB_MODULES = ./s3 ./gcs ./azure
MODULE_PATH = github.com/KARTIKrocks/objstore

all: tidy fmt vet lint build test

## Install development tools (skips if already present)
setup:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	}
	@command -v goimports >/dev/null 2>&1 || { \
		echo "Installing goimports $(GOIMPORTS_VERSION)..."; \
		go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION); \
	}

## Run all checks (CI)
ci: tidy fmt-check vet lint test-race

## Build all modules
build:
	@for mod in $(MODULES); do \
		echo "==> build $$mod"; \
		(cd $$mod && go build ./...); \
	done

## Run tests across all modules
test:
	@for mod in $(MODULES); do \
		echo "==> test $$mod"; \
		(cd $$mod && go test ./...); \
	done

## Run tests with race detector across all modules
test-race:
	@for mod in $(MODULES); do \
		echo "==> test-race $$mod"; \
		(cd $$mod && go test -race -count=1 ./...); \
	done

## Run tests with coverage and generate report
coverage:
	@for mod in $(MODULES); do \
		echo "==> coverage $$mod"; \
		(cd $$mod && go test -race -coverprofile=coverage.out -covermode=atomic ./... && \
		go tool cover -func=coverage.out | tail -1); \
	done
	@echo "Per-module reports: go tool cover -html=<module>/coverage.out"

## Run linter across all modules
lint: setup
	@for mod in $(MODULES); do \
		echo "==> lint $$mod"; \
		(cd $$mod && golangci-lint run --timeout=5m ./...); \
	done

## Run golangci-lint with auto-fix
lint-fix: setup
	@for mod in $(MODULES); do \
		echo "==> Linting (fix) $$mod"; \
		(cd $$mod && golangci-lint run --fix ./...) || exit 1; \
	done

## Fix code formatting and linting issues
fix: fmt lint-fix

## Format code
fmt: setup
	@gofmt -s -w .
	@goimports -w .

## Check formatting without modifying files (used in CI)
fmt-check: setup
	@test -z "$$(gofmt -s -l . | tee /dev/stderr)" || { echo "Unformatted files found. Run 'make fmt'."; exit 1; }
	@test -z "$$(goimports -l . | tee /dev/stderr)" || { echo "Unordered imports found. Run 'make fmt'."; exit 1; }

## Run go vet across all modules
vet:
	@for mod in $(MODULES); do \
		echo "==> vet $$mod"; \
		(cd $$mod && go vet ./...); \
	done

## Run go mod tidy across all modules
tidy:
	@for mod in $(MODULES); do \
		echo "==> tidy $$mod"; \
		(cd $$mod && go mod tidy); \
	done

## Run benchmarks
bench:
	@for mod in $(MODULES); do \
		echo "==> bench $$mod"; \
		(cd $$mod && go test -bench=. -benchmem ./...); \
	done

## Remove build artifacts and coverage files
clean:
	@rm -f coverage.out
	@go clean -cache -testcache

## Prepare sub-modules for release: strip replace directives, set version
## Usage: make release-prep VERSION=v0.1.0
release-prep:
ifndef VERSION
	$(error VERSION is required. Usage: make release-prep VERSION=v0.1.0)
endif
	@for mod in $(SUB_MODULES); do \
		echo "==> release-prep $$mod"; \
		(cd $$mod && \
		go mod edit -dropreplace $(MODULE_PATH) && \
		go mod edit -require $(MODULE_PATH)@$(VERSION)); \
	done
	@echo ""
	@echo "Done! Sub-modules now point to $(MODULE_PATH)@$(VERSION)"
	@echo "Next steps:"
	@echo "  git add -A && git commit -m 'Prepare release $(VERSION)'"
	@echo "  git tag $(VERSION)"
	@echo "  git tag s3/$(VERSION)"
	@echo "  git tag gcs/$(VERSION)"
	@echo "  git tag azure/$(VERSION)"
	@echo "  git push origin main --tags"

## Restore replace directives for local development after a release
release-local:
	@for mod in $(SUB_MODULES); do \
		echo "==> release-local $$mod"; \
		(cd $$mod && \
		go mod edit -replace $(MODULE_PATH)=../ && \
		go mod tidy); \
	done
	@echo ""
	@echo "Done! Sub-modules restored to local replace directives."

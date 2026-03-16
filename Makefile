.PHONY: all test test-race coverage lint fmt vet tidy build bench clean ci release-prep release-local

MODULES = . ./s3 ./gcs ./azure
SUB_MODULES = ./s3 ./gcs ./azure
MODULE_PATH = github.com/KARTIKrocks/objstore

all: tidy fmt lint test

## Run all checks (CI)
ci: tidy fmt vet lint test-race

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
	@go test -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | tail -1
	@echo "Full report: go tool cover -html=coverage.out"

## Run linter across all modules
lint:
	@for mod in $(MODULES); do \
		echo "==> lint $$mod"; \
		(cd $$mod && golangci-lint run --timeout=5m ./...); \
	done

## Format code
fmt:
	@gofmt -s -w .
	@goimports -w .

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

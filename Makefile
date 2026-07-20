.PHONY: help tidy fmt vet lint test test-race test-integration coverage build clean

GO ?= go
GOLANGCI_LINT ?= golangci-lint

help:
	@echo "Translaas Go SDK — development targets"
	@echo ""
	@echo "  make tidy       Run go mod tidy"
	@echo "  make fmt        Format Go sources (gofmt + gofumpt if installed)"
	@echo "  make vet        Run go vet"
	@echo "  make lint       Run golangci-lint"
	@echo "  make test       Run unit tests"
	@echo "  make test-race  Run tests with -race (requires CGO)"
	@echo "  make test-integration  Run live API integration tests (requires TRANSLAAS_API_KEY)"
	@echo "  make coverage   Run tests with coverage report"
	@echo "  make build      Build all packages"
	@echo "  make clean      Remove coverage artifacts"
	@echo ""
	@echo "Runnable samples: https://github.com/acuencadev/translaas-sdk-examples (go/)"

tidy:
	$(GO) mod tidy

fmt:
	$(GO) fmt ./...
	@command -v gofumpt >/dev/null 2>&1 && gofumpt -w . || true

vet:
	$(GO) vet ./...

lint:
	$(GOLANGCI_LINT) run ./...

test:
	$(GO) test ./...

test-race:
	CGO_ENABLED=1 $(GO) test -race ./...

test-integration:
	$(GO) test -tags=integration -count=1 -timeout 5m ./tests/integration/...

coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

build:
	$(GO) build ./...

clean:
	rm -f coverage.out

SHELL    := /usr/bin/env bash
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT   := $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo "")
DATE     := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
PKG      := github.com/voska/loby
LDFLAGS  := -s -w \
            -X $(PKG)/internal/version.Version=$(VERSION) \
            -X $(PKG)/internal/version.Commit=$(COMMIT) \
            -X $(PKG)/internal/version.Date=$(DATE)
BIN      := bin/loby
TOOLS    := $(CURDIR)/.tools

export PATH := $(TOOLS):$(PATH)

.PHONY: build fmt fmt-check lint vet test test-integration ci tools clean install help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  %-18s %s\n", $$1, $$2}'

build: ## Build the loby binary into bin/
	@mkdir -p bin
	@CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/loby
	@echo "built $(BIN) ($(VERSION))"

install: ## go install loby into $$GOBIN
	@CGO_ENABLED=0 go install -trimpath -ldflags "$(LDFLAGS)" ./cmd/loby

fmt: tools ## Format Go code (gofumpt + goimports)
	@$(TOOLS)/goimports -local $(PKG) -w .
	@$(TOOLS)/gofumpt -w .

fmt-check: fmt ## Fail if formatting changes anything
	@git diff --exit-code -- '*.go' go.mod go.sum

lint: tools ## Run golangci-lint
	@$(TOOLS)/golangci-lint run ./...

vet: ## go vet ./...
	@go vet ./...

test: ## Run unit tests with race detector
	@go test -race -count=1 ./...

test-integration: ## Run integration tests (requires LOB_API_KEY)
	@go test -race -count=1 -tags=integration ./...

ci: fmt-check vet lint test build ## Full local CI gate

tools: $(TOOLS)/gofumpt $(TOOLS)/goimports $(TOOLS)/golangci-lint ## Install pinned dev tools

$(TOOLS)/gofumpt:
	@mkdir -p $(TOOLS)
	@GOBIN=$(TOOLS) go install mvdan.cc/gofumpt@latest
$(TOOLS)/goimports:
	@mkdir -p $(TOOLS)
	@GOBIN=$(TOOLS) go install golang.org/x/tools/cmd/goimports@latest
$(TOOLS)/golangci-lint:
	@mkdir -p $(TOOLS)
	@GOBIN=$(TOOLS) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

clean: ## Remove build outputs
	@rm -rf bin/ dist/

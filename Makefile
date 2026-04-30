.PHONY: help build install test lint fmt vet smoke-test clean tidy

BINARY  := mcp-withallo
PKG     := ./...

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "Targets:\n"} /^[a-zA-Z_-]+:.*##/ {printf "  %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary into ./mcp-withallo
	go build -trimpath -ldflags="-s -w" -o $(BINARY) .

install: ## go install into $GOBIN
	go install .

test: ## Run unit tests with race detector
	go test -race -count=1 $(PKG)

vet: ## Run go vet
	go vet $(PKG)

fmt: ## Format all Go files
	gofmt -w .

lint: ## Check formatting + vet (CI-equivalent)
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "needs gofmt:"; echo "$$out"; exit 1; fi
	go vet $(PKG)

tidy: ## Tidy go.mod / go.sum
	go mod tidy

smoke-test: build ## End-to-end smoke against the real API (needs ALLO_API_KEY)
	@test -n "$$ALLO_API_KEY" || (echo "ALLO_API_KEY is required" && exit 1)
	@printf '%s\n' \
	  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"smoke","version":"0"}}}' \
	  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
	  '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
	  '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"allo_me","arguments":{}}}' \
	  | ./$(BINARY)

clean: ## Remove build artifacts
	rm -f $(BINARY)
	rm -rf dist/

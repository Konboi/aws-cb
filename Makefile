GO := go
BIN := cb

.PHONY: build clean

build: ## Build the cb CLI in the repository root
	$(GO) build -o $(BIN) ./cmd/cb

clean: ## Remove build artifacts
	rm -f $(BIN)

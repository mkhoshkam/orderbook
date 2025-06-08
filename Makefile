.PHONY: fmt lint test vet

# Format Go code
fmt:
	go fmt ./...
	$(shell go env GOPATH)/bin/goimports -w .

# Run linter
lint:
	$(shell go env GOPATH)/bin/golangci-lint run

# Run tests
test:
	go test -v ./...

# Run go vet
vet:
	go vet ./...

# Check formatting
fmt-check:
	@echo "Checking Go formatting..."
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "The following files are not properly formatted:"; \
		gofmt -s -l .; \
		echo "Please run 'make fmt' to fix formatting issues"; \
		exit 1; \
	fi
	@echo "All files are properly formatted"

# Install tools
install-tools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run all checks
check: fmt-check vet lint test 
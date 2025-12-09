.PHONY: build test test-integration lint clean run

# Binary name
BINARY=jsbug
BUILD_DIR=bin

# Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/jsbug

# Run unit tests (no Chrome required)
test:
	go test ./internal/...

# Run all tests including integration (Chrome required)
test-integration:
	go test -tags chrome ./...

# Run tests with verbose output
test-verbose:
	go test -v ./internal/...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html

# Run linters
lint:
	go vet ./...
	go fmt ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Run the application
run: build
	./$(BUILD_DIR)/$(BINARY) -c config.yaml

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build for multiple platforms
build-all:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/jsbug
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/jsbug
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/jsbug
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe ./cmd/jsbug

# Show help
help:
	@echo "Available targets:"
	@echo "  build            - Build the binary"
	@echo "  test             - Run unit tests"
	@echo "  test-integration - Run all tests including integration"
	@echo "  test-verbose     - Run tests with verbose output"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  lint             - Run linters"
	@echo "  clean            - Remove build artifacts"
	@echo "  run              - Build and run the application"
	@echo "  deps             - Install dependencies"
	@echo "  build-all        - Build for multiple platforms"

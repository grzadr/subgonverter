.PHONY: all test lint build clean

# Binary name and output directory
BINARY_NAME=subgonverter
BIN_DIR=bin

# Default target
all: build

# Build the binary (depends on tests and linting passing)
build: test lint
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) .
	@echo "Binary created at $(BIN_DIR)/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BIN_DIR)
	@echo "Clean complete"

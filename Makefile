.PHONY: build clean install-scripts test

BINARY_NAME=kdm
BUILD_DIR=bin
CMD_PATH=cmd/kdm

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

install-scripts:
	@echo "Installing bash scripts..."
	@chmod +x scripts/*.sh
	@echo "Scripts are ready to use in scripts/ directory"

test:
	@echo "Running tests..."
	@go test -v ./internal/...

lint:
	@echo "Running linter..."
	@golangci-lint run ./...

run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

help:
	@echo "Available targets:"
	@echo "  build           - Build the TUI application"
	@echo "  clean           - Remove build artifacts"
	@echo "  install-scripts - Make bash scripts executable"
	@echo "  test            - Run Go tests"
	@echo "  lint            - Run golangci-lint"
	@echo "  run             - Build and run the application"
	@echo "  help            - Show this help message"

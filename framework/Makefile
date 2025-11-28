.PHONY: build clean test

build:
	@echo "Building framework..."
	@mkdir -p build
	@go build -o build/framework ./cmd/main.go
	@echo "Build complete: build/framework"

clean:
	@echo "Cleaning build directory..."
	@rm -rf build
	@echo "Clean complete"

test:
	@echo "Running tests..."
	@go test -v ./parser/... ./plugin/... ./data/... ./codegen/...

run: build
	@./build/framework

help:
	@echo "Available targets:"
	@echo "  build  - Build the framework binary"
	@echo "  clean  - Remove build artifacts"
	@echo "  test   - Run tests"
	@echo "  run    - Build and run the framework"
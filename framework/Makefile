.PHONY: build clean test

GENDIR=./generated

all: gosl_parser build

build:
	@echo "Building framework..."
	@mkdir -p build
	@go build -o build/framework ./cmd/main.go
	@echo "Build complete: build/framework"

gosl_parser:
	@echo "Building antlr step..."
	@cd ./gosl/front && if [ ! -d "${GENDIR}" ]; then \
		echo "Antlr parser not generated. Generation started"; \
		antlr4 -Dlanguage=Go -package acsl -o ${GENDIR} ./grammar/ACSL.g4; \
	else \
		echo "Antlr parser already generated. Skip generation step"; \
	fi
	@echo "Finished building antlr parser"

clean:
	@echo "Cleaning build directory..."
	@rm -rf build
	@rm -rf gosl/front/generated
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
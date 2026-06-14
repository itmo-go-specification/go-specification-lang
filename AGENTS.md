# AGENTS.md

This file provides guidance to agents when working with code in this repository.

## Build Commands

- **Generate ANTLR parser**: `cd framework && make gosl_parser` or `go generate ./...` (requires Java)
- **Build framework**: `cd framework && make build` (runs parser generation, then compiles to `build/framework`)
- **Update golden files**: `cd framework && go test ./plugin/dynamic_checks/... -update`
- **Build Docker package**: `just build-deb` (builds .deb package with cross-platform support)
- **Build Docker container**: `just docker-gpma-runtime` (builds Docker image with GPMA installed)

## Gobra Plugin

The Gobra plugin provides formal verification of Go code using the Gobra verifier (based on Viper infrastructure). It translates Go code with GPMA specifications to Gobra language and runs verification.

### Plugin Configuration

The Gobra plugin supports the following configuration parameters:

- **backend**: Execution backend - `"native"` or `"docker"` (default: `"native"`)
- **java**: Path to Java executable (default: `"java"`)
- **gobra_jar**: Path to Gobra JAR file (default: `"gobra.jar"`)
- **z3_exe**: Path to Z3 executable (default: `"z3"`)
- **docker_image**: Docker image for Gobra (default: `"ghcr.io/viperproject/gobra:sha-7fa4f86"`)
- **debug**: Enable debug mode to see generated Gobra code and analyzer output (default: `false`)

### Usage Examples

```bash
# Basic verification with Docker backend (recommended for local development)
cd framework && ./build/framework file --plugin gobra --plugin-config gobra:backend:docker ../gobra-demo/success/contracts.go

# Verification with debug mode
cd framework && ./build/framework file --plugin gobra --plugin-config gobra:debug:true ../gobra-demo/success/asserts.go

# Verification with native backend
cd framework && ./build/framework file --plugin gobra --plugin-config gobra:backend:native ../gobra-demo/success/contracts.go
```

### Environment Variables

Gobra plugin settings can be overridden via environment variables:

- `GPMA_EXECUTION_BACKEND`: Execution backend (`"native"` or `"docker"`)
- `GPMA_DOCKER_IMAGE`: Docker image for Gobra
- `GPMA_JAVA_PATH`: Path to Java executable
- `GPMA_GOBRA_JAR`: Path to Gobra JAR file
- `GPMA_GOBRA_BACKEND`: Gobra backend (`"silicon"` or `"carbon"`)
- `GPMA_TIMEOUT`: Verification timeout (default: `"120s"`)
- `GPMA_LOG_LEVEL`: Gobra log level (default: `"ERROR"`)

### Gobra Demo Examples

Working examples are located in the `gobra-demo/` directory:

- **success/**: Examples with correct specifications that pass verification
  - `contracts.go`: Function contracts (preconditions/postconditions)
  - `asserts.go`: Assertions within function bodies
  - `assert_requires.go`: Combined assertions and preconditions

- **failure/**: Examples with incorrect specifications that fail verification
  - `asserts.go`: Incorrect assertions
  - `ensures.go`: Incorrect postconditions
  - `requires.go`: Violated preconditions

### Error Reporting

The Gobra plugin provides precise error reporting with source code locations:

- Errors are mapped from generated Gobra code back to original Go source positions
- Error format: `file:line:column: error message`
- Error types include: precondition violations, postcondition violations, assertion failures, invariant violations

### Debug Mode

Enable debug mode to see:
- Generated Gobra code (written to `debug.gobra`)
- Raw Gobra analyzer output
- Detailed verification process information

### Docker Testing

- **Local testing (recommended)**: Use Docker backend for local development on macOS/Apple Silicon
  ```bash
  cd framework && ./build/framework file --plugin gobra --plugin-config gobra:backend:docker ../gobra-demo/success/contracts.go
  ```
- **Gobra demo examples**: Working examples are in `gobra-demo/` directory with success and failure test cases
- **Debug mode**: Add `--plugin-config gobra:debug:true` to see generated Gobra code and analyzer output

## Critical Non-Obvious Requirements

- **ANTLR parser generation required before building**: The framework uses ANTLR-generated parsers for ACSL/GOSL specifications. Always run `make gosl_parser` or `go generate ./...` in the `framework/` directory before building or testing. Without this, imports to `gpma/framework/gosl/front/generated/grammar/*` will fail.
- **Plugin registration via init()**: All plugins MUST register themselves in their `init()` function using `plugin.Register("name", factory)`. The CLI imports plugins as blank imports (`_ "gpma/framework/plugin/..."`) to trigger registration.
- **Test data location**: Golden test files are in `framework/plugin/dynamic_checks/testdata/golden/` with structure matching `samples/projects/`. Use `-update` flag to regenerate golden files.
- **AST modification pattern**: Plugins receive `*data.ASTData` containing Go AST with specification annotations bound via `NodeInfo map[ast.Node]*AdditionalInfo`. Modifications should update the AST directly; the code generator handles formatting.
- **Specification language**: GPMA (Go Program Modelling Annotations) is an ACSL dialect for Go. Annotations are in `/*@ ... */` comments and parsed by the ANTLR-generated parser in `framework/gosl/front/`.

## Agent Restrictions

- **Forbidden directories**: `samples/` and `test_cases/` are **NOT** accessible to agents. These directories do not exist from the agent perspective.
- **Gobra demo examples**: Working examples of the gobra plugin are located in `gobra-demo/` directory, not in `samples/`.
- **Docker backend preference**: When testing Gobra functionality, prefer Docker backend over native backend for better cross-platform compatibility.

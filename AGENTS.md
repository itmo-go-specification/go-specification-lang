# AGENTS.md

This file provides guidance to agents when working with code in this repository.

## Build Commands

- **Generate ANTLR parser**: `cd framework && make gosl_parser` or `go generate ./...` (requires Java)
- **Build framework**: `cd framework && make build` (runs parser generation, then compiles to `build/framework`)
- **Update golden files**: `cd framework && go test ./plugin/dynamic_checks/... -update`
- **Build Docker package**: `just build-deb` (builds .deb package with cross-platform support)
- **Build Docker container**: `just docker-gpma-runtime` (builds Docker image with GPMA installed)

## Docker Testing

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

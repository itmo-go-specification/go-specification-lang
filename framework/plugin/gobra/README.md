# Gobra Plugin

The Gobra plugin provides formal verification of Go code using the Gobra verifier (based on Viper infrastructure). It translates Go code with GPMA specifications to Gobra language and runs verification to ensure correctness.

## Overview

Gobra is a formal verification tool for Go programs based on the Viper infrastructure. The plugin:

1. Parses Go code with GPMA specifications (ACSL-style annotations)
2. Translates the code to Gobra language
3. Runs the Gobra verifier to check specifications
4. Maps verification errors back to original Go source locations

## Features

- **Formal verification**: Uses deductive verification to prove program correctness
- **Precise error reporting**: Maps Gobra errors back to original Go source positions
- **Hybrid mapping system**: Tracks code additions and their relationships to source
- **Multiple backends**: Supports native execution and Docker containerization
- **Debug mode**: Provides detailed output for troubleshooting

## Installation

The Gobra plugin is included in the GPMA framework. Build the framework to use it:

```bash
cd framework && make build
```

## Configuration

### Plugin Parameters

Configure the Gobra plugin using `--plugin-config` flags:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `backend` | string | `"native"` | Execution backend: `"native"` or `"docker"` |
| `java` | string | `"java"` | Path to Java executable |
| `gobra_jar` | string | `"gobra.jar"` | Path to Gobra JAR file |
| `z3_exe` | string | `"z3"` | Path to Z3 executable |
| `docker_image` | string | `"ghcr.io/viperproject/gobra:sha-7fa4f86"` | Docker image for Gobra |
| `debug` | bool | `false` | Enable debug mode |

### Environment Variables

Override settings via environment variables:

- `GPMA_EXECUTION_BACKEND`: Execution backend (`"native"` or `"docker"`)
- `GPMA_DOCKER_IMAGE`: Docker image for Gobra
- `GPMA_JAVA_PATH`: Path to Java executable
- `GPMA_GOBRA_JAR`: Path to Gobra JAR file
- `GPMA_GOBRA_BACKEND`: Gobra backend (`"silicon"` or `"carbon"`)
- `GPMA_TIMEOUT`: Verification timeout (default: `"120s"`)
- `GPMA_LOG_LEVEL`: Gobra log level (default: `"ERROR"`)

## Usage

### Basic Verification

```bash
# Verify a Go file with specifications
cd framework && ./build/framework file --plugin gobra ../path/to/file.go
```

### Docker Backend (Recommended for Local Development)

```bash
# Use Docker backend for cross-platform compatibility
cd framework && ./build/framework file --plugin gobra --plugin-config gobra:backend:docker ../path/to/file.go
```

### Debug Mode

```bash
# Enable debug mode to see generated Gobra code
cd framework && ./build/framework file --plugin gobra --plugin-config gobra:debug:true ../path/to/file.go
```

This creates a `debug.gobra` file with the generated Gobra code and shows detailed analyzer output.

## Supported Specifications

The Gobra plugin supports GPMA specifications including:

### Function Contracts

```go
/*@ 
  @ requires x >= 0;
  @ ensures res == x + 1;
  @*/
func AddOne(x int) (res int) {
    return x + 1
}
```

### Assertions

```go
func Abs(x int) int {
    /*@ assert x >= 0 ==> res == x; @*/
    /*@ assert x < 0 ==> res == -x; @*/
    if x >= 0 {
        return x
    }
    return -x
}
```

### Loop Invariants

```go
func Sum(n int) int {
    sum := 0
    i := 0
    /*@ 
      @ loop_invariant 0 <= i && i <= n;
      @ loop_invariant sum == i * (i - 1) / 2;
      @ decreases n - i;
      @*/
    for i < n {
        sum += i
        i++
    }
    return sum
}
```

## Error Reporting

The plugin provides precise error reporting with source code locations:

```
Verification failed
	Error: precondition might not hold
		x >= 0
	Error: postcondition might not hold
		res == x + 2
	Error: condition at 10:13 may fail
		x == x + 1
```

Error format: `file:line:column: error message`

Error types include:
- Precondition violations
- Postcondition violations
- Assertion failures
- Invariant violations

## Examples

Working examples are available in the `gobra-demo/` directory:

### Success Examples

- `gobra-demo/success/contracts.go`: Function contracts that pass verification
- `gobra-demo/success/asserts.go`: Correct assertions within function bodies
- `gobra-demo/success/assert_requires.go`: Combined assertions and preconditions

### Failure Examples

- `gobra-demo/failure/asserts.go`: Incorrect assertions that fail verification
- `gobra-demo/failure/ensures.go`: Incorrect postconditions
- `gobra-demo/failure/requires.go`: Violated preconditions

## Architecture

### Translation Process

1. **Parse Go code**: Parse Go source with GPMA specifications
2. **Generate Gobra code**: Translate to Gobra language with hybrid mapping
3. **Run verification**: Execute Gobra verifier
4. **Map errors**: Translate verification errors back to source positions

### Hybrid Mapping System

The plugin uses a hybrid mapping system to track code additions:

- **Interval mappings**: For displaying whole intervals of added code
- **Point mappings**: For precise predicate locations

This enables accurate error reporting by mapping Gobra error positions back to original Go source locations.

## Backend Options

### Native Backend

Runs Gobra directly on the host system:

```bash
--plugin-config gobra:backend:native
```

Requirements:
- Java runtime
- Gobra JAR file
- Z3 solver

### Docker Backend

Runs Gobra in a Docker container:

```bash
--plugin-config gobra:backend:docker
```

Benefits:
- Cross-platform compatibility
- Isolated environment
- No local dependencies

## Debugging

### Enable Debug Mode

```bash
--plugin-config gobra:debug:true
```

Debug mode provides:
- Generated Gobra code (`debug.gobra`)
- Raw Gobra analyzer output
- Detailed verification process information

### Common Issues

**Gobra not found**: Ensure Gobra JAR is installed or use Docker backend

**Timeout errors**: Increase timeout with `GPMA_TIMEOUT` environment variable

**Mapping errors**: Check debug output for generated code structure

## Testing

Run integration tests:

```bash
cd framework && go test -tags integration ./plugin/gobra/...
```

## References

- [Gobra Project](https://github.com/viperproject/gobra)
- [Viper Infrastructure](https://github.com/viperproject/viper)
- [GPMA Specification Language](../../ACSL.md)
- [Gobra Runner Documentation](../../gobra_runner/README.md)

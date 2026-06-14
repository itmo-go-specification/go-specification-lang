# Gobra Runner Refactoring Plan

## Problem Statement

The current `gobra_runner` package has several architectural issues:

1. **Messy parameter passing**: `Runner.RunWithMapping()` receives both `gobraCode` and `*HybridMapping` as separate parameters, creating tight coupling
2. **File I/O in wrong place**: `mapErrorsWithHybrid()` reads source files directly via `os.ReadFile()`, which is a side effect that shouldn't be in mapping logic
3. **Unused legacy code**: `mapping.go` contains deprecated types that are marked as deprecated but still exist
4. **Mixed responsibilities**: The `Runner` handles both execution and error mapping
5. **User-unfriendly output**: `VerificationError` contains generated code positions (`GenLine`, `GenColumn`, `GenFile`) that end users don't care about
6. **Inefficient data flow**: Source content is read multiple times and passed through multiple layers unnecessarily
7. **Unnecessary line:column conversion**: `VerificationError` contains line/column when plugins should handle formatting themselves

## Design Principles

1. **Separation of concerns**: Each component has a single, well-defined responsibility
2. **Explicit dependencies**: All required data is passed explicitly, no hidden file I/O
3. **Clean interfaces**: Use interfaces to allow different implementations
4. **Minimal data**: Only expose essential information, let plugins handle formatting
5. **Testability**: Each component can be tested independently

## New Architecture

### Core Interfaces

```go
// SourcePosition represents a position in source code
type SourcePosition struct {
    File   string
    Offset int
}

// ToLineCol converts offset to line:column for display
// Uses the shared OffsetToLineCol utility function
func (p *SourcePosition) ToLineCol(content []byte) (line, col int)

// SourceMapping provides mapping from generated to source code
type SourceMapping interface {
    // MapError maps a generated code offset to source position
    MapError(genOffset int) (SourcePosition, error)
    
    // MapErrorLineCol maps line:col to source position
    MapErrorLineCol(line, col int, genContent []byte) (SourcePosition, error)
}

// SourceContent represents source file content with metadata
type SourceContent struct {
    Path    string
    Content []byte
}

// VerificationRequest contains all data needed for verification
type VerificationRequest struct {
    GeneratedCode string
    Mapping       SourceMapping
    Source        *SourceContent // Optional, for line:col conversion
}

// Verifier handles Gobra verification
type Verifier interface {
    Verify(ctx context.Context, req VerificationRequest) (*VerificationResult, error)
}
```

### Simplified Result Types

```go
// VerificationResult contains verification results
type VerificationResult struct {
    Success   bool
    TimedOut  bool
    Errors    []VerificationError
    RawOutput string // For debugging only
}

// VerificationError represents an error in source code
type VerificationError struct {
    // Source location (offset-based only)
    File   string
    Offset int
    
    // Error details
    Kind    string
    Message string
    
    // Mapping metadata (for debugging)
    MappingInfo *ErrorMappingInfo
}

// ToLineCol converts offset to line:column for display
// This is a convenience method for plugins that need line:column format
// Uses the shared OffsetToLineCol utility function
func (e *VerificationError) ToLineCol(sourceContent []byte) (line, col int)

// ErrorMappingInfo provides metadata about error mapping
type ErrorMappingInfo struct {
    MappingType        MappingType
    AdditionKind       string
    HasPredicateMapping bool
    Confidence         float64
}
```

### Concrete Types

```go
// HybridMapping implements SourceMapping
type HybridMapping struct {
    additions     []Interval
    pointMappings []PointMapping
}

// NewHybridMapping creates empty mapping
func NewHybridMapping("") *HybridMapping

// AddInterval adds a code addition interval
func (h *HybridMapping) AddInterval(interval Interval) error

// AddPointMapping adds a point-to-point mapping
func (h *HybridMapping) AddPointMapping(mapping PointMapping)

// MapError implements SourceMapping
func (h *HybridMapping) MapError(genOffset int) (SourcePosition, error)

// MapErrorLineCol implements SourceMapping
func (h *HybridMapping) MapErrorLineCol(line, col int, genContent []byte) (SourcePosition, error)

// ErrorMapper handles error mapping with source content
type ErrorMapper struct {
    mapping SourceMapping
    source  *SourceContent
    genCode []byte
}

// NewErrorMapper creates mapper with all needed data
func NewErrorMapper(mapping SourceMapping, source *SourceContent, genCode string) *ErrorMapper

// MapError maps raw error to VerificationError (offset-based only)
// Uses OffsetToLineCol for any line:column conversion needed internally
func (m *ErrorMapper) MapError(raw rawError) VerificationError
```

### Simplified Runner

```go
// Runner focuses only on execution
type Runner struct {
    settings Settings
}

// NewRunner creates a new runner
func NewRunner(settings Settings) *Runner

// Verify executes Gobra verification
func (r *Runner) Verify(ctx context.Context, req VerificationRequest) (*VerificationResult, error) {
    // 1. Write temp file with generated code
    // 2. Execute Gobra
    // 3. Parse raw output
    // 4. Use ErrorMapper to map errors
    // 5. Return result
}

// VerifyWithoutMapping is a convenience method for simple cases
func (r *Runner) VerifyWithoutMapping(ctx context.Context, gobraCode string) (*VerificationResult, error)
```

## Refactoring Steps

### Phase 1: Prepare New Interfaces

1. Create new interface definitions in `interfaces.go`
2. Create simplified result types in `result.go`
3. Keep old types for backward compatibility during transition

### Phase 2: Refactor HybridMapping

1. Remove `SourceFile` field from `HybridMapping`
2. Update `NewHybridMapping("")` to not require file path
3. Ensure all mapping methods work with offsets only
4. Update tests to use new interface

### Phase 3: Create ErrorMapper

1. Create `ErrorMapper` type that combines mapping and source content
2. Move error mapping logic from `runner.go` to `ErrorMapper`
3. Remove file I/O from mapping logic
4. Add comprehensive tests

### Phase 4: Refactor Runner

1. Update `Runner` to use new `VerificationRequest` type
2. Replace `RunWithMapping()` with `Verify()`
3. Create `VerifyWithoutMapping()` for simple cases
4. Update all error mapping to use `ErrorMapper`
5. Remove generated code positions from `VerificationError`

### Phase 5: Update Plugin Integration

1. Update `GobraPlugin` to create `VerificationRequest`
2. Pass source content explicitly instead of file paths
3. Update translation result to use new mapping interface
4. Update tests

### Phase 6: Remove Legacy Code

1. Delete deprecated `mapping.go` file
2. Remove backward compatibility shims
3. Clean up unused imports and code
4. Update documentation

### Phase 7: Final Testing

1. Run all existing tests
2. Add integration tests for new interfaces
3. Test with real Gobra examples
4. Verify error messages are user-friendly

## File Changes

### Files to Modify

1. `framework/gobra_runner/hybrid_types.go` - Remove `SourceFile` field
2. `framework/gobra_runner/hybrid_mapping.go` - Update to new interface
3. `framework/gobra_runner/result.go` - Simplify `VerificationError`
4. `framework/gobra_runner/runner.go` - Use new interfaces
5. `framework/plugin/gobra/gobra_plugin.go` - Update to use new API
6. `framework/plugin/gobra/translate.go` - Update mapping creation

### Files to Create

1. `framework/gobra_runner/interfaces.go` - New interface definitions
2. `framework/gobra_runner/error_mapper.go` - Error mapping logic

### Files to Enhance

1. `framework/gobra_runner/hybrid_utils.go` - Make OffsetToLineCol more prominent and reusable

### Files to Delete

1. `framework/gobra_runner/mapping.go` - Deprecated legacy code

## Migration Strategy

### Direct Migration Approach

No backward compatibility shims - direct migration to new API:

1. Update all callers to use new interfaces immediately
2. Replace all old API calls with new API calls
3. Update all tests to use new interfaces
4. Remove old code completely

### Testing Strategy

1. Update all existing tests to use new API immediately
2. Add comprehensive tests for new interfaces
3. Add comprehensive tests for OffsetToLineCol utility function
4. Ensure all tests pass with new implementation
5. Remove any old test code

## Benefits

1. **Cleaner code**: Each component has single responsibility
2. **Better testability**: Components can be tested in isolation
3. **Minimal data**: Errors contain only offset, plugins handle formatting
4. **No hidden I/O**: All file operations are explicit
5. **Flexible interfaces**: Easy to add new mapping strategies
6. **Removed legacy**: No more deprecated code cluttering the codebase
7. **Reusable utilities**: OffsetToLineCol is used consistently across all offset-to-line:column conversions

## Risks and Mitigations

### Risk: Breaking changes for plugin authors

**Mitigation**: Direct migration - update all plugins immediately to use new API

### Risk: Performance regression from multiple content copies

**Mitigation**: Use byte slices efficiently, avoid unnecessary copies

### Risk: Complex migration process

**Mitigation**: Comprehensive testing ensures correctness, direct approach is cleaner

## Success Criteria

1. All existing tests pass
2. New tests cover new interfaces
3. Error messages contain only offset (no line/column, no generated positions)
4. No file I/O in mapping logic
5. Legacy code removed
6. Documentation updated
7. Real Gobra examples work correctly
8. Plugins can format errors as needed using ToLineCol() convenience method
9. OffsetToLineCol is used consistently for all offset-to-line:column conversions
10. OffsetToLineCol has comprehensive test coverage

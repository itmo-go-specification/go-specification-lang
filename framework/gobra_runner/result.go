package gobra_runner

import "fmt"

type VerificationResult struct {
	Success   bool
	TimedOut  bool
	Errors    []VerificationError
	RawOutput string
}

// VerificationError represents an error reported by the Gobra verifier.
// It contains information about both the generated code location and the
// corresponding source code location (when available through mapping).
type VerificationError struct {
	// File is the source file path where the error occurred (empty if unmapped)
	File string

	// Line is the line number in the source file (0 if unmapped)
	Line int

	// Column is the column number in the source file (0 if unmapped)
	Column int

	// Offset is the byte offset in the source file (0 if unmapped)
	Offset int

	// Kind specifies the type of error (e.g., "assertion failed", "precondition violated")
	Kind string

	// Message is the detailed error message from the verifier
	Message string

	// GenLine is the line number in the generated Gobra code
	GenLine int

	// GenColumn is the column number in the generated Gobra code
	GenColumn int

	// GenFile is the path to the generated Gobra file (optional)
	GenFile string

	// MappingInfo contains additional information about how this error was mapped
	// (e.g., whether it was mapped via an interval or point mapping)
	MappingInfo *ErrorMappingInfo `json:"mappingInfo,omitempty"`
}

// ErrorMappingInfo provides metadata about how an error was mapped from
// generated code to source code.
type ErrorMappingInfo struct {
	// MappingType indicates how the error was mapped
	MappingType MappingType `json:"mappingType"`

	// AdditionKind specifies the type of addition if the error was in added code (optional string)
	AdditionKind string `json:"additionKind,omitempty"`

	// HasPredicateMapping indicates whether a precise predicate mapping was available
	HasPredicateMapping bool `json:"hasPredicateMapping"`

	// Confidence indicates how confident we are in this mapping (0.0 to 1.0)
	Confidence float64 `json:"confidence"`
}

// MappingType represents the method used to map an error from generated to source code
type MappingType string

const (
	// MappingTypePoint indicates the error was mapped using a point-to-point mapping
	MappingTypePoint MappingType = "point"

	// MappingTypeInterval indicates the error was mapped using an interval mapping
	MappingTypeInterval MappingType = "interval"

	// MappingTypeFallback indicates the error was mapped using the fallback algorithm
	MappingTypeFallback MappingType = "fallback"

	// MappingTypeUnmapped indicates the error could not be mapped to source code
	MappingTypeUnmapped MappingType = "unmapped"
)

// IsMapped returns true if the error has been successfully mapped to source code
func (e VerificationError) IsMapped() bool {
	return e.File != "" && e.Line > 0
}

// IsInGeneratedCode returns true if the error is in generated code that has no
// corresponding source location
func (e VerificationError) IsInGeneratedCode() bool {
	return !e.IsMapped()
}

// IsInAddedCode returns true if the error is in code that was added during translation
func (e VerificationError) IsInAddedCode() bool {
	return e.MappingInfo != nil &&
		(e.MappingInfo.MappingType == MappingTypeInterval ||
			e.MappingInfo.MappingType == MappingTypePoint)
}

// IsUnmapped returns true if the error could not be mapped to source code
// This typically indicates a problem with the mapping system or generated code
func (e VerificationError) IsUnmapped() bool {
	return e.MappingInfo != nil && e.MappingInfo.MappingType == MappingTypeUnmapped
}

// String returns a human-readable representation of the error
func (e VerificationError) String() string {
	loc := fmt.Sprintf("%s:%d:%d", e.File, e.Line, e.Column)
	if e.File == "" {
		loc = fmt.Sprintf("<generated>:%d:%d", e.GenLine, e.GenColumn)
		if e.GenFile != "" {
			loc = fmt.Sprintf("%s:%d:%d", e.GenFile, e.GenLine, e.GenColumn)
		}
	}
	return fmt.Sprintf("%s: %s: %s", loc, e.Kind, e.Message)
}

// DetailedString returns a more detailed representation including mapping information
func (e VerificationError) DetailedString() string {
	base := e.String()
	if e.MappingInfo != nil {
		return fmt.Sprintf("%s [mapped via %s, confidence: %.2f]",
			base, e.MappingInfo.MappingType, e.MappingInfo.Confidence)
	}
	return base
}

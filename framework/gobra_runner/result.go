package gobra_runner

type VerificationResult struct {
	Success   bool
	TimedOut  bool
	Errors    []VerificationError
	RawOutput string
}

// VerificationError represents an error in source code
type VerificationError struct {
	// Position is the source position where the error occurred
	Position SourcePosition

	// Kind specifies the type of error (e.g., "assertion failed", "precondition violated")
	Kind string

	// Message is the detailed error message from the verifier
	Message string
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

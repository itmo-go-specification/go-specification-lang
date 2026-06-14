package gobra_runner

import "context"

// SourcePosition represents a position in source code using byte offset
type SourcePosition struct {
	File   string
	Offset int // Byte offset in the file (0-based)
}

// ToLineCol converts offset to line:column for display
// Uses the shared OffsetToLineCol utility function
func (p *SourcePosition) ToLineCol(content []byte) (line, col int) {
	line, col, _ = OffsetToLineCol(content, p.Offset)
	return line, col
}

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
}

// Verifier handles Gobra verification
type Verifier interface {
	Verify(ctx context.Context, req VerificationRequest) (*VerificationResult, error)
}

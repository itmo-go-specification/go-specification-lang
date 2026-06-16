package gobra_runner

// ErrorMapper handles error mapping with source content
type ErrorMapper struct {
	mapping SourceMapping
	genCode []byte
}

// NewErrorMapper creates mapper with all needed data
func NewErrorMapper(mapping SourceMapping, genCode string) *ErrorMapper {
	return &ErrorMapper{
		mapping: mapping,
		genCode: []byte(genCode),
	}
}

// MapError maps raw error to VerificationError (offset-based only)
func (m *ErrorMapper) MapError(raw rawError) VerificationError {
	ve := VerificationError{
		Message: raw.message,
	}

	// Try to map using hybrid mapping
	if m.mapping != nil {
		// Convert line:col to offset for internal processing
		pos, err := m.mapping.MapErrorLineCol(raw.genLine, raw.genColumn, m.genCode)
		if err == nil {
			ve.Position = pos
		} else {
		}
	}

	// Infer error kind from message
	if ve.Kind == "" {
		ve.Kind = inferKind(raw.message)
	}

	return ve
}

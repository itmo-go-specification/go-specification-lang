package gobra_runner

import "fmt"

type VerificationResult struct {
	Success   bool
	TimedOut  bool
	Errors    []VerificationError
	RawOutput string
}

type VerificationError struct {
	File   string
	Line   int
	Column int

	Kind string

	Message string

	GenLine   int
	GenColumn int
}

func (e VerificationError) String() string {
	loc := fmt.Sprintf("%s:%d:%d", e.File, e.Line, e.Column)
	if e.File == "" {
		loc = fmt.Sprintf("<generated>:%d:%d", e.GenLine, e.GenColumn)
	}
	return fmt.Sprintf("%s: %s: %s", loc, e.Kind, e.Message)
}

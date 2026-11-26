package formats

import (
	"encoding/json"
	"fmt"

	"gpma/framework/errors"
)

// VSCodeFormatter implements ErrorFormatter for VS Code diagnostics format
type VSCodeFormatter struct{}

// NewVSCodeFormatter creates a new VS Code formatter
func NewVSCodeFormatter() *VSCodeFormatter {
	return &VSCodeFormatter{}
}

// FormatErrors formats a slice of errors as VS Code diagnostics
func (f *VSCodeFormatter) FormatErrors(errs []*errors.Error, config errors.FormatterConfig) (string, error) {
	if len(errs) == 0 {
		return `{"diagnostics":[]}`, nil
	}

	diagnostics := make([]map[string]interface{}, 0, len(errs))
	for _, err := range errs {
		diag, err := f.errorToDiagnostic(err)
		if err != nil {
			return "", err
		}
		diagnostics = append(diagnostics, diag)
	}

	output := map[string]interface{}{
		"diagnostics": diagnostics,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal diagnostics to JSON: %w", err)
	}

	return string(data), nil
}

// FormatError formats a single error as VS Code diagnostic
func (f *VSCodeFormatter) FormatError(err *errors.Error, config errors.FormatterConfig) (string, error) {
	diag, diagErr := f.errorToDiagnostic(err)
	if diagErr != nil {
		return "", diagErr
	}

	data, marshalErr := json.MarshalIndent(diag, "", "  ")
	if marshalErr != nil {
		return "", fmt.Errorf("failed to marshal diagnostic to JSON: %w", marshalErr)
	}

	return string(data), nil
}

// errorToDiagnostic converts an Error to VS Code diagnostic format
func (f *VSCodeFormatter) errorToDiagnostic(err *errors.Error) (map[string]interface{}, error) {
	diag := map[string]interface{}{
		"message": err.Message,
		"source":  "gpma",
	}

	// Set severity (VS Code uses: 1=Error, 2=Warning, 3=Info, 4=Hint)
	diag["severity"] = f.severityToVSCode(err.Severity)

	// Set range
	if err.Range != nil {
		diag["range"] = map[string]interface{}{
			"start": map[string]interface{}{
				"line":      err.Range.Start.Line - 1,   // VS Code uses 0-based lines
				"character": err.Range.Start.Column - 1, // VS Code uses 0-based columns
			},
			"end": map[string]interface{}{
				"line":      err.Range.End.Line - 1,
				"character": err.Range.End.Column - 1,
			},
		}
	} else {
		// Default to single character at position
		diag["range"] = map[string]interface{}{
			"start": map[string]interface{}{
				"line":      err.Position.Line - 1,
				"character": err.Position.Column - 1,
			},
			"end": map[string]interface{}{
				"line":      err.Position.Line - 1,
				"character": err.Position.Column, // Single character range
			},
		}
	}

	return diag, nil
}

// severityToVSCode converts Severity to VS Code severity number
func (f *VSCodeFormatter) severityToVSCode(severity errors.Severity) int {
	switch severity {
	case errors.SeverityCritical:
		return 1 // Error
	case errors.SeverityWarning:
		return 2 // Warning
	case errors.SeverityInfo:
		return 3 // Info
	default:
		return 0 // Default to Error
	}
}

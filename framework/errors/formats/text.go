package formats

import (
	"fmt"
	"strings"

	"gpma/framework/errors"
)

// TextFormatter implements ErrorFormatter for human-readable text output
type TextFormatter struct{}

// NewTextFormatter creates a new text formatter
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}

// FormatErrors formats a slice of errors as text
func (f *TextFormatter) FormatErrors(errs []*errors.Error, config errors.FormatterConfig) (string, error) {
	if len(errs) == 0 {
		return "", nil
	}

	var builder strings.Builder
	for _, err := range errs {
		line, err := f.FormatError(err, config)
		if err != nil {
			return "", err
		}
		builder.WriteString(line)
		builder.WriteString("\n")
	}

	return strings.TrimSpace(builder.String()), nil
}

// FormatError formats a single error as text
func (f *TextFormatter) FormatError(err *errors.Error, config errors.FormatterConfig) (string, error) {
	severityLabel := f.severityLabel(err.Severity)

	position := fmt.Sprintf("%s:%d:%d", err.Position.File, err.Position.Line, err.Position.Column)

	if err.Source != "" {
		return fmt.Sprintf("[%s] %s: %s: %s", err.Source, position, severityLabel, err.Message), nil
	}
	return fmt.Sprintf("%s: %s: %s", position, severityLabel, err.Message), nil
}

// severityLabel returns a human-readable label for the severity
func (f *TextFormatter) severityLabel(severity errors.Severity) string {
	switch severity {
	case errors.SeverityCritical:
		return "error"
	case errors.SeverityWarning:
		return "warning"
	case errors.SeverityInfo:
		return "info"
	default:
		return string(severity)
	}
}

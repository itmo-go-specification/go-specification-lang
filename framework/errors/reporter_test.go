package errors

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

type testJSONFormatter struct{}

func (testJSONFormatter) FormatError(err *Error, _ FormatterConfig) (string, error) {
	data, marshalErr := json.Marshal(err)
	return string(data), marshalErr
}

func (testJSONFormatter) FormatErrors(errs []*Error, _ FormatterConfig) (string, error) {
	data, marshalErr := json.Marshal(errs)
	return string(data), marshalErr
}

func TestErrorReporterImpl_ReportError(t *testing.T) {
	tests := []struct {
		name      string
		error     *Error
		format    OutputFormat
		wantError bool
	}{
		{
			name: "report error with text format",
			error: &Error{
				Position: Position{
					File:   "test.go",
					Line:   10,
					Column: 5,
				},
				Message:  "test error",
				Severity: SeverityCritical,
				Source:   "test-plugin",
			},
			format:    FormatText,
			wantError: false,
		},
		{
			name: "report error with json format",
			error: &Error{
				Position: Position{
					File:   "test.go",
					Line:   10,
					Column: 5,
				},
				Message:  "test error",
				Severity: SeverityCritical,
				Source:   "test-plugin",
			},
			format:    FormatJSON,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printer := NewStreamPrinter(&buf)

			formatter, err := getFormatter(tt.format)
			if err != nil {
				t.Fatalf("Failed to get formatter: %v", err)
			}

			reporter := &ErrorReporterImpl{
				formatter: formatter,
				printer:   printer,
				config: ReporterConfig{
					Format: tt.format,
				},
			}

			reportErr := reporter.ReportError(tt.error)
			if (reportErr != nil) != tt.wantError {
				t.Errorf("ErrorReporterImpl.ReportError() error = %v, wantError %v", reportErr, tt.wantError)
				return
			}

			output := buf.String()
			if output == "" {
				t.Errorf("Expected output, got empty string")
			}

			if !strings.Contains(output, tt.error.Message) {
				t.Errorf("Expected output to contain error message, got: %s", output)
			}
		})
	}
}

func TestErrorReporterImpl_ReportErrors(t *testing.T) {
	errors := []*Error{
		{
			Position: Position{
				File:   "test.go",
				Line:   10,
				Column: 5,
			},
			Message:  "first error",
			Severity: SeverityCritical,
			Source:   "test-plugin",
		},
		{
			Position: Position{
				File:   "test.go",
				Line:   20,
				Column: 10,
			},
			Message:  "second error",
			Severity: SeverityWarning,
			Source:   "test-plugin",
		},
	}

	var buf bytes.Buffer
	printer := NewStreamPrinter(&buf)

	formatter, err := getFormatter(FormatText)
	if err != nil {
		t.Fatalf("Failed to get formatter: %v", err)
	}

	reporter := &ErrorReporterImpl{
		formatter: formatter,
		printer:   printer,
		config: ReporterConfig{
			Format: FormatText,
		},
	}

	reportErr := reporter.ReportErrors(errors)
	if reportErr != nil {
		t.Errorf("ErrorReporterImpl.ReportErrors() error = %v", reportErr)
		return
	}

	output := buf.String()
	if output == "" {
		t.Errorf("Expected output, got empty string")
	}

	if !strings.Contains(output, "first error") || !strings.Contains(output, "second error") {
		t.Errorf("Expected output to contain both error messages, got: %s", output)
	}
}

func TestErrorReporterImpl_SetFormatter(t *testing.T) {
	var buf bytes.Buffer
	printer := NewStreamPrinter(&buf)

	formatter, err := getFormatter(FormatText)
	if err != nil {
		t.Fatalf("Failed to get formatter: %v", err)
	}

	reporter := &ErrorReporterImpl{
		formatter: formatter,
		printer:   printer,
		config: ReporterConfig{
			Format: FormatText,
		},
	}

	// Change formatter to JSON without depending on formatter_factory registration.
	reporter.SetFormatter(testJSONFormatter{})

	testError := &Error{
		Position: Position{
			File:   "test.go",
			Line:   10,
			Column: 5,
		},
		Message:  "test error",
		Severity: SeverityCritical,
		Source:   "test-plugin",
	}

	reportErr := reporter.ReportError(testError)
	if reportErr != nil {
		t.Errorf("ErrorReporterImpl.ReportError() error = %v", reportErr)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "{") || !strings.Contains(output, "}") {
		t.Errorf("Expected JSON output, got: %s", output)
	}
}

func TestErrorReporterImpl_SetPrinter(t *testing.T) {
	var buf1 bytes.Buffer
	var buf2 bytes.Buffer

	printer1 := NewStreamPrinter(&buf1)
	printer2 := NewStreamPrinter(&buf2)

	formatter, err := getFormatter(FormatText)
	if err != nil {
		t.Fatalf("Failed to get formatter: %v", err)
	}

	reporter := &ErrorReporterImpl{
		formatter: formatter,
		printer:   printer1,
		config: ReporterConfig{
			Format: FormatText,
		},
	}

	testError := &Error{
		Position: Position{
			File:   "test.go",
			Line:   10,
			Column: 5,
		},
		Message:  "test error",
		Severity: SeverityCritical,
		Source:   "test-plugin",
	}

	// Report to first printer
	reportErr := reporter.ReportError(testError)
	if reportErr != nil {
		t.Errorf("ErrorReporterImpl.ReportError() error = %v", reportErr)
		return
	}

	// Change printer
	reporter.SetPrinter(printer2)

	// Report to second printer
	err = reporter.ReportError(testError)
	if err != nil {
		t.Errorf("ErrorReporterImpl.ReportError() error = %v", err)
		return
	}

	if buf1.String() == "" {
		t.Errorf("Expected output in first buffer")
	}

	if buf2.String() == "" {
		t.Errorf("Expected output in second buffer")
	}
}

func TestNewErrorReporter(t *testing.T) {
	tests := []struct {
		name    string
		config  ReporterConfig
		wantErr bool
	}{
		{
			name: "create reporter with stdout",
			config: ReporterConfig{
				Format:            FormatText,
				OutputDestination: DestinationStdout,
			},
			wantErr: false,
		},
		{
			name: "create reporter with stderr",
			config: ReporterConfig{
				Format:            FormatText,
				OutputDestination: DestinationStderr,
			},
			wantErr: false,
		},
		{
			name: "create reporter with file",
			config: ReporterConfig{
				Format:            FormatText,
				OutputDestination: DestinationFile,
				FilePath:          "/tmp/test_errors.log",
				Append:            false,
			},
			wantErr: false,
		},
		{
			name: "create reporter with file but no path",
			config: ReporterConfig{
				Format:            FormatText,
				OutputDestination: DestinationFile,
				FilePath:          "",
				Append:            false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter, err := NewErrorReporter(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewErrorReporter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && reporter != nil {
				if err := reporter.Close(); err != nil {
					t.Errorf("failed to close reporter: %v", err)
				}
			}
		})
	}
}

func TestReporterError(t *testing.T) {
	err := &ReporterError{Message: "test error"}
	if err.Error() != "test error" {
		t.Errorf("ReporterError.Error() = %v, want %v", err.Error(), "test error")
	}

	if !IsReporterError(err) {
		t.Errorf("IsReporterError() = false, want true")
	}

	// Test with nil
	if IsReporterError(nil) {
		t.Errorf("IsReporterError() = true for nil, want false")
	}
}

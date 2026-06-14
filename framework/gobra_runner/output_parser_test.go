package gobra_runner

import (
	"testing"
)

func TestParseGobraOutput_primaryPattern(t *testing.T) {
	output := `[INFO] Starting Gobra...
[ERROR] /tmp/gobra_input_123.go@10.16: Precondition of call might not hold. Assertion x >= 0 might not hold.
[ERROR] /tmp/gobra_input_123.go@14.1: Postcondition might not hold. Assertion result >= 0 might not hold.
[ERROR] verification failed in 2.4 seconds`

	genFile := "/tmp/gobra_input_123.go"
	errors := parseGobraOutput(output, genFile)

	if len(errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errors))
	}

	if errors[0].genLine != 10 || errors[0].genColumn != 16 {
		t.Errorf("error[0]: expected @10.16, got @%d.%d", errors[0].genLine, errors[0].genColumn)
	}
	if errors[1].genLine != 14 || errors[1].genColumn != 1 {
		t.Errorf("error[1]: expected @14.1, got @%d.%d", errors[1].genLine, errors[1].genColumn)
	}
}

func TestParseGobraOutput_altPattern(t *testing.T) {
	output := `ERROR at '/tmp/gobra_abc.go' @ (5, 3): Precondition of call might not hold.`

	errors := parseGobraOutput(output, "/tmp/gobra_abc.go")
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if errors[0].genLine != 5 || errors[0].genColumn != 3 {
		t.Errorf("expected (5,3), got (%d,%d)", errors[0].genLine, errors[0].genColumn)
	}
}

func TestParseGobraOutput_filtersByFile(t *testing.T) {
	output := `[ERROR] /other/file.go@1.1: some error
[ERROR] /tmp/mine.go@5.2: my error`

	errors := parseGobraOutput(output, "/tmp/mine.go")
	if len(errors) != 1 {
		t.Fatalf("expected 1 error after filtering, got %d", len(errors))
	}
	if errors[0].genLine != 5 {
		t.Errorf("expected line 5, got %d", errors[0].genLine)
	}
}

func TestParseGobraOutput_empty(t *testing.T) {
	output := "[INFO] Successfully verified 1 member.\n"
	errors := parseGobraOutput(output, "/tmp/x.go")
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d", len(errors))
	}
}

func TestInferKind(t *testing.T) {
	cases := []struct {
		msg  string
		want string
	}{
		{"Precondition of call might not hold", "precondition"},
		{"Postcondition might not hold", "postcondition"},
		{"Loop invariant might not be preserved", "invariant"},
		{"Assert statement might not hold", "assertion"},
		{"Unknown error occurred", "unknown"},
	}
	for _, c := range cases {
		got := inferKind(c.msg)
		if got != c.want {
			t.Errorf("inferKind(%q) = %q, want %q", c.msg, got, c.want)
		}
	}
}

// TestCoordMappingLookup removed - CoordMapping is deprecated and removed
// Use HybridMapping tests instead

func TestMapErrors_withMapping(t *testing.T) {
	// Create a simple test with direct offset mapping
	mapping := NewHybridMapping("")

	// Simple generated code where we can easily calculate offsets
	genCode := "package main\n\nfunc main() {\n\tx := 1\n\treturn x\n}\n"

	// Add an interval that maps a section in generated code to source code
	// The generated code has "x := 1" starting at offset 28
	mapping.AddInterval(Interval{
		GenStart:  28,                                         // Start of "x := 1"
		GenEnd:    35,                                         // End of "x := 1"
		SourcePos: SourcePosition{File: "abs.go", Offset: 20}, // Maps to offset 20 in source
	})

	// Create an error at line 4, column 3 (which is the "x" in "x := 1")
	raw := []rawError{
		{genLine: 4, genColumn: 3, message: "Precondition might not hold"},
	}

	result := mapErrorsWithHybrid(raw, mapping, genCode)
	if len(result) != 1 {
		t.Fatalf("expected 1 mapped error, got %d", len(result))
	}

	ve := result[0]
	if ve.Position.File != "abs.go" {
		t.Errorf("Position.File: got %q, want %q", ve.Position.File, "abs.go")
	}
	// Check that offset is mapped correctly
	if ve.Position.Offset != 20 {
		t.Errorf("Position.Offset: got %d, want 20", ve.Position.Offset)
	}
	if ve.Kind != "precondition" {
		t.Errorf("Kind: got %q, want precondition", ve.Kind)
	}
	// Note: GenLine and GenColumn are no longer exposed in VerificationError
	// The new API only provides offset-based source positions
}

func TestMapErrors_nilMapping(t *testing.T) {
	raw := []rawError{
		{genLine: 10, genColumn: 1, message: "Postcondition might not hold"},
	}

	genCode := "package main\n\nfunc main() {\n\t// Line 1\n\t// Line 2\n\t// Line 3\n\t// Line 4\n\t// Line 5\n\t// Line 6\n\t// Line 7\n\t// Line 8\n\t// Line 9\n\t// Line 10\n}\n"

	result := mapErrorsWithHybrid(raw, nil, genCode)
	if len(result) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result))
	}

	ve := result[0]
	if ve.Position.File != "" {
		t.Errorf("expected empty Position.File without mapping, got %q", ve.Position.File)
	}
	if ve.Kind != "postcondition" {
		t.Errorf("Kind inferred wrong: got %q", ve.Kind)
	}
}

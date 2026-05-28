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

func TestCoordMappingLookup(t *testing.T) {
	m := CoordMapping{
		Entries: []MappingEntry{
			{GenLine: 5, GenColumn: 1, Orig: SourcePos{File: "orig.go", Line: 3, Column: 1}, Kind: "precondition"},
			{GenLine: 10, GenColumn: 1, Orig: SourcePos{File: "orig.go", Line: 7, Column: 1}, Kind: "postcondition"},
			{GenLine: 15, GenColumn: 3, Orig: SourcePos{File: "orig.go", Line: 12, Column: 3}, Kind: "assertion"},
		},
	}

	tests := []struct {
		line         int
		col          int
		wantOk       bool
		wantKind     string
		wantOrigLine int
	}{
		{5, 1, true, "precondition", 3},
		{7, 1, true, "precondition", 3},
		{10, 1, true, "postcondition", 7},
		{14, 5, true, "postcondition", 7},
		{15, 3, true, "assertion", 12},
		{20, 1, true, "assertion", 12},
		{4, 1, false, "", 0},
	}

	for _, tt := range tests {
		entry, ok := m.Lookup(tt.line, tt.col)
		if ok != tt.wantOk {
			t.Errorf("Lookup(%d,%d): ok=%v, want %v", tt.line, tt.col, ok, tt.wantOk)
			continue
		}
		if !tt.wantOk {
			continue
		}
		if entry.Kind != tt.wantKind {
			t.Errorf("Lookup(%d,%d): kind=%q, want %q", tt.line, tt.col, entry.Kind, tt.wantKind)
		}
		if entry.Orig.Line != tt.wantOrigLine {
			t.Errorf("Lookup(%d,%d): orig.Line=%d, want %d", tt.line, tt.col, entry.Orig.Line, tt.wantOrigLine)
		}
	}
}

func TestMapErrors_withMapping(t *testing.T) {
	mapping := &CoordMapping{
		Entries: []MappingEntry{
			{GenLine: 5, GenColumn: 1, Orig: SourcePos{File: "abs.go", Line: 3, Column: 1}, Kind: "precondition"},
		},
	}

	raw := []rawError{
		{genLine: 5, genColumn: 16, message: "Precondition might not hold"},
	}

	result := mapErrors(raw, mapping)
	if len(result) != 1 {
		t.Fatalf("expected 1 mapped error, got %d", len(result))
	}

	ve := result[0]
	if ve.File != "abs.go" {
		t.Errorf("File: got %q, want %q", ve.File, "abs.go")
	}
	if ve.Line != 3 {
		t.Errorf("Line: got %d, want 3", ve.Line)
	}
	if ve.Kind != "precondition" {
		t.Errorf("Kind: got %q, want precondition", ve.Kind)
	}
	if ve.GenLine != 5 || ve.GenColumn != 16 {
		t.Errorf("GenLine/Col: got %d/%d, want 5/16", ve.GenLine, ve.GenColumn)
	}
}

func TestMapErrors_nilMapping(t *testing.T) {
	raw := []rawError{
		{genLine: 10, genColumn: 1, message: "Postcondition might not hold"},
	}

	result := mapErrors(raw, nil)
	if len(result) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result))
	}

	ve := result[0]
	if ve.File != "" {
		t.Errorf("expected empty File without mapping, got %q", ve.File)
	}
	if ve.Kind != "postcondition" {
		t.Errorf("Kind inferred wrong: got %q", ve.Kind)
	}
}

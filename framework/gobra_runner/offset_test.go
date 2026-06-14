package gobra_runner

import (
	"testing"
)

// TestOffsetBasedMapping tests the core offset-based mapping functionality
func TestOffsetBasedMapping(t *testing.T) {
	// Create a simple mapping
	mapping := NewHybridMapping("")

	// Add an interval using offsets (not line/column)
	err := mapping.AddInterval(Interval{
		GenStart:  100,                                         // Offset in generated code
		GenEnd:    200,                                         // Offset in generated code
		SourcePos: SourcePosition{File: "test.go", Offset: 50}, // Offset in original code
	})
	if err != nil {
		t.Fatalf("Failed to add interval: %v", err)
	}

	// Test that the interval contains offsets correctly
	interval := mapping.FindInterval(150) // Should be inside the interval
	if interval == nil {
		t.Fatal("Expected to find interval at offset 150")
	}
	if interval.SourcePos.Offset != 50 {
		t.Errorf("Expected source offset 50, got %d", interval.SourcePos.Offset)
	}

	// Test that offsets outside the interval are not found
	interval = mapping.FindInterval(50) // Should be outside
	if interval != nil {
		t.Error("Expected not to find interval at offset 50")
	}

	// Test point mapping
	mapping.AddPointMapping(PointMapping{
		GenOffset: 150,                                         // Offset in generated code
		Orig:      SourcePosition{File: "test.go", Offset: 75}, // Offset in original code
	})

	pointMapping := mapping.FindPointMapping(150)
	if pointMapping == nil {
		t.Fatal("Expected to find point mapping at offset 150")
	}
	if pointMapping.Orig.Offset != 75 {
		t.Errorf("Expected source offset 75, got %d", pointMapping.Orig.Offset)
	}

	// Test MapError with offset
	pos, err := mapping.MapError(150)
	if err != nil {
		t.Fatalf("MapError failed: %v", err)
	}
	if pos.Offset != 75 {
		t.Errorf("Expected mapped offset 75, got %d", pos.Offset)
	}

	// Test MapErrorLineCol with conversion to/from line:col
	genCode := "package main\n\nfunc main() {\n\t// Line 1\n\t// Line 2\n\t// Line 3\n\t// Line 4\n\t// Line 5\n\t// Line 6\n\t// Line 7\n\t// Line 8\n\t// Line 9\n\t// Line 10\n}\n"

	// Convert line:col to offset
	offset, err := LineColToOffset([]byte(genCode), 5, 10)
	if err != nil {
		t.Fatalf("LineColToOffset failed: %v", err)
	}
	t.Logf("Line 5, Column 10 -> Offset %d", offset)

	// Convert offset back to line:col
	testPos := SourcePosition{Offset: offset}
	line, col := testPos.ToLineCol([]byte(genCode))
	t.Logf("Offset %d -> Line %d, Column %d", offset, line, col)

	if line != 5 || col != 10 {
		t.Errorf("Expected Line 5, Column 10, got Line %d, Column %d", line, col)
	}
}

// TestPositionToLineCol tests the Position.ToLineCol conversion
func TestPositionToLineCol(t *testing.T) {
	content := []byte("line1\nline2\nline3\nline4\nline5")

	tests := []struct {
		offset   int
		wantLine int
		wantCol  int
	}{
		{0, 1, 1},  // Start of file
		{4, 1, 5},  // End of line1
		{5, 1, 6},  // Newline character at end of line1
		{6, 2, 1},  // Start of line2
		{12, 3, 1}, // Start of line3
		{24, 5, 1}, // Start of line5
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			pos := SourcePosition{Offset: tt.offset}
			line, col := pos.ToLineCol(content)
			if line != tt.wantLine || col != tt.wantCol {
				t.Errorf("Offset %d: got Line %d, Column %d, want Line %d, Column %d",
					tt.offset, line, col, tt.wantLine, tt.wantCol)
			}
		})
	}
}

package gobra_runner

import (
	"testing"
)

func TestInterval_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		interval Interval
		want     bool
	}{
		{
			name: "valid interval",
			interval: Interval{
				GenStart:  10,
				GenEnd:    20,
				SourcePos: Position{File: "test.go", Line: 5},
			},
			want: true,
		},
		{
			name: "valid interval with columns",
			interval: Interval{
				GenStart:    10,
				GenEnd:      10,
				GenStartCol: 5,
				GenEndCol:   15,
				SourcePos:   Position{File: "test.go", Line: 5},
			},
			want: true,
		},
		{
			name: "invalid - start > end",
			interval: Interval{
				GenStart:  20,
				GenEnd:    10,
				SourcePos: Position{File: "test.go", Line: 5},
			},
			want: false,
		},
		{
			name: "invalid - negative start",
			interval: Interval{
				GenStart:  -1,
				GenEnd:    10,
				SourcePos: Position{File: "test.go", Line: 5},
			},
			want: false,
		},
		{
			name: "invalid - column start > column end",
			interval: Interval{
				GenStart:    10,
				GenEnd:      10,
				GenStartCol: 15,
				GenEndCol:   5,
				SourcePos:   Position{File: "test.go", Line: 5},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interval.IsValid(); got != tt.want {
				t.Errorf("Interval.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterval_Contains(t *testing.T) {
	tests := []struct {
		name     string
		interval Interval
		line     int
		col      int
		want     bool
	}{
		{
			name: "position inside interval",
			interval: Interval{
				GenStart: 10,
				GenEnd:   20,
			},
			line: 15,
			col:  5,
			want: true,
		},
		{
			name: "position before interval",
			interval: Interval{
				GenStart: 10,
				GenEnd:   20,
			},
			line: 5,
			col:  5,
			want: false,
		},
		{
			name: "position after interval",
			interval: Interval{
				GenStart: 10,
				GenEnd:   20,
			},
			line: 25,
			col:  5,
			want: false,
		},
		{
			name: "position on start line before column",
			interval: Interval{
				GenStart:    10,
				GenEnd:      20,
				GenStartCol: 5,
			},
			line: 10,
			col:  3,
			want: false,
		},
		{
			name: "position on start line after column",
			interval: Interval{
				GenStart:    10,
				GenEnd:      20,
				GenStartCol: 5,
			},
			line: 10,
			col:  7,
			want: true,
		},
		{
			name: "position on end line after column",
			interval: Interval{
				GenStart:  10,
				GenEnd:    20,
				GenEndCol: 15,
			},
			line: 20,
			col:  20,
			want: false,
		},
		{
			name: "position on end line before column",
			interval: Interval{
				GenStart:  10,
				GenEnd:    20,
				GenEndCol: 15,
			},
			line: 20,
			col:  10,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interval.Contains(tt.line, tt.col); got != tt.want {
				t.Errorf("Interval.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterval_Length(t *testing.T) {
	tests := []struct {
		name     string
		interval Interval
		want     int
	}{
		{
			name: "single line interval",
			interval: Interval{
				GenStart: 10,
				GenEnd:   10,
			},
			want: 1,
		},
		{
			name: "multi line interval",
			interval: Interval{
				GenStart: 10,
				GenEnd:   20,
			},
			want: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interval.Length(); got != tt.want {
				t.Errorf("Interval.Length() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHybridMapping_NewHybridMapping(t *testing.T) {
	mapping := NewHybridMapping("source.go")

	if mapping.SourceFile != "source.go" {
		t.Errorf("NewHybridMapping() SourceFile = %v, want %v", mapping.SourceFile, "source.go")
	}
	if len(mapping.Additions) != 0 {
		t.Errorf("NewHybridMapping() Additions length = %v, want 0", len(mapping.Additions))
	}
	if len(mapping.PointMappings) != 0 {
		t.Errorf("NewHybridMapping() PointMappings length = %v, want 0", len(mapping.PointMappings))
	}
}

func TestHybridMapping_AddInterval(t *testing.T) {
	mapping := NewHybridMapping("source.go")

	interval := Interval{
		GenStart:  10,
		GenEnd:    20,
		SourcePos: Position{File: "source.go", Line: 5},
	}

	err := mapping.AddInterval(interval)
	if err != nil {
		t.Errorf("AddInterval() error = %v", err)
	}

	if len(mapping.Additions) != 1 {
		t.Errorf("AddInterval() Additions length = %v, want 1", len(mapping.Additions))
	}
}

func TestHybridMapping_AddInterval_Invalid(t *testing.T) {
	mapping := NewHybridMapping("source.go")

	invalidInterval := Interval{
		GenStart:  20,
		GenEnd:    10,
		SourcePos: Position{File: "source.go", Line: 5},
	}

	err := mapping.AddInterval(invalidInterval)
	if err == nil {
		t.Errorf("AddInterval() expected error for invalid interval, got nil")
	}
}

func TestHybridMapping_AddInterval_Overlap(t *testing.T) {
	mapping := NewHybridMapping("source.go")

	interval1 := Interval{
		GenStart:  10,
		GenEnd:    20,
		SourcePos: Position{File: "source.go", Line: 5},
	}

	interval2 := Interval{
		GenStart:  15,
		GenEnd:    25,
		SourcePos: Position{File: "source.go", Line: 10},
	}

	mapping.AddInterval(interval1)
	err := mapping.AddInterval(interval2)

	if err == nil {
		t.Errorf("AddInterval() expected error for overlapping intervals, got nil")
	}
}

func TestHybridMapping_AutoSort(t *testing.T) {
	mapping := NewHybridMapping("source.go")

	mapping.AddInterval(Interval{
		GenStart:  30,
		GenEnd:    40,
		SourcePos: Position{File: "source.go", Line: 15},
	})

	mapping.AddInterval(Interval{
		GenStart:  10,
		GenEnd:    20,
		SourcePos: Position{File: "source.go", Line: 5},
	})

	mapping.AddInterval(Interval{
		GenStart:  22,
		GenEnd:    28,
		SourcePos: Position{File: "source.go", Line: 10},
	})

	// Intervals should be automatically sorted
	if len(mapping.Additions) != 3 {
		t.Errorf("Expected 3 intervals, got %d", len(mapping.Additions))
	}
	if mapping.Additions[0].GenStart != 10 {
		t.Errorf("First interval GenStart = %v, want 10", mapping.Additions[0].GenStart)
	}
	if mapping.Additions[1].GenStart != 22 {
		t.Errorf("Second interval GenStart = %v, want 22", mapping.Additions[1].GenStart)
	}
	if mapping.Additions[2].GenStart != 30 {
		t.Errorf("Third interval GenStart = %v, want 30", mapping.Additions[2].GenStart)
	}
}

func TestHybridMapping_OverlapDetection(t *testing.T) {
	t.Run("valid_non_overlapping", func(t *testing.T) {
		mapping := NewHybridMapping("source.go")

		err := mapping.AddInterval(Interval{
			GenStart:  10,
			GenEnd:    20,
			SourcePos: Position{File: "source.go", Line: 5},
		})
		if err != nil {
			t.Errorf("AddInterval() error = %v", err)
		}

		err = mapping.AddInterval(Interval{
			GenStart:  30,
			GenEnd:    40,
			SourcePos: Position{File: "source.go", Line: 15},
		})
		if err != nil {
			t.Errorf("AddInterval() error = %v", err)
		}
	})

	t.Run("overlapping_intervals", func(t *testing.T) {
		mapping := NewHybridMapping("source.go")

		mapping.AddInterval(Interval{
			GenStart:  10,
			GenEnd:    20,
			SourcePos: Position{File: "source.go", Line: 5},
		})

		err := mapping.AddInterval(Interval{
			GenStart:  15,
			GenEnd:    25,
			SourcePos: Position{File: "source.go", Line: 10},
		})

		if err == nil {
			t.Errorf("AddInterval() expected error for overlapping intervals, got nil")
		}
	})
}

func TestHybridMapping_FindInterval(t *testing.T) {
	mapping := NewHybridMapping("source.go")

	mapping.AddInterval(Interval{
		GenStart:  30,
		GenEnd:    40,
		SourcePos: Position{File: "source.go", Line: 15},
	})

	mapping.AddInterval(Interval{
		GenStart:  10,
		GenEnd:    20,
		SourcePos: Position{File: "source.go", Line: 5},
	})

	tests := []struct {
		name     string
		line     int
		col      int
		wantLine int
	}{
		{"inside_first_interval", 15, 5, 5},
		{"inside_second_interval", 35, 5, 15},
		{"before_all_intervals", 5, 5, 0},
		{"after_all_intervals", 50, 5, 0},
		{"between_intervals", 25, 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interval := mapping.FindInterval(tt.line, tt.col)
			if tt.wantLine == 0 {
				if interval != nil {
					t.Errorf("FindInterval() = %v, want nil", interval)
				}
			} else {
				if interval == nil {
					t.Errorf("FindInterval() = nil, want interval with SourcePos.Line %d", tt.wantLine)
				} else if interval.SourcePos.Line != tt.wantLine {
					t.Errorf("FindInterval() SourcePos.Line = %v, want %d", interval.SourcePos.Line, tt.wantLine)
				}
			}
		})
	}
}

func TestHybridMapping_FindPointMapping(t *testing.T) {
	mapping := NewHybridMapping("source.go")

	mapping.AddPointMapping(PointMapping{
		GenLine:   20,
		GenColumn: 10,
		Orig:      Position{File: "source.go", Line: 10, Column: 15},
	})

	mapping.AddPointMapping(PointMapping{
		GenLine:   10,
		GenColumn: 5,
		Orig:      Position{File: "source.go", Line: 5, Column: 10},
	})

	tests := []struct {
		name       string
		line       int
		col        int
		wantFound  bool
		wantLine   int
		wantColumn int
	}{
		{"exact_match", 10, 5, true, 5, 10},
		{"different_line", 15, 5, false, 0, 0},
		{"different_column", 10, 10, false, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := mapping.FindPointMapping(tt.line, tt.col)
			if tt.wantFound {
				if mapping == nil {
					t.Errorf("FindPointMapping() = nil, want mapping")
				} else if mapping.Orig.Line != tt.wantLine || mapping.Orig.Column != tt.wantColumn {
					t.Errorf("FindPointMapping() = %+v, want Line %d Column %d", mapping.Orig, tt.wantLine, tt.wantColumn)
				}
			} else {
				if mapping != nil {
					t.Errorf("FindPointMapping() = %+v, want nil", mapping)
				}
			}
		})
	}
}

func TestHybridMapping_GetTotalAdditionsBefore(t *testing.T) {
	mapping := NewHybridMapping("source.go")

	mapping.AddInterval(Interval{
		GenStart:  30,
		GenEnd:    35,
		SourcePos: Position{File: "source.go", Line: 15},
	})

	mapping.AddInterval(Interval{
		GenStart:  10,
		GenEnd:    15,
		SourcePos: Position{File: "source.go", Line: 5},
	})

	mapping.AddInterval(Interval{
		GenStart:  20,
		GenEnd:    25,
		SourcePos: Position{File: "source.go", Line: 10},
	})

	tests := []struct {
		name string
		line int
		want int
	}{
		{"before_all_additions", 5, 0},
		{"after_first_addition", 17, 6},
		{"after_second_addition", 27, 12},
		{"after_third_addition", 40, 18},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapping.GetTotalAdditionsBefore(tt.line); got != tt.want {
				t.Errorf("GetTotalAdditionsBefore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHybridMapping_MapError(t *testing.T) {
	tests := []struct {
		name    string
		mapping *HybridMapping
		line    int
		col     int
		wantPos Position
		wantErr bool
	}{
		{
			name: "error in predicate with point mapping",
			mapping: func() *HybridMapping {
				h := NewHybridMapping("source.go")
				h.AddInterval(Interval{
					GenStart:  10,
					GenEnd:    15,
					SourcePos: Position{File: "source.go", Line: 5, Column: 10},
				})
				h.AddPointMapping(PointMapping{
					GenLine:   12,
					GenColumn: 5,
					Orig:      Position{File: "source.go", Line: 5, Column: 15},
				})
				return h
			}(),
			line:    12,
			col:     5,
			wantPos: Position{File: "source.go", Line: 5, Column: 15},
			wantErr: false,
		},
		{
			name: "error in addition with predicate span",
			mapping: func() *HybridMapping {
				h := NewHybridMapping("source.go")
				h.AddInterval(Interval{
					GenStart:  10,
					GenEnd:    15,
					SourcePos: Position{File: "source.go", Line: 5, Column: 10},
					PredicateSpan: &Position{
						File:   "source.go",
						Line:   5,
						Column: 20,
					},
				})
				return h
			}(),
			line:    12,
			col:     5,
			wantPos: Position{File: "source.go", Line: 5, Column: 20},
			wantErr: false,
		},
		{
			name: "error in addition without point mapping or predicate span",
			mapping: func() *HybridMapping {
				h := NewHybridMapping("source.go")
				h.AddInterval(Interval{
					GenStart:  10,
					GenEnd:    15,
					SourcePos: Position{File: "source.go", Line: 5, Column: 10},
				})
				return h
			}(),
			line:    12,
			col:     5,
			wantPos: Position{File: "source.go", Line: 5, Column: 10},
			wantErr: false,
		},
		{
			name: "error in original code before additions",
			mapping: func() *HybridMapping {
				h := NewHybridMapping("source.go")
				h.AddInterval(Interval{
					GenStart:  10,
					GenEnd:    15,
					SourcePos: Position{File: "source.go", Line: 5, Column: 10},
				})
				return h
			}(),
			line:    5,
			col:     10,
			wantPos: Position{File: "source.go", Line: 5, Column: 10},
			wantErr: false,
		},
		{
			name: "error in original code after additions",
			mapping: func() *HybridMapping {
				h := NewHybridMapping("source.go")
				h.AddInterval(Interval{
					GenStart:  10,
					GenEnd:    15,
					SourcePos: Position{File: "source.go", Line: 5, Column: 10},
				})
				return h
			}(),
			line:    20,
			col:     5,
			wantPos: Position{File: "source.go", Line: 14, Column: 5}, // 20 - 6 (lines in addition: 10-15 inclusive)
			wantErr: false,
		},
		{
			name: "invalid position - negative line",
			mapping: func() *HybridMapping {
				return NewHybridMapping("source.go")
			}(),
			line:    -1,
			col:     5,
			wantErr: true,
		},
		{
			name: "invalid position - zero column",
			mapping: func() *HybridMapping {
				return NewHybridMapping("source.go")
			}(),
			line:    10,
			col:     0,
			wantErr: true,
		},
		{
			name: "empty mapping - error in original code",
			mapping: func() *HybridMapping {
				return NewHybridMapping("source.go")
			}(),
			line:    10,
			col:     5,
			wantPos: Position{File: "source.go", Line: 10, Column: 5},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPos, err := tt.mapping.MapError(tt.line, tt.col)
			if tt.wantErr {
				if err == nil {
					t.Errorf("MapError() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("MapError() unexpected error: %v", err)
				return
			}
			if gotPos != tt.wantPos {
				t.Errorf("MapError() = %+v, want %+v", gotPos, tt.wantPos)
			}
		})
	}
}

func TestLineColToOffset(t *testing.T) {
	tests := []struct {
		name    string
		lines   []string
		line    int
		col     int
		want    int
		wantErr bool
	}{
		{
			name:  "simple case",
			lines: []string{"line1", "line2", "line3"},
			line:  2,
			col:   3,
			want:  8, // "line1" (5) + "\n" (1) + 2 (0-based col from col 3)
		},
		{
			name:  "first line",
			lines: []string{"line1", "line2", "line3"},
			line:  1,
			col:   1,
			want:  0,
		},
		{
			name:  "last line",
			lines: []string{"line1", "line2", "line3"},
			line:  3,
			col:   5,
			want:  16, // "line1\n" (6) + "line2\n" (6) + 4 (0-based col from col 5)
		},
		{
			name:  "empty lines",
			lines: []string{"", "", ""},
			line:  2,
			col:   1,
			want:  1, // "\n" (1) + 0 (0-based col)
		},
		{
			name:    "invalid line - negative",
			lines:   []string{"line1", "line2"},
			line:    -1,
			col:     1,
			wantErr: true,
		},
		{
			name:    "invalid line - zero",
			lines:   []string{"line1", "line2"},
			line:    0,
			col:     1,
			wantErr: true,
		},
		{
			name:    "invalid line - exceeds file length",
			lines:   []string{"line1", "line2"},
			line:    5,
			col:     1,
			wantErr: true,
		},
		{
			name:    "invalid column - negative",
			lines:   []string{"line1", "line2"},
			line:    1,
			col:     -1,
			wantErr: true,
		},
		{
			name:    "invalid column - zero",
			lines:   []string{"line1", "line2"},
			line:    1,
			col:     0,
			wantErr: true,
		},
		{
			name:    "invalid column - exceeds line length",
			lines:   []string{"line1", "line2"},
			line:    1,
			col:     10,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LineColToOffset(tt.lines, tt.line, tt.col)
			if tt.wantErr {
				if err == nil {
					t.Errorf("LineColToOffset() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("LineColToOffset() unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("LineColToOffset() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestOffsetToLineCol(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		offset   int
		wantLine int
		wantCol  int
		wantErr  bool
	}{
		{
			name:     "simple case",
			content:  []byte("line1\nline2\nline3"),
			offset:   8,
			wantLine: 2,
			wantCol:  3,
		},
		{
			name:     "start of file",
			content:  []byte("line1\nline2\nline3"),
			offset:   0,
			wantLine: 1,
			wantCol:  1,
		},
		{
			name:     "end of file",
			content:  []byte("line1\nline2\nline3"),
			offset:   16,
			wantLine: 3,
			wantCol:  5,
		},
		{
			name:     "after newline",
			content:  []byte("line1\nline2\nline3"),
			offset:   6,
			wantLine: 2,
			wantCol:  1,
		},
		{
			name:     "empty content",
			content:  []byte(""),
			offset:   0,
			wantLine: 1,
			wantCol:  1,
		},
		{
			name:     "single line",
			content:  []byte("single line"),
			offset:   7,
			wantLine: 1,
			wantCol:  8,
		},
		{
			name:     "multiple newlines",
			content:  []byte("\n\n\n"),
			offset:   2,
			wantLine: 3,
			wantCol:  1,
		},
		{
			name:    "invalid offset - negative",
			content: []byte("line1\nline2"),
			offset:  -1,
			wantErr: true,
		},
		{
			name:    "invalid offset - exceeds content length",
			content: []byte("line1\nline2"),
			offset:  100,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLine, gotCol, err := OffsetToLineCol(tt.content, tt.offset)
			if tt.wantErr {
				if err == nil {
					t.Errorf("OffsetToLineCol() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("OffsetToLineCol() unexpected error: %v", err)
				return
			}
			if gotLine != tt.wantLine || gotCol != tt.wantCol {
				t.Errorf("OffsetToLineCol() = (%d, %d), want (%d, %d)", gotLine, gotCol, tt.wantLine, tt.wantCol)
			}
		})
	}
}

func TestLineColOffsetRoundtrip(t *testing.T) {
	content := []byte("line1\nline2\nline3\nline4")
	lines := []string{"line1", "line2", "line3", "line4"}

	testCases := []struct {
		line int
		col  int
	}{
		{1, 1}, {1, 5}, {2, 1}, {2, 5}, {3, 1}, {3, 5}, {4, 1}, {4, 5},
	}

	for _, tc := range testCases {
		t.Run("roundtrip", func(t *testing.T) {
			// Convert line:col to offset
			offset, err := LineColToOffset(lines, tc.line, tc.col)
			if err != nil {
				t.Fatalf("LineColToOffset() error: %v", err)
			}

			// Convert offset back to line:col
			gotLine, gotCol, err := OffsetToLineCol(content, offset)
			if err != nil {
				t.Fatalf("OffsetToLineCol() error: %v", err)
			}

			// Verify roundtrip
			if gotLine != tc.line || gotCol != tc.col {
				t.Errorf("Roundtrip failed: (%d, %d) -> offset %d -> (%d, %d)",
					tc.line, tc.col, offset, gotLine, gotCol)
			}
		})
	}
}

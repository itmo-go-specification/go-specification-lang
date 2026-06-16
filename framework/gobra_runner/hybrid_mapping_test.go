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
				GenStart:  100, // Offset in generated code
				GenEnd:    200,
				SourcePos: SourcePosition{File: "test.go", Offset: 50}, // Offset in original code
			},
			want: true,
		},
		{
			name: "valid interval with same start and end",
			interval: Interval{
				GenStart:  100, // Offset in generated code
				GenEnd:    100,
				SourcePos: SourcePosition{File: "test.go", Offset: 50}, // Offset in original code
			},
			want: true,
		},
		{
			name: "invalid - start > end",
			interval: Interval{
				GenStart:  200, // Offset in generated code
				GenEnd:    100,
				SourcePos: SourcePosition{File: "test.go", Offset: 50}, // Offset in original code
			},
			want: false,
		},
		{
			name: "invalid - negative start",
			interval: Interval{
				GenStart:  -1,
				GenEnd:    10,
				SourcePos: SourcePosition{File: "test.go", Offset: 5},
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
		offset   int
		want     bool
	}{
		{
			name: "position inside interval",
			interval: Interval{
				GenStart: 10,
				GenEnd:   20,
			},
			offset: 15,
			want:   true,
		},
		{
			name: "position before interval",
			interval: Interval{
				GenStart: 10,
				GenEnd:   20,
			},
			offset: 5,
			want:   false,
		},
		{
			name: "position after interval",
			interval: Interval{
				GenStart: 10,
				GenEnd:   20,
			},
			offset: 25,
			want:   false,
		},
		{
			name: "position on start boundary",
			interval: Interval{
				GenStart: 10,
				GenEnd:   20,
			},
			offset: 10,
			want:   true,
		},
		{
			name: "position on end boundary",
			interval: Interval{
				GenStart: 10,
				GenEnd:   20,
			},
			offset: 20,
			want:   true,
		},
		{
			name: "position after end boundary",
			interval: Interval{
				GenStart: 10,
				GenEnd:   20,
			},
			offset: 25,
			want:   false,
		},
		{
			name: "position on end boundary",
			interval: Interval{
				GenStart: 10,
				GenEnd:   20,
			},
			offset: 20,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interval.Contains(tt.offset); got != tt.want {
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
	mapping := NewHybridMapping("")

	if len(mapping.Replacements) != 0 {
		t.Errorf("NewHybridMapping() Additions length = %v, want 0", len(mapping.Replacements))
	}
	if len(mapping.PointMappings) != 0 {
		t.Errorf("NewHybridMapping() PointMappings length = %v, want 0", len(mapping.PointMappings))
	}
}

func TestHybridMapping_AddInterval(t *testing.T) {
	mapping := NewHybridMapping("")

	interval := Interval{
		GenStart:  10,
		GenEnd:    20,
		SourcePos: SourcePosition{File: "source.go", Offset: 5},
	}

	err := mapping.AddInterval(interval)
	if err != nil {
		t.Errorf("AddInterval() error = %v", err)
	}

	if len(mapping.Replacements) != 1 {
		t.Errorf("AddInterval() Additions length = %v, want 1", len(mapping.Replacements))
	}
}

func TestHybridMapping_AddInterval_Invalid(t *testing.T) {
	mapping := NewHybridMapping("")

	invalidInterval := Interval{
		GenStart:  20,
		GenEnd:    10,
		SourcePos: SourcePosition{File: "source.go", Offset: 5},
	}

	err := mapping.AddInterval(invalidInterval)
	if err == nil {
		t.Errorf("AddInterval() expected error for invalid interval, got nil")
	}
}

func TestHybridMapping_AddInterval_Overlap(t *testing.T) {
	mapping := NewHybridMapping("")

	interval1 := Interval{
		GenStart:  10,
		GenEnd:    20,
		SourcePos: SourcePosition{File: "source.go", Offset: 5},
	}

	interval2 := Interval{
		GenStart:  15,
		GenEnd:    25,
		SourcePos: SourcePosition{File: "source.go", Offset: 10},
	}

	mapping.AddInterval(interval1)
	err := mapping.AddInterval(interval2)

	if err == nil {
		t.Errorf("AddInterval() expected error for overlapping intervals, got nil")
	}
}

func TestHybridMapping_AutoSort(t *testing.T) {
	mapping := NewHybridMapping("")

	mapping.AddInterval(Interval{
		GenStart:  30,
		GenEnd:    40,
		SourcePos: SourcePosition{File: "source.go", Offset: 15},
	})

	mapping.AddInterval(Interval{
		GenStart:  10,
		GenEnd:    20,
		SourcePos: SourcePosition{File: "source.go", Offset: 5},
	})

	mapping.AddInterval(Interval{
		GenStart:  22,
		GenEnd:    28,
		SourcePos: SourcePosition{File: "source.go", Offset: 10},
	})

	// Intervals should be automatically sorted
	if len(mapping.Replacements) != 3 {
		t.Errorf("Expected 3 intervals, got %d", len(mapping.Replacements))
	}
	if mapping.Replacements[0].GenStart != 10 {
		t.Errorf("First interval GenStart = %v, want 10", mapping.Replacements[0].GenStart)
	}
	if mapping.Replacements[1].GenStart != 22 {
		t.Errorf("Second interval GenStart = %v, want 22", mapping.Replacements[1].GenStart)
	}
	if mapping.Replacements[2].GenStart != 30 {
		t.Errorf("Third interval GenStart = %v, want 30", mapping.Replacements[2].GenStart)
	}
}

func TestHybridMapping_OverlapDetection(t *testing.T) {
	t.Run("valid_non_overlapping", func(t *testing.T) {
		mapping := NewHybridMapping("")

		err := mapping.AddInterval(Interval{
			GenStart:  10,
			GenEnd:    20,
			SourcePos: SourcePosition{File: "source.go", Offset: 5},
		})
		if err != nil {
			t.Errorf("AddInterval() error = %v", err)
		}

		err = mapping.AddInterval(Interval{
			GenStart:  30,
			GenEnd:    40,
			SourcePos: SourcePosition{File: "source.go", Offset: 15},
		})
		if err != nil {
			t.Errorf("AddInterval() error = %v", err)
		}
	})

	t.Run("overlapping_intervals", func(t *testing.T) {
		mapping := NewHybridMapping("")

		mapping.AddInterval(Interval{
			GenStart:  10,
			GenEnd:    20,
			SourcePos: SourcePosition{File: "source.go", Offset: 5},
		})

		err := mapping.AddInterval(Interval{
			GenStart:  15,
			GenEnd:    25,
			SourcePos: SourcePosition{File: "source.go", Offset: 10},
		})

		if err == nil {
			t.Errorf("AddInterval() expected error for overlapping intervals, got nil")
		}
	})
}

func TestHybridMapping_FindInterval(t *testing.T) {
	mapping := NewHybridMapping("")

	mapping.AddInterval(Interval{
		GenStart:  30,
		GenEnd:    40,
		SourcePos: SourcePosition{File: "source.go", Offset: 15},
	})

	mapping.AddInterval(Interval{
		GenStart:  10,
		GenEnd:    20,
		SourcePos: SourcePosition{File: "source.go", Offset: 5},
	})

	tests := []struct {
		name     string
		offset   int
		wantLine int
	}{
		{"inside_first_interval", 15, 5},
		{"inside_second_interval", 35, 15},
		{"before_all_intervals", 5, 0},
		{"after_all_intervals", 50, 0},
		{"between_intervals", 25, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interval := mapping.FindInterval(tt.offset)
			if tt.wantLine == 0 {
				if interval != nil {
					t.Errorf("FindInterval() = %v, want nil", interval)
				}
			} else {
				if interval == nil {
					t.Errorf("FindInterval() = nil, want interval with SourcePos.Line %d", tt.wantLine)
				} else if interval.SourcePos.Offset != tt.wantLine {
					t.Errorf("FindInterval() SourcePos.Line = %v, want %d", interval.SourcePos.Offset, tt.wantLine)
				}
			}
		})
	}
}

func TestHybridMapping_FindPointMapping(t *testing.T) {
	mapping := NewHybridMapping("")

	mapping.AddPointMapping(PointMapping{
		GenOffset: 200,                                            // Offset in generated code
		Orig:      SourcePosition{File: "source.go", Offset: 100}, // Offset in original code
	})

	mapping.AddPointMapping(PointMapping{
		GenOffset: 100,                                           // Offset in generated code
		Orig:      SourcePosition{File: "source.go", Offset: 50}, // Offset in original code
	})

	tests := []struct {
		name       string
		offset     int
		wantFound  bool
		wantOffset int
	}{
		{"exact_match", 100, true, 50},
		{"different_offset", 150, false, 0},
		{"before_mapping", 50, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pointMapping := mapping.FindPointMapping(tt.offset)
			if tt.wantFound {
				if pointMapping == nil {
					t.Errorf("FindPointMapping() = nil, want mapping")
				} else if pointMapping.Orig.Offset != tt.wantOffset {
					t.Errorf("FindPointMapping() = %+v, want Offset %d", pointMapping.Orig, tt.wantOffset)
				}
			} else {
				if pointMapping != nil {
					t.Errorf("FindPointMapping() = %+v, want nil", mapping)
				}
			}
		})
	}
}

func TestHybridMapping_GetTotalAdditionsBefore(t *testing.T) {
	mapping := NewHybridMapping("")

	mapping.AddInterval(Interval{
		GenStart:  30,
		GenEnd:    35,
		SourcePos: SourcePosition{File: "source.go", Offset: 15},
	})

	mapping.AddInterval(Interval{
		GenStart:  10,
		GenEnd:    15,
		SourcePos: SourcePosition{File: "source.go", Offset: 5},
	})

	mapping.AddInterval(Interval{
		GenStart:  20,
		GenEnd:    25,
		SourcePos: SourcePosition{File: "source.go", Offset: 10},
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
		offset  int
		wantPos SourcePosition
		wantErr bool
	}{
		{
			name: "error in predicate with point mapping",
			mapping: func() *HybridMapping {
				h := NewHybridMapping("")
				h.AddInterval(Interval{
					GenStart:  10,
					GenEnd:    15,
					SourcePos: SourcePosition{File: "source.go", Offset: 5},
				})
				h.AddPointMapping(PointMapping{
					GenOffset: 12,
					Orig:      SourcePosition{File: "source.go", Offset: 5},
				})
				return h
			}(),
			offset:  12,
			wantPos: SourcePosition{File: "source.go", Offset: 5},
		},
		{
			name: "error in addition with predicate span",
			mapping: func() *HybridMapping {
				h := NewHybridMapping("")
				h.AddInterval(Interval{
					GenStart:  10,
					GenEnd:    15,
					SourcePos: SourcePosition{File: "source.go", Offset: 5},
					PredicateSpan: &SourcePosition{
						File:   "source.go",
						Offset: 20,
					},
				})
				return h
			}(),
			offset:  12,
			wantPos: SourcePosition{File: "source.go", Offset: 20},
		},
		{
			name: "error in addition without point mapping or predicate span",
			mapping: func() *HybridMapping {
				h := NewHybridMapping("")
				h.AddInterval(Interval{
					GenStart:  10,
					GenEnd:    15,
					SourcePos: SourcePosition{File: "source.go", Offset: 5},
				})
				return h
			}(),
			offset:  12,
			wantPos: SourcePosition{File: "source.go", Offset: 5},
		},
		{
			name: "error in original code before additions",
			mapping: func() *HybridMapping {
				h := NewHybridMapping("")
				h.AddInterval(Interval{
					GenStart:  10,
					GenEnd:    15,
					SourcePos: SourcePosition{File: "source.go", Offset: 5},
				})
				return h
			}(),
			offset:  5,
			wantPos: SourcePosition{Offset: 5},
		},
		{
			name: "error in original code after additions",
			mapping: func() *HybridMapping {
				h := NewHybridMapping("")
				h.AddInterval(Interval{
					GenStart:  10,
					GenEnd:    15,
					SourcePos: SourcePosition{File: "source.go", Offset: 5},
				})
				return h
			}(),
			offset:  20,
			wantPos: SourcePosition{Offset: 14}, // 20 - 6 (lines in addition: 10-15 inclusive)
		},
		{
			name: "invalid position - negative line",
			mapping: func() *HybridMapping {
				return NewHybridMapping("")
			}(),
			offset:  -1,
			wantErr: true,
		},
		{
			name: "invalid position - negative offset",
			mapping: func() *HybridMapping {
				return NewHybridMapping("")
			}(),
			offset:  -1,
			wantErr: true,
		},
		{
			name: "empty mapping - error in original code",
			mapping: func() *HybridMapping {
				return NewHybridMapping("")
			}(),
			offset:  10,
			wantPos: SourcePosition{Offset: 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPos, err := tt.mapping.MapError(tt.offset)
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
		content []byte
		line    int
		col     int
		want    int
		wantErr bool
	}{
		{
			name:    "simple case",
			content: []byte("line1\nline2\nline3"),
			line:    2,
			col:     3,
			want:    8, // "line1" (5) + "\n" (1) + 2 (0-based col from col 3)
		},
		{
			name:    "first line",
			content: []byte("line1\nline2\nline3"),
			line:    1,
			col:     1,
			want:    0,
		},
		{
			name:    "last line",
			content: []byte("line1\nline2\nline3"),
			line:    3,
			col:     5,
			want:    16, // "line1\n" (6) + "line2\n" (6) + 4 (0-based col from col 5)
		},
		{
			name:    "empty lines",
			content: []byte("\n\n"),
			line:    2,
			col:     1,
			want:    1, // "\n" (1) + 0 (0-based col)
		},
		{
			name:    "invalid line - negative",
			content: []byte("line1\nline2\nline3"),
			line:    -1,
			wantErr: true,
		},
		{
			name:    "invalid line - zero",
			content: []byte("line1\nline2\nline3"),
			line:    0,
			wantErr: true,
		},
		{
			name:    "invalid line - exceeds file length",
			content: []byte("line1\nline2\nline3"),
			line:    5,
			wantErr: true,
		},
		{
			name:    "invalid column - negative",
			content: []byte("line1\nline2\nline3"),
			line:    1,
			col:     -1,
			wantErr: true,
		},
		{
			name:    "invalid column - zero",
			content: []byte("line1\nline2\nline3"),
			line:    1,
			wantErr: true,
		},
		{
			name:    "invalid column - exceeds line length",
			content: []byte("line1\nline2\nline3"),
			line:    1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LineColToOffset(tt.content, tt.line, tt.col)
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

	testCases := []struct {
		line int
		col  int
	}{
		{1, 1}, {1, 5}, {2, 1}, {2, 5}, {3, 1}, {3, 5}, {4, 1}, {4, 5},
	}

	for _, tc := range testCases {
		t.Run("roundtrip", func(t *testing.T) {
			// Convert line:col to offset
			offset, err := LineColToOffset(content, tc.line, tc.col)
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

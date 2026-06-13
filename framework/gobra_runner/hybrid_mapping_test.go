package gobra_runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAdditionInterval_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		interval AdditionInterval
		want     bool
	}{
		{
			name: "valid interval",
			interval: AdditionInterval{
				GenStart:  10,
				GenEnd:    20,
				SourcePos: SourcePos{File: "test.go", Line: 5},
				Kind:      AdditionRequires,
			},
			want: true,
		},
		{
			name: "valid interval with columns",
			interval: AdditionInterval{
				GenStart:    10,
				GenEnd:      10,
				GenStartCol: 5,
				GenEndCol:   15,
				SourcePos:   SourcePos{File: "test.go", Line: 5},
				Kind:        AdditionRequires,
			},
			want: true,
		},
		{
			name: "invalid - start > end",
			interval: AdditionInterval{
				GenStart:  20,
				GenEnd:    10,
				SourcePos: SourcePos{File: "test.go", Line: 5},
				Kind:      AdditionRequires,
			},
			want: false,
		},
		{
			name: "invalid - negative start",
			interval: AdditionInterval{
				GenStart:  -1,
				GenEnd:    10,
				SourcePos: SourcePos{File: "test.go", Line: 5},
				Kind:      AdditionRequires,
			},
			want: false,
		},
		{
			name: "invalid - column start > column end",
			interval: AdditionInterval{
				GenStart:    10,
				GenEnd:      10,
				GenStartCol: 15,
				GenEndCol:   5,
				SourcePos:   SourcePos{File: "test.go", Line: 5},
				Kind:        AdditionRequires,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interval.IsValid(); got != tt.want {
				t.Errorf("AdditionInterval.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAdditionInterval_Contains(t *testing.T) {
	tests := []struct {
		name     string
		interval AdditionInterval
		line     int
		col      int
		want     bool
	}{
		{
			name: "position inside interval",
			interval: AdditionInterval{
				GenStart: 10,
				GenEnd:   20,
			},
			line: 15,
			col:  5,
			want: true,
		},
		{
			name: "position before interval",
			interval: AdditionInterval{
				GenStart: 10,
				GenEnd:   20,
			},
			line: 5,
			col:  5,
			want: false,
		},
		{
			name: "position after interval",
			interval: AdditionInterval{
				GenStart: 10,
				GenEnd:   20,
			},
			line: 25,
			col:  5,
			want: false,
		},
		{
			name: "position on start line before column",
			interval: AdditionInterval{
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
			interval: AdditionInterval{
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
			interval: AdditionInterval{
				GenStart:  10,
				GenEnd:    20,
				GenEndCol: 15,
			},
			line: 20,
			col:  20,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interval.Contains(tt.line, tt.col); got != tt.want {
				t.Errorf("AdditionInterval.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAdditionInterval_Length(t *testing.T) {
	tests := []struct {
		name     string
		interval AdditionInterval
		want     int
	}{
		{
			name: "single line",
			interval: AdditionInterval{
				GenStart: 10,
				GenEnd:   10,
			},
			want: 1,
		},
		{
			name: "multiple lines",
			interval: AdditionInterval{
				GenStart: 10,
				GenEnd:   20,
			},
			want: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interval.Length(); got != tt.want {
				t.Errorf("AdditionInterval.Length() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHybridMapping_NewHybridMapping(t *testing.T) {
	mapping := NewHybridMapping("gen.gobra", "source.go")

	if mapping.GenFile != "gen.gobra" {
		t.Errorf("Expected GenFile to be 'gen.gobra', got '%s'", mapping.GenFile)
	}
	if mapping.SourceFile != "source.go" {
		t.Errorf("Expected SourceFile to be 'source.go', got '%s'", mapping.SourceFile)
	}
	if len(mapping.Additions) != 0 {
		t.Errorf("Expected empty Additions, got %d items", len(mapping.Additions))
	}
	if len(mapping.PointMappings) != 0 {
		t.Errorf("Expected empty PointMappings, got %d items", len(mapping.PointMappings))
	}
}

func TestHybridMapping_AddInterval(t *testing.T) {
	mapping := NewHybridMapping("gen.gobra", "source.go")

	interval := AdditionInterval{
		GenStart:  10,
		GenEnd:    20,
		SourcePos: SourcePos{File: "source.go", Line: 5},
		Kind:      AdditionRequires,
	}

	err := mapping.AddInterval(interval)
	if err != nil {
		t.Errorf("AddInterval() failed: %v", err)
	}

	if len(mapping.Additions) != 1 {
		t.Errorf("Expected 1 addition, got %d", len(mapping.Additions))
	}
}

func TestHybridMapping_AddInterval_Invalid(t *testing.T) {
	mapping := NewHybridMapping("gen.gobra", "source.go")

	interval := AdditionInterval{
		GenStart:  20,
		GenEnd:    10, // Invalid: start > end
		SourcePos: SourcePos{File: "source.go", Line: 5},
		Kind:      AdditionRequires,
	}

	err := mapping.AddInterval(interval)
	if err == nil {
		t.Error("AddInterval() should fail for invalid interval")
	}
}

func TestHybridMapping_AddInterval_Overlap(t *testing.T) {
	mapping := NewHybridMapping("gen.gobra", "source.go")

	interval1 := AdditionInterval{
		GenStart:  10,
		GenEnd:    20,
		SourcePos: SourcePos{File: "source.go", Line: 5},
		Kind:      AdditionRequires,
	}

	interval2 := AdditionInterval{
		GenStart:  15,
		GenEnd:    25, // Overlaps with interval1
		SourcePos: SourcePos{File: "source.go", Line: 10},
		Kind:      AdditionEnsures,
	}

	mapping.AddInterval(interval1)
	err := mapping.AddInterval(interval2)

	if err == nil {
		t.Error("AddInterval() should fail for overlapping intervals")
	}
}

func TestHybridMapping_Sort(t *testing.T) {
	mapping := NewHybridMapping("gen.gobra", "source.go")

	// Add intervals in random order
	mapping.AddInterval(AdditionInterval{GenStart: 30, GenEnd: 40, SourcePos: SourcePos{Line: 15}, Kind: AdditionAssert})
	mapping.AddInterval(AdditionInterval{GenStart: 10, GenEnd: 20, SourcePos: SourcePos{Line: 5}, Kind: AdditionRequires})
	mapping.AddInterval(AdditionInterval{GenStart: 20, GenEnd: 30, SourcePos: SourcePos{Line: 10}, Kind: AdditionEnsures})

	mapping.Sort()

	// Check if sorted
	for i := 1; i < len(mapping.Additions); i++ {
		if mapping.Additions[i-1].GenStart > mapping.Additions[i].GenStart {
			t.Errorf("Additions not sorted: interval %d starts before interval %d", i-1, i)
		}
	}
}

func TestHybridMapping_Validate(t *testing.T) {
	t.Run("valid mapping", func(t *testing.T) {
		mapping := NewHybridMapping("gen.gobra", "source.go")
		mapping.AddInterval(AdditionInterval{GenStart: 10, GenEnd: 20, SourcePos: SourcePos{Line: 5}, Kind: AdditionRequires})
		mapping.AddInterval(AdditionInterval{GenStart: 30, GenEnd: 40, SourcePos: SourcePos{Line: 10}, Kind: AdditionEnsures})
		mapping.Sort()

		err := mapping.Validate()
		if err != nil {
			t.Errorf("Validate() failed for valid mapping: %v", err)
		}
	})

	t.Run("invalid - not sorted", func(t *testing.T) {
		mapping := NewHybridMapping("gen.gobra", "source.go")
		mapping.AddInterval(AdditionInterval{GenStart: 30, GenEnd: 40, SourcePos: SourcePos{Line: 10}, Kind: AdditionEnsures})
		mapping.AddInterval(AdditionInterval{GenStart: 10, GenEnd: 20, SourcePos: SourcePos{Line: 5}, Kind: AdditionRequires})
		// Don't sort

		err := mapping.Validate()
		if err == nil {
			t.Error("Validate() should fail for unsorted mapping")
		}
	})

	t.Run("invalid - overlapping intervals", func(t *testing.T) {
		mapping := NewHybridMapping("gen.gobra", "source.go")
		// Manually create overlapping intervals (bypassing AddInterval check)
		mapping.Additions = []AdditionInterval{
			{GenStart: 10, GenEnd: 20, SourcePos: SourcePos{Line: 5}, Kind: AdditionRequires},
			{GenStart: 15, GenEnd: 25, SourcePos: SourcePos{Line: 10}, Kind: AdditionEnsures},
		}

		err := mapping.Validate()
		if err == nil {
			t.Error("Validate() should fail for overlapping intervals")
		}
	})
}

func TestHybridMapping_FindInterval(t *testing.T) {
	mapping := NewHybridMapping("gen.gobra", "source.go")
	mapping.AddInterval(AdditionInterval{GenStart: 10, GenEnd: 20, SourcePos: SourcePos{Line: 5}, Kind: AdditionRequires})
	mapping.AddInterval(AdditionInterval{GenStart: 30, GenEnd: 40, SourcePos: SourcePos{Line: 10}, Kind: AdditionEnsures})
	mapping.Sort()

	tests := []struct {
		name string
		line int
		col  int
		want bool
	}{
		{"inside first interval", 15, 5, true},
		{"inside second interval", 35, 5, true},
		{"before all intervals", 5, 5, false},
		{"after all intervals", 50, 5, false},
		{"between intervals", 25, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interval := mapping.FindInterval(tt.line, tt.col)
			found := interval != nil
			if found != tt.want {
				t.Errorf("FindInterval() = %v, want %v", found, tt.want)
			}
		})
	}
}

func TestHybridMapping_FindPointMapping(t *testing.T) {
	mapping := NewHybridMapping("gen.gobra", "source.go")
	mapping.AddPointMapping(MappingEntry{
		GenLine:   10,
		GenColumn: 5,
		Orig:      SourcePos{File: "source.go", Line: 5, Column: 10},
		Kind:      "predicate",
	})
	mapping.Sort()

	tests := []struct {
		name string
		line int
		col  int
		want bool
	}{
		{"exact match", 10, 5, true},
		{"different line", 11, 5, false},
		{"different column", 10, 6, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := mapping.FindPointMapping(tt.line, tt.col)
			found := mapping != nil
			if found != tt.want {
				t.Errorf("FindPointMapping() = %v, want %v", found, tt.want)
			}
		})
	}
}

func TestHybridMapping_GetTotalAdditionsBefore(t *testing.T) {
	mapping := NewHybridMapping("gen.gobra", "source.go")
	mapping.AddInterval(AdditionInterval{GenStart: 10, GenEnd: 15, SourcePos: SourcePos{Line: 5}, Kind: AdditionRequires}) // 6 lines
	mapping.AddInterval(AdditionInterval{GenStart: 20, GenEnd: 25, SourcePos: SourcePos{Line: 10}, Kind: AdditionEnsures}) // 6 lines
	mapping.AddInterval(AdditionInterval{GenStart: 30, GenEnd: 35, SourcePos: SourcePos{Line: 15}, Kind: AdditionAssert})  // 6 lines

	tests := []struct {
		name string
		line int
		want int
	}{
		{"before all additions", 5, 0},
		{"after first addition", 18, 6},
		{"after second addition", 28, 12},
		{"after third addition", 40, 18},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapping.GetTotalAdditionsBefore(tt.line)
			if got != tt.want {
				t.Errorf("GetTotalAdditionsBefore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHybridMapping_JSON(t *testing.T) {
	original := NewHybridMapping("gen.gobra", "source.go")
	original.AddInterval(AdditionInterval{
		GenStart:    10,
		GenEnd:      20,
		GenStartCol: 5,
		GenEndCol:   15,
		SourcePos:   SourcePos{File: "source.go", Line: 5, Column: 10},
		Kind:        AdditionRequires,
		Description: "test interval",
	})
	original.AddPointMapping(MappingEntry{
		GenLine:   25,
		GenColumn: 10,
		Orig:      SourcePos{File: "source.go", Line: 8, Column: 5},
		Kind:      "predicate",
	})

	// Convert to JSON
	data, err := original.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() failed: %v", err)
	}

	// Convert back from JSON
	restored, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON() failed: %v", err)
	}

	// Verify
	if restored.GenFile != original.GenFile {
		t.Errorf("GenFile mismatch: got %v, want %v", restored.GenFile, original.GenFile)
	}
	if restored.SourceFile != original.SourceFile {
		t.Errorf("SourceFile mismatch: got %v, want %v", restored.SourceFile, original.SourceFile)
	}
	if len(restored.Additions) != len(original.Additions) {
		t.Errorf("Additions count mismatch: got %d, want %d", len(restored.Additions), len(original.Additions))
	}
	if len(restored.PointMappings) != len(original.PointMappings) {
		t.Errorf("PointMappings count mismatch: got %d, want %d", len(restored.PointMappings), len(original.PointMappings))
	}
}

func TestHybridMapping_FileOperations(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mapping.json")

	original := NewHybridMapping("gen.gobra", "source.go")
	original.AddInterval(AdditionInterval{
		GenStart:  10,
		GenEnd:    20,
		SourcePos: SourcePos{File: "source.go", Line: 5},
		Kind:      AdditionRequires,
	})

	// Save to file
	err := SaveHybridMappingToFile(original, filePath)
	if err != nil {
		t.Fatalf("SaveHybridMappingToFile() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("File was not created")
	}

	// Load from file
	loaded, err := LoadHybridMappingFromFile(filePath)
	if err != nil {
		t.Fatalf("LoadHybridMappingFromFile() failed: %v", err)
	}

	// Verify
	if loaded.GenFile != original.GenFile {
		t.Errorf("GenFile mismatch: got %v, want %v", loaded.GenFile, original.GenFile)
	}
	if len(loaded.Additions) != len(original.Additions) {
		t.Errorf("Additions count mismatch: got %d, want %d", len(loaded.Additions), len(original.Additions))
	}
}

func TestVerificationError_IsMapped(t *testing.T) {
	tests := []struct {
		name  string
		error VerificationError
		want  bool
	}{
		{
			name: "mapped error",
			error: VerificationError{
				File:   "source.go",
				Line:   10,
				Column: 5,
			},
			want: true,
		},
		{
			name: "unmapped error - no file",
			error: VerificationError{
				Line:   10,
				Column: 5,
			},
			want: false,
		},
		{
			name: "unmapped error - no line",
			error: VerificationError{
				File:   "source.go",
				Column: 5,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.error.IsMapped(); got != tt.want {
				t.Errorf("VerificationError.IsMapped() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerificationError_IsInAddedCode(t *testing.T) {
	tests := []struct {
		name  string
		error VerificationError
		want  bool
	}{
		{
			name: "error in added code via interval",
			error: VerificationError{
				File: "source.go",
				Line: 10,
				MappingInfo: &ErrorMappingInfo{
					MappingType: MappingTypeInterval,
				},
			},
			want: true,
		},
		{
			name: "error in added code via point",
			error: VerificationError{
				File: "source.go",
				Line: 10,
				MappingInfo: &ErrorMappingInfo{
					MappingType: MappingTypePoint,
				},
			},
			want: true,
		},
		{
			name: "error in original code",
			error: VerificationError{
				File: "source.go",
				Line: 10,
				MappingInfo: &ErrorMappingInfo{
					MappingType: MappingTypeFallback,
				},
			},
			want: false,
		},
		{
			name: "error without mapping info",
			error: VerificationError{
				File: "source.go",
				Line: 10,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.error.IsInAddedCode(); got != tt.want {
				t.Errorf("VerificationError.IsInAddedCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerificationError_String(t *testing.T) {
	tests := []struct {
		name     string
		error    VerificationError
		contains string
	}{
		{
			name: "mapped error",
			error: VerificationError{
				File:    "source.go",
				Line:    10,
				Column:  5,
				Kind:    "assertion failed",
				Message: "test error",
			},
			contains: "source.go:10:5",
		},
		{
			name: "unmapped error",
			error: VerificationError{
				GenLine:   15,
				GenColumn: 3,
				Kind:      "assertion failed",
				Message:   "test error",
			},
			contains: "<generated>:15:3",
		},
		{
			name: "unmapped error with gen file",
			error: VerificationError{
				GenFile:   "gen.gobra",
				GenLine:   15,
				GenColumn: 3,
				Kind:      "assertion failed",
				Message:   "test error",
			},
			contains: "gen.gobra:15:3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := tt.error.String()
			if !containsString(str, tt.contains) {
				t.Errorf("String() = %v, should contain %v", str, tt.contains)
			}
		})
	}
}

func TestVerificationError_DetailedString(t *testing.T) {
	error := VerificationError{
		File:    "source.go",
		Line:    10,
		Column:  5,
		Kind:    "assertion failed",
		Message: "test error",
		MappingInfo: &ErrorMappingInfo{
			MappingType: MappingTypeInterval,
			Confidence:  0.95,
		},
	}

	str := error.DetailedString()

	if !containsString(str, "source.go:10:5") {
		t.Errorf("DetailedString() should contain file location")
	}
	if !containsString(str, "interval") {
		t.Errorf("DetailedString() should contain mapping type")
	}
	if !containsString(str, "0.95") {
		t.Errorf("DetailedString() should contain confidence")
	}
}

func TestIntervalsOverlap(t *testing.T) {
	tests := []struct {
		name string
		a    AdditionInterval
		b    AdditionInterval
		want bool
	}{
		{
			name: "non-overlapping - a before b",
			a:    AdditionInterval{GenStart: 10, GenEnd: 20},
			b:    AdditionInterval{GenStart: 30, GenEnd: 40},
			want: false,
		},
		{
			name: "non-overlapping - b before a",
			a:    AdditionInterval{GenStart: 30, GenEnd: 40},
			b:    AdditionInterval{GenStart: 10, GenEnd: 20},
			want: false,
		},
		{
			name: "overlapping - partial",
			a:    AdditionInterval{GenStart: 10, GenEnd: 30},
			b:    AdditionInterval{GenStart: 20, GenEnd: 40},
			want: true,
		},
		{
			name: "overlapping - one contains other",
			a:    AdditionInterval{GenStart: 10, GenEnd: 40},
			b:    AdditionInterval{GenStart: 20, GenEnd: 30},
			want: true,
		},
		{
			name: "touching - a end == b start",
			a:    AdditionInterval{GenStart: 10, GenEnd: 20},
			b:    AdditionInterval{GenStart: 20, GenEnd: 30},
			want: true,
		},
		{
			name: "same line - non-overlapping columns",
			a:    AdditionInterval{GenStart: 10, GenEnd: 10, GenStartCol: 1, GenEndCol: 5},
			b:    AdditionInterval{GenStart: 10, GenEnd: 10, GenStartCol: 10, GenEndCol: 15},
			want: false,
		},
		{
			name: "same line - overlapping columns",
			a:    AdditionInterval{GenStart: 10, GenEnd: 10, GenStartCol: 1, GenEndCol: 10},
			b:    AdditionInterval{GenStart: 10, GenEnd: 10, GenStartCol: 5, GenEndCol: 15},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intervalsOverlap(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("intervalsOverlap() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

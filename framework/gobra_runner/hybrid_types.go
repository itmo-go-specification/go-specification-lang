package gobra_runner

// Position represents a position in source code
type Position struct {
	File   string
	Line   int
	Column int
}

// Interval represents an interval of code that was added during translation
type Interval struct {
	// GenStart is the starting line in the generated code
	GenStart int

	// GenEnd is the ending line in the generated code
	GenEnd int

	// GenStartCol is the starting column in the generated code (optional)
	GenStartCol int

	// GenEndCol is the ending column in the generated code (optional)
	GenEndCol int

	// SourcePos is the position in the original source code
	SourcePos Position

	// PredicateSpan is an optional precise mapping for predicates
	PredicateSpan *Position
}

// IsValid checks if the interval is valid (start <= end, positions are positive)
func (a *Interval) IsValid() bool {
	if a.GenStart <= 0 || a.GenEnd <= 0 {
		return false
	}
	if a.GenStart > a.GenEnd {
		return false
	}
	if a.GenStart == a.GenEnd && a.GenStartCol > 0 && a.GenEndCol > 0 {
		if a.GenStartCol > a.GenEndCol {
			return false
		}
	}
	return true
}

// Contains checks if a given position (line, col) is within this interval
func (a *Interval) Contains(line, col int) bool {
	if line < a.GenStart || line > a.GenEnd {
		return false
	}

	// If on the start line, check column
	if line == a.GenStart && a.GenStartCol > 0 && col < a.GenStartCol {
		return false
	}

	// If on the end line, check column
	if line == a.GenEnd && a.GenEndCol > 0 && col > a.GenEndCol {
		return false
	}

	return true
}

// Length returns the length of this interval in lines
func (a *Interval) Length() int {
	return a.GenEnd - a.GenStart + 1
}

// PointMapping represents a point-to-point mapping
type PointMapping struct {
	GenLine   int
	GenColumn int
	Orig      Position
}

// HybridMapping is a container for tracking code additions and their mappings
type HybridMapping struct {
	// SourceFile is the path to the original Go source file
	SourceFile string

	// Additions is a sorted list of all code additions, sorted by GenStart
	Additions []Interval

	// PointMappings contains optional point-to-point mappings
	PointMappings []PointMapping
}

// NewHybridMapping creates a new empty HybridMapping
func NewHybridMapping(sourceFile string) *HybridMapping {
	return &HybridMapping{
		SourceFile:    sourceFile,
		Additions:     make([]Interval, 0),
		PointMappings: make([]PointMapping, 0),
	}
}

package gobra_runner

// Position represents a position in source code using byte offset
type Position struct {
	File   string
	Offset int // Byte offset in the file (0-based)
}

// ToLineCol converts the offset to line and column numbers
// This should only be called for final error reporting
func (p *Position) ToLineCol(content []byte) (line, col int) {
	if p.Offset < 0 {
		return 1, 1
	}
	if p.Offset >= len(content) {
		// Offset is at end of file
		line = 1
		col = len(content) + 1
		for i := 0; i < len(content); i++ {
			if content[i] == '\n' {
				line++
				col = 1
			} else {
				col++
			}
		}
		return line, col
	}

	line = 1
	col = 1
	for i := 0; i < p.Offset; i++ {
		if content[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}

	return line, col
}

// Interval represents an interval of code that was added during translation
type Interval struct {
	// GenStart is the starting byte offset in the generated code
	GenStart int

	// GenEnd is the ending byte offset in the generated code
	GenEnd int

	// SourcePos is the position in the original source code
	SourcePos Position

	// PredicateSpan is an optional precise mapping for predicates
	PredicateSpan *Position
}

// IsValid checks if the interval is valid (start <= end, positions are non-negative)
func (a *Interval) IsValid() bool {
	if a.GenStart < 0 || a.GenEnd < 0 {
		return false
	}
	if a.GenStart > a.GenEnd {
		return false
	}
	return true
}

// Contains checks if a given offset is within this interval
func (a *Interval) Contains(offset int) bool {
	return offset >= a.GenStart && offset <= a.GenEnd
}

// Length returns the length of this interval in bytes
func (a *Interval) Length() int {
	return a.GenEnd - a.GenStart + 1
}

// PointMapping represents a point-to-point mapping using byte offsets
type PointMapping struct {
	GenOffset int // Offset in the generated code
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

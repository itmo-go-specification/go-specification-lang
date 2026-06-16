package gobra_runner

// Interval represents an interval of code that was added during translation
type Interval struct {
	// GenStart is the starting byte offset in the generated code
	GenStart int

	// GenEnd is the ending byte offset in the generated code
	GenEnd int

	// SourcePos is the position in the original source code
	SourcePos SourcePosition

	// SourceEnd is the end position in the original source code
	SourceEnd SourcePosition

	// PredicateSpan is an optional precise mapping for predicates
	PredicateSpan *SourcePosition
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

// Delta returns the length of the added interval minus the length of the replaced interval
func (a *Interval) Delta() int {
	return a.GenEnd - a.GenStart + 1 - (a.SourceEnd.Offset - a.SourcePos.Offset)
}

// PointMapping represents a point-to-point mapping using byte offsets
type PointMapping struct {
	GenOffset int // Offset in the generated code
	Orig      SourcePosition
}

// HybridMapping is a container for tracking code additions and their mappings
// Implements SourceMapping interface
type HybridMapping struct {
	// Replacements is a sorted list of all code replacements, sorted by GenStart
	Replacements []Interval

	// PointMappings contains optional point-to-point mappings
	PointMappings []PointMapping
	SourceFile    string
}

// NewHybridMapping creates a new empty HybridMapping
func NewHybridMapping(sourceFile string) *HybridMapping {
	return &HybridMapping{
		Replacements:     make([]Interval, 0),
		PointMappings: make([]PointMapping, 0),
		SourceFile:    sourceFile,
	}
}

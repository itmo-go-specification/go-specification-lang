package gobra_runner

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// AdditionType represents the type of code addition that was made during translation
type AdditionType string

const (
	// AdditionRequires represents a requires clause added to a function
	AdditionRequires AdditionType = "requires"
	// AdditionEnsures represents an ensures clause added to a function
	AdditionEnsures AdditionType = "ensures"
	// AdditionAssert represents an assert statement added to the code
	AdditionAssert AdditionType = "assert"
	// AdditionInvariant represents a loop invariant added to a loop
	AdditionInvariant AdditionType = "invariant"
	// AdditionDecreases represents a decreases clause added to a loop
	AdditionDecreases AdditionType = "decreases"
	// AdditionPredicate represents a predicate definition added to the code
	AdditionPredicate AdditionType = "predicate"
	// AdditionLemma represents a lemma added to the code
	AdditionLemma AdditionType = "lemma"
	// AdditionOther represents any other type of addition
	AdditionOther AdditionType = "other"
)

// AdditionInterval represents an interval of code that was added during translation
// from Go to Gobra. It tracks both the position in the generated code and the
// corresponding position in the original source code.
type AdditionInterval struct {
	// GenStart is the starting position (line number) in the generated Gobra code
	GenStart int `json:"genStart"`

	// GenEnd is the ending position (line number) in the generated Gobra code
	GenEnd int `json:"genEnd"`

	// GenStartCol is the starting column in the generated Gobra code (optional)
	GenStartCol int `json:"genStartCol,omitempty"`

	// GenEndCol is the ending column in the generated Gobra code (optional)
	GenEndCol int `json:"genEndCol,omitempty"`

	// SourcePos is the position in the original source code where this addition
	// conceptually belongs (e.g., the line where a function is defined)
	SourcePos SourcePos `json:"sourcePos"`

	// Kind specifies the type of addition (requires, ensures, assert, etc.)
	Kind AdditionType `json:"kind"`

	// PredicateSpan is an optional precise mapping for predicates within the addition.
	// If provided, errors within this interval can be mapped to specific predicate
	// locations in the source code.
	PredicateSpan *SourcePos `json:"predicateSpan,omitempty"`

	// Description provides additional context about this addition (optional)
	Description string `json:"description,omitempty"`
}

// IsValid checks if the interval is valid (start <= end, positions are positive)
func (a *AdditionInterval) IsValid() bool {
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
func (a *AdditionInterval) Contains(line, col int) bool {
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
func (a *AdditionInterval) Length() int {
	return a.GenEnd - a.GenStart + 1
}

// HybridMapping is a container for tracking code additions and their mappings
// between generated Gobra code and original Go source code.
type HybridMapping struct {
	// GenFile is the path to the generated Gobra file
	GenFile string `json:"genFile"`

	// SourceFile is the path to the original Go source file
	SourceFile string `json:"sourceFile"`

	// Additions is a sorted list of all code additions, sorted by GenStart
	Additions []AdditionInterval `json:"additions"`

	// PointMappings contains optional point-to-point mappings for specific
	// locations that don't fit into the interval model
	PointMappings []MappingEntry `json:"pointMappings,omitempty"`
}

// NewHybridMapping creates a new empty HybridMapping
func NewHybridMapping(genFile, sourceFile string) *HybridMapping {
	return &HybridMapping{
		GenFile:       genFile,
		SourceFile:    sourceFile,
		Additions:     make([]AdditionInterval, 0),
		PointMappings: make([]MappingEntry, 0),
	}
}

// AddInterval adds a new addition interval to the mapping
func (h *HybridMapping) AddInterval(interval AdditionInterval) error {
	if !interval.IsValid() {
		return fmt.Errorf("invalid addition interval: %+v", interval)
	}

	// Check for overlaps with existing intervals
	for _, existing := range h.Additions {
		if intervalsOverlap(existing, interval) {
			return fmt.Errorf("interval overlaps with existing interval: existing=[%d:%d], new=[%d:%d]",
				existing.GenStart, existing.GenEnd, interval.GenStart, interval.GenEnd)
		}
	}

	h.Additions = append(h.Additions, interval)
	return nil
}

// AddPointMapping adds a point-to-point mapping
func (h *HybridMapping) AddPointMapping(mapping MappingEntry) {
	h.PointMappings = append(h.PointMappings, mapping)
}

// Sort sorts all additions and point mappings by their position in the generated code
func (h *HybridMapping) Sort() {
	// Sort additions by GenStart, then GenStartCol
	sort.Slice(h.Additions, func(i, j int) bool {
		a, b := h.Additions[i], h.Additions[j]
		if a.GenStart != b.GenStart {
			return a.GenStart < b.GenStart
		}
		if a.GenStartCol != b.GenStartCol {
			return a.GenStartCol < b.GenStartCol
		}
		return a.GenEnd < b.GenEnd
	})

	// Sort point mappings by GenLine, then GenColumn
	sort.Slice(h.PointMappings, func(i, j int) bool {
		a, b := h.PointMappings[i], h.PointMappings[j]
		if a.GenLine != b.GenLine {
			return a.GenLine < b.GenLine
		}
		return a.GenColumn < b.GenColumn
	})
}

// Validate checks if the mapping is valid (no overlapping intervals, sorted)
func (h *HybridMapping) Validate() error {
	// Check if additions are sorted
	for i := 1; i < len(h.Additions); i++ {
		prev, curr := h.Additions[i-1], h.Additions[i]
		if prev.GenStart > curr.GenStart {
			return fmt.Errorf("additions not sorted: interval %d starts before interval %d", i-1, i)
		}
		if intervalsOverlap(prev, curr) {
			return fmt.Errorf("overlapping intervals: [%d:%d] and [%d:%d]",
				prev.GenStart, prev.GenEnd, curr.GenStart, curr.GenEnd)
		}
	}

	// Validate each interval
	for i, interval := range h.Additions {
		if !interval.IsValid() {
			return fmt.Errorf("invalid interval at index %d: %+v", i, interval)
		}
	}

	return nil
}

// FindInterval finds the addition interval that contains the given position
func (h *HybridMapping) FindInterval(line, col int) *AdditionInterval {
	// Binary search for the interval
	idx := sort.Search(len(h.Additions), func(i int) bool {
		return h.Additions[i].GenStart > line
	})

	// Check the interval before the found position
	if idx > 0 {
		interval := &h.Additions[idx-1]
		if interval.Contains(line, col) {
			return interval
		}
	}

	return nil
}

// FindPointMapping finds a point mapping for the given position
func (h *HybridMapping) FindPointMapping(line, col int) *MappingEntry {
	// Binary search for the point mapping
	idx := sort.Search(len(h.PointMappings), func(i int) bool {
		m := h.PointMappings[i]
		if m.GenLine != line {
			return m.GenLine > line
		}
		return m.GenColumn >= col
	})

	// Check if we found an exact match
	if idx < len(h.PointMappings) {
		m := h.PointMappings[idx]
		if m.GenLine == line && m.GenColumn == col {
			return &m
		}
	}

	return nil
}

// GetTotalAdditionsBefore returns the total number of lines added before a given position
func (h *HybridMapping) GetTotalAdditionsBefore(line int) int {
	total := 0
	for _, interval := range h.Additions {
		if interval.GenEnd < line {
			total += interval.Length()
		}
	}
	return total
}

// intervalsOverlap checks if two intervals overlap
func intervalsOverlap(a, b AdditionInterval) bool {
	// Check if intervals overlap (including touching)
	if a.GenEnd < b.GenStart || b.GenEnd < a.GenStart {
		return false
	}

	// If they're on the same line, check columns
	if a.GenStart == b.GenStart && a.GenEnd == b.GenEnd {
		if a.GenStartCol > 0 && b.GenEndCol > 0 {
			// Both have column info - check overlap
			return !(a.GenEndCol < b.GenStartCol || b.GenEndCol < a.GenStartCol)
		}
	}

	return true
}

// ToJSON converts the HybridMapping to JSON
func (h *HybridMapping) ToJSON() ([]byte, error) {
	return json.MarshalIndent(h, "", "  ")
}

// FromJSON creates a HybridMapping from JSON
func FromJSON(data []byte) (*HybridMapping, error) {
	var mapping HybridMapping
	if err := json.Unmarshal(data, &mapping); err != nil {
		return nil, err
	}
	return &mapping, nil
}

// SaveHybridMappingToFile saves a HybridMapping to a file
func SaveHybridMappingToFile(mapping *HybridMapping, path string) error {
	data, err := mapping.ToJSON()
	if err != nil {
		return err
	}
	return writeFile(path, data, 0o644)
}

// LoadHybridMappingFromFile loads a HybridMapping from a file
func LoadHybridMappingFromFile(path string) (*HybridMapping, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, err
	}
	return FromJSON(data)
}

// Helper functions for file operations (to avoid os import in tests)
func writeFile(path string, data []byte, perm os.FileMode) error {
	return writeFileImpl(path, data, perm)
}

func readFile(path string) ([]byte, error) {
	return readFileImpl(path)
}

// These are implemented in mapping.go to avoid circular imports
var (
	writeFileImpl func(path string, data []byte, perm os.FileMode) error
	readFileImpl  func(path string) ([]byte, error)
)

func initFileOps(write func(path string, data []byte, perm os.FileMode) error, read func(path string) ([]byte, error)) {
	writeFileImpl = write
	readFileImpl = read
}

package gobra_runner

import (
	"encoding/json"
	"os"
	"sort"
)

// SourcePos represents a position in source code (legacy, use Position instead)
// Deprecated: Use Position instead
type SourcePos struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"col"`
}

// MappingEntry represents a mapping entry (legacy, use PointMapping instead)
// Deprecated: Use PointMapping instead
type MappingEntry struct {
	GenLine   int       `json:"genLine"`
	GenColumn int       `json:"genCol"`
	Orig      SourcePos `json:"orig"`
	Kind      string    `json:"kind"`
}

// CoordMapping represents coordinate mappings (legacy)
// Deprecated: Use HybridMapping instead
type CoordMapping struct {
	GenFile string         `json:"genFile"`
	Entries []MappingEntry `json:"entries"`
}

// Lookup finds a mapping entry for the given position
// Deprecated: Use HybridMapping.MapError instead
func (m *CoordMapping) Lookup(line, col int) (MappingEntry, bool) {
	if len(m.Entries) == 0 {
		return MappingEntry{}, false
	}

	idx := sort.Search(len(m.Entries), func(i int) bool {
		e := m.Entries[i]
		if e.GenLine != line {
			return e.GenLine > line
		}
		return e.GenColumn > col
	})

	if idx == 0 {
		return MappingEntry{}, false
	}
	return m.Entries[idx-1], true
}

// Sort sorts the mapping entries
// Deprecated: HybridMapping maintains sorted order automatically
func (m *CoordMapping) Sort() {
	sort.Slice(m.Entries, func(i, j int) bool {
		a, b := m.Entries[i], m.Entries[j]
		if a.GenLine != b.GenLine {
			return a.GenLine < b.GenLine
		}
		return a.GenColumn < b.GenColumn
	})
}

// LoadMappingFromFile loads a coordinate mapping from a file
// Deprecated: Use HybridMapping with JSON serialization instead
func LoadMappingFromFile(path string) (*CoordMapping, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m CoordMapping
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// SaveMappingToFile saves a coordinate mapping to a file
// Deprecated: Use HybridMapping with JSON serialization instead
func SaveMappingToFile(m *CoordMapping, path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

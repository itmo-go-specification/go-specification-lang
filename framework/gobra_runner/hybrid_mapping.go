package gobra_runner

import (
	"fmt"
	"sort"
)

// AddInterval adds a new addition interval to the mapping.
// Automatically maintains sorted order and validates overlaps.
func (h *HybridMapping) AddInterval(interval Interval) error {
	if !interval.IsValid() {
		return fmt.Errorf("invalid addition interval: %+v", interval)
	}

	// Find insertion position using binary search
	insertPos := sort.Search(len(h.Additions), func(i int) bool {
		return h.Additions[i].GenStart > interval.GenStart
	})

	// Check for overlaps with adjacent intervals
	if insertPos > 0 {
		prev := &h.Additions[insertPos-1]
		if intervalsOverlap(*prev, interval) {
			return fmt.Errorf("interval overlaps with previous interval: [%d:%d] and [%d:%d]",
				prev.GenStart, prev.GenEnd, interval.GenStart, interval.GenEnd)
		}
	}
	if insertPos < len(h.Additions) {
		next := &h.Additions[insertPos]
		if intervalsOverlap(interval, *next) {
			return fmt.Errorf("interval overlaps with next interval: [%d:%d] and [%d:%d]",
				interval.GenStart, interval.GenEnd, next.GenStart, next.GenEnd)
		}
	}

	// Insert at the correct position to maintain sorted order
	h.Additions = append(h.Additions, Interval{})
	copy(h.Additions[insertPos+1:], h.Additions[insertPos:])
	h.Additions[insertPos] = interval

	return nil
}

// AddPointMapping adds a point-to-point mapping.
// Automatically maintains sorted order.
func (h *HybridMapping) AddPointMapping(mapping PointMapping) {
	// Find insertion position using binary search
	insertPos := sort.Search(len(h.PointMappings), func(i int) bool {
		m := h.PointMappings[i]
		if m.GenLine != mapping.GenLine {
			return m.GenLine > mapping.GenLine
		}
		return m.GenColumn >= mapping.GenColumn
	})

	// Insert at the correct position to maintain sorted order
	h.PointMappings = append(h.PointMappings, PointMapping{})
	copy(h.PointMappings[insertPos+1:], h.PointMappings[insertPos:])
	h.PointMappings[insertPos] = mapping
}

// FindInterval finds the addition interval that contains the given position
func (h *HybridMapping) FindInterval(line, col int) *Interval {
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
func (h *HybridMapping) FindPointMapping(line, col int) *PointMapping {
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

// MapError maps an error position from generated code to the original source code
func (h *HybridMapping) MapError(line, col int) (Position, error) {
	// Validate input
	if line <= 0 || col <= 0 {
		return Position{}, fmt.Errorf("invalid position: line=%d, col=%d", line, col)
	}

	// Level 1: Check if the error is within an addition interval
	interval := h.FindInterval(line, col)
	if interval != nil {
		// Use point mapping if available
		pointMapping := h.FindPointMapping(line, col)
		if pointMapping != nil {
			return pointMapping.Orig, nil
		}

		// Use predicate span if available
		if interval.PredicateSpan != nil {
			return *interval.PredicateSpan, nil
		}

		// Fall back to the general source position for this addition
		return interval.SourcePos, nil
	}

	// Level 2: Error is in original code - compute position by subtracting additions
	originalLine := line - h.GetTotalAdditionsBefore(line)

	// Ensure we don't return negative or zero line numbers
	if originalLine <= 0 {
		originalLine = 1
	}

	return Position{
		File:   h.SourceFile,
		Line:   originalLine,
		Column: col,
	}, nil
}

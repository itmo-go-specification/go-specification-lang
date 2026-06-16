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
	insertPos := sort.Search(len(h.Replacements), func(i int) bool {
		return h.Replacements[i].GenStart > interval.GenStart
	})

	// Check for overlaps with adjacent intervals
	if insertPos > 0 {
		prev := &h.Replacements[insertPos-1]
		if intervalsOverlap(*prev, interval) {
			return fmt.Errorf("interval overlaps with previous interval: [%d:%d] and [%d:%d]",
				prev.GenStart, prev.GenEnd, interval.GenStart, interval.GenEnd)
		}
	}
	if insertPos < len(h.Replacements) {
		next := &h.Replacements[insertPos]
		if intervalsOverlap(interval, *next) {
			return fmt.Errorf("interval overlaps with next interval: [%d:%d] and [%d:%d]",
				interval.GenStart, interval.GenEnd, next.GenStart, next.GenEnd)
		}
	}

	// Insert at the correct position to maintain sorted order
	h.Replacements = append(h.Replacements, Interval{})
	copy(h.Replacements[insertPos+1:], h.Replacements[insertPos:])
	h.Replacements[insertPos] = interval

	return nil
}

// AddPointMapping adds a point-to-point mapping.
// Automatically maintains sorted order.
func (h *HybridMapping) AddPointMapping(mapping PointMapping) {
	// Find insertion position using binary search
	insertPos := sort.Search(len(h.PointMappings), func(i int) bool {
		return h.PointMappings[i].GenOffset >= mapping.GenOffset
	})

	// Insert at the correct position to maintain sorted order
	h.PointMappings = append(h.PointMappings, PointMapping{})
	copy(h.PointMappings[insertPos+1:], h.PointMappings[insertPos:])
	h.PointMappings[insertPos] = mapping
}

// FindInterval finds the addition interval that contains the given offset
func (h *HybridMapping) FindInterval(offset int) *Interval {
	// Binary search for the interval
	idx := sort.Search(len(h.Replacements), func(i int) bool {
		return h.Replacements[i].GenStart > offset
	})

	// Check the interval before the found position
	if idx > 0 {
		interval := &h.Replacements[idx-1]
		if interval.Contains(offset) {
			return interval
		}
	}

	return nil
}

// FindPointMapping finds a point mapping for the given offset
func (h *HybridMapping) FindPointMapping(offset int) *PointMapping {
	// Binary search for the point mapping
	idx := sort.Search(len(h.PointMappings), func(i int) bool {
		return h.PointMappings[i].GenOffset >= offset
	})

	// Check if we found an exact match
	if idx < len(h.PointMappings) {
		m := h.PointMappings[idx]
		if m.GenOffset == offset {
			return &m
		}
	}

	return nil
}

// GetTotalAdditionsBefore returns the total number of bytes added before a given offset
func (h *HybridMapping) GetTotalAdditionsBefore(offset int) int {
	total := 0
	for _, interval := range h.Replacements {
		if interval.GenEnd < offset {
			total += interval.Delta()
		}
	}
	return total
}

// MapError maps an error position from generated code to the original source code
// This works with byte offsets internally
func (h *HybridMapping) MapError(offset int) (SourcePosition, error) {
	// Validate input
	if offset < 0 {
		return SourcePosition{}, fmt.Errorf("invalid offset: %d", offset)
	}

	// Level 1: Check if the error is within an addition interval
	interval := h.FindInterval(offset)
	if interval != nil {
		// Use point mapping if available
		pointMapping := h.FindPointMapping(offset)
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
	originalOffset := offset - h.GetTotalAdditionsBefore(offset)

	// Ensure we don't return negative offsets
	if originalOffset < 0 {
		panic("impossible negative offset")
		originalOffset = 0
	}

	return SourcePosition{
		File:   h.SourceFile,
		Offset: originalOffset,
	}, nil
}

// MapErrorLineCol maps an error position from generated code (line, col) to the original source code
// This is a convenience function that converts line/col to offset, maps it, then converts back
// This should only be used for final error reporting
func (h *HybridMapping) MapErrorLineCol(line, col int, genContent []byte) (SourcePosition, error) {
	// Convert line/col to offset using OffsetToLineCol in reverse
	offset, err := LineColToOffset(genContent, line, col)
	if err != nil {
		return SourcePosition{}, fmt.Errorf("failed to convert line:col to offset: %w", err)
	}

	// Map the error using offset
	return h.MapError(offset)
}

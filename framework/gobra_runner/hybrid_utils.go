package gobra_runner

import (
	"fmt"
)

// LineColToOffset converts a line:column position to a byte offset
// This is a helper function that can be used when working with file contents
func LineColToOffset(lines []string, line, col int) (int, error) {
	if line <= 0 || col <= 0 {
		return 0, fmt.Errorf("invalid position: line=%d, col=%d", line, col)
	}
	if line > len(lines) {
		return 0, fmt.Errorf("line %d exceeds file length %d", line, len(lines))
	}

	offset := 0
	for i := 0; i < line-1; i++ {
		offset += len(lines[i]) + 1 // +1 for newline
	}

	// Add column offset (convert to 0-based)
	if col > len(lines[line-1])+1 {
		return 0, fmt.Errorf("column %d exceeds line length %d", col, len(lines[line-1])+1)
	}
	offset += col - 1

	return offset, nil
}

// OffsetToLineCol converts a byte offset to a line:column position
// This is a helper function that can be used when working with file contents
func OffsetToLineCol(content []byte, offset int) (line, col int, err error) {
	if offset < 0 || offset > len(content) {
		return 0, 0, fmt.Errorf("offset %d out of bounds [0, %d]", offset, len(content))
	}

	line = 1
	col = 1
	for i := 0; i < offset; i++ {
		if content[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}

	return line, col, nil
}

// intervalsOverlap checks if two intervals overlap
func intervalsOverlap(a, b Interval) bool {
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

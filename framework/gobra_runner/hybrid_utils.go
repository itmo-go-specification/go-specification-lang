package gobra_runner

import (
	"fmt"
)

// LineColToOffset converts a line:column position to a byte offset
// This is a helper function that can be used when working with file contents
func LineColToOffset(content []byte, line, col int) (int, error) {
	if line <= 0 || col <= 0 {
		return 0, fmt.Errorf("invalid position: line=%d, col=%d", line, col)
	}

	offset := 0
	currentLine := 1
	currentCol := 1

	for i := 0; i < len(content); i++ {
		if currentLine == line && currentCol == col {
			return offset, nil
		}

		if content[i] == '\n' {
			currentLine++
			currentCol = 1
		} else {
			currentCol++
		}
		offset++
	}

	// If we reach the end, check if the position is at the end of the file
	if currentLine == line && currentCol == col {
		return offset, nil
	}

	return 0, fmt.Errorf("position line:%d, col:%d not found in content", line, col)
}

// OffsetToLineCol converts a byte offset to a line:column position
// This is a helper function that can be used when working with file contents
func OffsetToLineCol(content []byte, offset int) (line, col int, err error) {
	if offset < 0 || offset > len(content) {
		fmt.Printf("offset %d out of bounds [0, %d]", offset, len(content))
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

	return true
}

package gobra

import (
	"fmt"
	"gpma/framework/data"
	specir "gpma/framework/gosl/ir"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
)

// ErrorLocation holds the parsed coordinates from a Gobra error message.
type ErrorLocation struct {
	Line   int
	Column int
}

// errorLocationRe matches patterns like </expr.gobra:84:5> and captures
// the file name, line number, and column number.
var errorLocationRe = regexp.MustCompile(`</?([^>:]+):(\d+):(\d+)>`)

// FindErrorLocations scans s for Gobra-style error location tags and returns
// all matches with their parsed coordinates.
func FindErrorLocations(s string) []ErrorLocation {
	matches := errorLocationRe.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return nil
	}
	results := make([]ErrorLocation, 0, len(matches))
	for _, m := range matches {
		line, err := strconv.Atoi(m[2])
		if err != nil {
			continue
		}
		col, err := strconv.Atoi(m[3])
		if err != nil {
			continue
		}
		results = append(results, ErrorLocation{
			Line:   line,
			Column: col,
		})
	}
	return results
}

// OffsetToLineCol converts a byte offset into a 1-based (line, col) pair.
func OffsetToLineCol(text string, offset int) (line, col int, err error) {
	if offset < 0 || offset > len(text) {
		return 0, 0, fmt.Errorf("offset %d out of range [0, %d]", offset, len(text))
	}
	line = 1
	col = 1
	for i, ch := range text {
		if i == offset {
			return line, col, nil
		}
		if ch == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col, nil // offset == len(text), i.e. one past the last character
}

// LineColToOffset converts a 1-based (line, col) pair into a byte offset.
func LineColToOffset(text string, line, col int) (int, error) {
	if line < 1 || col < 1 {
		return 0, fmt.Errorf("line and col are 1-based, got line=%d col=%d", line, col)
	}
	currentLine := 1
	currentCol := 1
	for i, ch := range text {
		if currentLine == line && currentCol == col {
			return i, nil
		}
		if ch == '\n' {
			currentLine++
			currentCol = 1
		} else {
			currentCol++
		}
	}
	// handle position one past the last character
	if currentLine == line && currentCol == col {
		return len(text), nil
	}
	return 0, fmt.Errorf("position %d:%d not found in text", line, col)
}

// RunGobraImage runs the const docker image, mounting inputFile into the
// container at /input/<basename> and passing that path as the sole argument.
// Stderr is forwarded to the current process; stdout is returned as a string.
func RunGobraImage(inputFile string) (string, error) {
	absPath, err := filepath.Abs(inputFile)
	if err != nil {
		return "", fmt.Errorf("resolve input file path: %w", err)
	}

	if _, err := os.Stat(absPath); err != nil {
		return "", fmt.Errorf("input file not accessible: %w", err)
	}

	containerPath := "/input/" + filepath.Base(absPath)

	cmd := exec.Command(
		"docker", "run", "--rm",
		"-v", absPath+":"+containerPath+":ro",
		"ghcr.io/viperproject/gobra:sha-7fa4f86",
		"-i", containerPath,
	)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	//if err != nil {
	//	return "", fmt.Errorf("docker run: %w", err)
	//}
	return string(out), nil
}

// WriteTempFile writes content to a newly created temporary file and returns
// its path. The caller is responsible for removing the file when done.
func WriteTempFile(content string) (string, error) {
	f, err := os.CreateTemp("", "gpma-*.txt")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		os.Remove(f.Name())
		return "", fmt.Errorf("write temp file: %w", err)
	}
	return f.Name(), nil
}

// FindNamedPredicateByStartOffset searches all NodeInfo entries for a
// NamedPredicate whose Span.Start equals the given offset.
// Returns the first match found, or nil if none matches.
func FindNamedPredicateByStartOffset(data *data.ASTData, offset int) *specir.NamedPredicate {
	for _, info := range data.NodeInfo {
		if info == nil {
			continue
		}
		if fc := info.FuncContract; fc != nil {
			for i := range fc.Requires {
				if fc.Requires[i].Span.Start == offset {
					return &fc.Requires[i]
				}
			}
			for i := range fc.Ensures {
				if fc.Ensures[i].Span.Start == offset {
					return &fc.Ensures[i]
				}
			}
		}
		if ls := info.LoopSpec; ls != nil {
			for i := range ls.Invariants {
				if ls.Invariants[i].Span.Start == offset {
					return &ls.Invariants[i]
				}
			}
		}
		if as := info.Asserts; as != nil {
			if as.Assertion.Span.Start == offset {
				return &as.Assertion
			}
		}
	}
	return nil
}


func RunGobraVerifier(astData *data.ASTData) ([]specir.NamedPredicate, error) {
	generated, err := GenerateGobra(astData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate gobra: %v", err)
	}

	tmpFile, err := WriteTempFile(generated.GetGobraSource())
	if err != nil {
		return nil, fmt.Errorf("failed to write to /tmp: %v", err)
	}
	defer os.Remove(tmpFile)

	gobraReport, err := RunGobraImage(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get gobra report: %v", err)
	}

	failed_predicates := []specir.NamedPredicate{}
	for _, error_location := range FindErrorLocations(gobraReport) {
		gobraOffset, err := LineColToOffset(generated.GetGobraSource(), error_location.Line, error_location.Column)
		if err != nil {
			return nil, fmt.Errorf("failed to get error offset: %v", err)
		}

		sourceOffset, found := generated.predicateSpansMap[gobraOffset]
		if !found {
			line, col, err := OffsetToLineCol(astData.Source, sourceOffset.Start)
			if err != nil {
				return nil, fmt.Errorf("failed to get error coords: %v", err)
			}
			return nil, fmt.Errorf("failed to find corresponding predicate offset: %d:%d", line, col)
		}

		predicate := FindNamedPredicateByStartOffset(astData, sourceOffset.Start)
		if predicate == nil {
			return nil, fmt.Errorf("failed to find a predicate with given starting offset: %d", sourceOffset.Start)
		}

		failed_predicates = append(failed_predicates, *predicate)
	}

	return failed_predicates, nil
}

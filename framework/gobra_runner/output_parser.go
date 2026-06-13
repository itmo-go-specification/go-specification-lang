package gobra_runner

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type rawError struct {
	genFile   string
	genLine   int
	genColumn int
	message   string
}

var (
	reGobraError = regexp.MustCompile(
		`(?m)^\[ERROR\]\s+(.+?)@(\d+)\.(\d+):\s*(.+)$`,
	)

	reGobraErrorAlt = regexp.MustCompile(
		`(?m)^ERROR at '(.+?)'\s*@\s*\((\d+),\s*(\d+)\):\s*(.+)$`,
	)

	reGobraLogError = regexp.MustCompile(
		`Error at:\s*<(.+):(\d+):(\d+)>\s*(.+)`,
	)

	reSuccess = regexp.MustCompile(`(?i)successfully verified`)
)

func parseGobraOutput(output, genFilePath string) []rawError {
	var errors []rawError

	patterns := []*regexp.Regexp{reGobraError, reGobraErrorAlt, reGobraLogError}

	for _, pattern := range patterns {
		for _, m := range pattern.FindAllStringSubmatch(output, -1) {
			file := strings.TrimSpace(m[1])
			line, _ := strconv.Atoi(m[2])
			col, _ := strconv.Atoi(m[3])
			msg := strings.TrimSpace(m[4])

			if genFilePath != "" && !isSameFile(file, genFilePath) {
				continue
			}

			errors = append(errors, rawError{
				genFile:   file,
				genLine:   line,
				genColumn: col,
				message:   msg,
			})
		}
		if len(errors) > 0 {
			break
		}
	}

	return errors
}

func isSameFile(a, b string) bool {
	if a == b {
		return true
	}
	return filepath.Base(a) == filepath.Base(b)
}

func isSuccess(output string) bool {
	return reSuccess.MatchString(output)
}

func inferKind(message string) string {
	msg := strings.ToLower(message)
	switch {
	case strings.Contains(msg, "precondition"):
		return "precondition"
	case strings.Contains(msg, "postcondition"):
		return "postcondition"
	case strings.Contains(msg, "invariant"):
		return "invariant"
	case strings.Contains(msg, "assert"):
		return "assertion"
	default:
		return "unknown"
	}
}

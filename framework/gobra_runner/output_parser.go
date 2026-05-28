package gobra_runner

import (
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

	reSuccess = regexp.MustCompile(`(?m)^(?:\[INFO\]\s+)?[Ss]uccessfully verified`)
)

func parseGobraOutput(output, genFilePath string) []rawError {
	var errors []rawError

	for _, pattern := range []*regexp.Regexp{reGobraError, reGobraErrorAlt} {
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
	return a == b || strings.HasSuffix(a, b) || strings.HasSuffix(b, a)
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

package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckAt_TrueCondition_NoPanic(t *testing.T) {
	t.Setenv(modeEnv, "panic")
	CheckAt(true, "test", 1)
	CheckAt(true, "test", 1, "should not panic")
	CheckAt(true, "test", 1, "msg1", "msg2")
}

func TestCheckAt_FalseCondition_PanicMode(t *testing.T) {
	t.Setenv(modeEnv, "panic")

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got none")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !strings.Contains(msg, "check failed at") {
			t.Errorf("panic message missing prefix, got: %s", msg)
		}
		if !strings.Contains(msg, "test.go:42") {
			t.Errorf("panic message missing source location, got: %s", msg)
		}
	}()

	CheckAt(false, "test.go", 42)
}

func TestCheckAt_FalseCondition_PanicMode_WithMessage(t *testing.T) {
	t.Setenv(modeEnv, "panic")

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got none")
		}
		msg := r.(string)
		if !strings.Contains(msg, "precondition violated") {
			t.Errorf("panic message missing custom text, got: %s", msg)
		}
	}()

	CheckAt(false, "test.go", 1, "precondition violated")
}

func TestCheckAt_FalseCondition_ConsoleMode(t *testing.T) {
	t.Setenv(modeEnv, "console")

	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	CheckAt(false, "test.go", 1, "console test")

	w.Close()
	os.Stderr = oldStderr

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "check failed at") {
		t.Errorf("expected 'check failed at' in stderr, got: %s", output)
	}
	if !strings.Contains(output, "console test") {
		t.Errorf("expected 'console test' in stderr, got: %s", output)
	}
}

func TestCheckAt_FalseCondition_SilentMode(t *testing.T) {
	t.Setenv(modeEnv, "silent")
	CheckAt(false, "test.go", 1, "should be silent")
}

func TestCheckAt_FalseCondition_FileMode(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "check.log")

	t.Setenv(modeEnv, "file")
	t.Setenv(pathEnv, logPath)

	CheckAt(false, "test.go", 1, "file mode test")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "check failed at") {
		t.Errorf("log file missing 'check failed at', got: %s", content)
	}
	if !strings.Contains(content, "file mode test") {
		t.Errorf("log file missing 'file mode test', got: %s", content)
	}
}

func TestCheckAt_FileMode_MultipleWrites(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "check.log")

	t.Setenv(modeEnv, "file")
	t.Setenv(pathEnv, logPath)

	CheckAt(false, "test.go", 1, "first")
	CheckAt(false, "test.go", 2, "second")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 log lines, got %d: %q", len(lines), string(data))
	}
}

func TestCheckAt_FileMode_NoPath_Panics(t *testing.T) {
	t.Setenv(modeEnv, "file")
	t.Setenv(pathEnv, "")

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic when CHECK_LOG_PATH is empty")
		}
		msg := r.(string)
		if !strings.Contains(msg, "CHECK_LOG_PATH") {
			t.Errorf("panic message should mention CHECK_LOG_PATH, got: %s", msg)
		}
	}()

	CheckAt(false, "test.go", 1, "no path")
}

func TestCheckAt_DefaultMode_IsPanic(t *testing.T) {
	t.Setenv(modeEnv, "")

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic in default mode")
		}
	}()

	CheckAt(false, "test.go", 1)
}

func TestParseMode_ValidInputs(t *testing.T) {
	tests := []struct {
		input string
		want  Mode
	}{
		{"panic", PanicMode},
		{"PANIC", PanicMode},
		{"PanicMode", PanicMode},
		{"0", PanicMode},
		{"console", LogToConsoleMode},
		{"CONSOLE", LogToConsoleMode},
		{"logtoconsole", LogToConsoleMode},
		{"LogToConsoleMode", LogToConsoleMode},
		{"1", LogToConsoleMode},
		{"file", LogToFileMode},
		{"FILE", LogToFileMode},
		{"logtofile", LogToFileMode},
		{"LogToFileMode", LogToFileMode},
		{"2", LogToFileMode},
		{"silent", SilentMode},
		{"SILENT", SilentMode},
		{"SilentMode", SilentMode},
		{"3", SilentMode},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseMode(tt.input)
			if err != nil {
				t.Fatalf("ParseMode(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseMode(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseMode_WithWhitespace(t *testing.T) {
	got, err := ParseMode("  panic  ")
	if err != nil {
		t.Fatalf("ParseMode with whitespace returned error: %v", err)
	}
	if got != PanicMode {
		t.Errorf("expected PanicMode, got %d", got)
	}
}

func TestParseMode_Unknown_ReturnsError(t *testing.T) {
	_, err := ParseMode("unknown_mode")
	if err == nil {
		t.Fatal("expected error for unknown mode")
	}
	if !strings.Contains(err.Error(), "unknown mode") {
		t.Errorf("error should contain 'unknown mode', got: %v", err)
	}
}

func TestParseMode_EmptyString_ReturnsError(t *testing.T) {
	_, err := ParseMode("")
	if err == nil {
		t.Fatal("expected error for empty string")
	}
}

func TestGetCurrentMode_InvalidEnv_FallbackToPanic(t *testing.T) {
	t.Setenv(modeEnv, "garbage_value")
	mode := getCurrentMode()
	if mode != PanicMode {
		t.Errorf("expected PanicMode fallback, got %d", mode)
	}
}

func TestGetCurrentMode_FromEnv(t *testing.T) {
	tests := []struct {
		env  string
		want Mode
	}{
		{"panic", PanicMode},
		{"console", LogToConsoleMode},
		{"file", LogToFileMode},
		{"silent", SilentMode},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			t.Setenv(modeEnv, tt.env)
			got := getCurrentMode()
			if got != tt.want {
				t.Errorf("getCurrentMode() with env=%q: got %d, want %d", tt.env, got, tt.want)
			}
		})
	}
}

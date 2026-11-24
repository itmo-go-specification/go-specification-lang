package dynamicchecks_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDynamicChecksSwitchCaseAssertion(t *testing.T) {
	// Reproduce issue #75 and run the switch with both a matching and a nonmatching value.
	output, err := runInstrumentedProgram(t, `package main

import "fmt"

func check(n int) {
	/*@ assert \true; */
	_ = 0
	switch n {
	case 1:
		/*@ assert \false; */
		return
	}
}

func main() {
	check(0)
	fmt.Println("unselected clause returned")
	check(1)
}
`)
	if !strings.Contains(output, "unselected clause returned") {
		t.Fatalf("switch failed when the case was skipped: %v\n%s", err, output)
	}
	if err == nil || !strings.Contains(output, "check failed") {
		t.Fatalf("expected an assertion failure in the selected case, got %v\n%s", err, output)
	}
}

func TestDynamicChecksSwitchDefaultAssertion(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

import "fmt"

func check(n int) {
	switch n {
	case 1:
		return
	default:
		/*@ assert \false; */
		return
	}
}

func main() {
	check(1)
	fmt.Println("unselected clause returned")
	check(0)
}
`)
	if !strings.Contains(output, "unselected clause returned") {
		t.Fatalf("switch failed when the default was skipped: %v\n%s", err, output)
	}
	if err == nil || !strings.Contains(output, "check failed") {
		t.Fatalf("expected an assertion failure in the default clause, got %v\n%s", err, output)
	}
}

func TestDynamicChecksTypeSwitchCaseAssertion(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

import "fmt"

func check(value any) {
	switch value.(type) {
	case int:
		/*@ assert \false; */
		return
	}
}

func main() {
	check("text")
	fmt.Println("unselected clause returned")
	check(1)
}
`)
	if !strings.Contains(output, "unselected clause returned") {
		t.Fatalf("type switch failed when the case was skipped: %v\n%s", err, output)
	}
	if err == nil || !strings.Contains(output, "check failed") {
		t.Fatalf("expected an assertion failure in the selected type case, got %v\n%s", err, output)
	}
}

func TestDynamicChecksSelectCaseAssertion(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

import "fmt"

func check(ch chan int) {
	select {
	case ch <- 1:
		/*@ assert \false; */
		return
	default:
		return
	}
}

func main() {
	check(nil)
	fmt.Println("unselected clause returned")
	check(make(chan int, 1))
}
`)
	if !strings.Contains(output, "unselected clause returned") {
		t.Fatalf("select failed when the send case was skipped: %v\n%s", err, output)
	}
	if err == nil || !strings.Contains(output, "check failed") {
		t.Fatalf("expected an assertion failure in the selected send case, got %v\n%s", err, output)
	}
}

func runInstrumentedProgram(t *testing.T, source string) (string, error) {
	t.Helper()
	dir := t.TempDir()
	input := filepath.ToSlash(filepath.Join(dir, "input.go"))
	if err := os.WriteFile(input, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	generated, err := runDynamicChecksPlugin(input)
	if err != nil {
		t.Fatal(err)
	}
	program := filepath.Join(dir, "instrumented.go")
	if err := os.WriteFile(program, []byte(generated), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.CommandContext(t.Context(), "go", "run", program)
	cmd.Env = append(cmd.Environ(), "CHECK_MODE=panic")
	output, err := cmd.CombinedOutput()
	return string(output), err
}

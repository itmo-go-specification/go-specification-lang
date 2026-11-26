package dynamicchecks_test

import (
	"strings"
	"testing"
)

func TestAtHereUsesCurrentValueInAssertion(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

func main() {
	x := 7
	/*@ assert \at(x, Here) == x; */
	return
}
`)
	if err != nil {
		t.Fatalf("Here assertion should use the current value: %v\n%s", err, output)
	}
}

func TestAtPreCapturesFunctionEntryValueForAssertion(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

func check(x int) {
	x++
	/*@ assert \at(x, Pre) == 7; */
	return
}

func main() { check(7) }
`)
	if err != nil {
		t.Fatalf("Pre assertion should use the function-entry value: %v\n%s", err, output)
	}
}

func TestAtPreWorksInPostcondition(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

import "fmt"

/*@ ensures \result == \at(x, Pre) + 1; */
func increment(x int) int {
	x++
	return x
}

func main() { fmt.Println(increment(7)) }
`)
	if err != nil {
		t.Fatalf("Pre in a postcondition should use the entry value: %v\n%s", err, output)
	}
	if !strings.Contains(output, "8") {
		t.Fatalf("instrumented function returned unexpected output: %s", output)
	}
}

func TestAtHereStillReportsFailedAssertion(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

func main() {
	x := 7
	/*@ assert \at(x, Here) == 8; */
	return
}
`)
	if err == nil || !strings.Contains(output, "assertion") {
		t.Fatalf("expected Here assertion failure, got %v\n%s", err, output)
	}
}

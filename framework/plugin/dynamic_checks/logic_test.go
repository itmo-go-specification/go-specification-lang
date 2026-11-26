package dynamicchecks_test

import (
	"strings"
	"testing"
)

func TestLogicFunctionIsGeneratedForContract(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

/*@ logic successor(x integer) integer = x + 1; */
/*@ requires successor(1) == 2; */
func check() {}

func main() { check() }
`)
	if err != nil {
		t.Fatalf("generated logic function should satisfy the contract: %v\n%s", err, output)
	}
}

func TestLogicFunctionParticipatesInFailedCheck(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

/*@ logic successor(x integer) integer = x + 1; */
/*@ requires successor(1) == 3; */
func check() {}

func main() { check() }
`)
	if err == nil || !strings.Contains(output, "precondition") {
		t.Fatalf("expected failed check through generated logic function, got %v\n%s", err, output)
	}
}

func TestLogicPredicateIsGenerated(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

/*@ predicate positive(x integer) = x > 0; */
/*@ requires positive(1); */
func check() {}

func main() { check() }
`)
	if err != nil {
		t.Fatalf("generated logic predicate should satisfy the contract: %v\n%s", err, output)
	}
}

func TestLogicOnlyFileDoesNotGetUnusedChecksImport(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

import "fmt"

/*@ logic successor(x integer) integer = x + 1; */
func main() { fmt.Println(successor(1)) }
`)
	if err != nil {
		t.Fatalf("logic-only generated program should compile: %v\n%s", err, output)
	}
	if !strings.Contains(output, "2") {
		t.Fatalf("generated logic function returned unexpected output: %s", output)
	}
}

package dynamicchecks_test

import (
	"strings"
	"testing"
)

func TestNamedBehaviorChecksRequiresWhenAssumptionHolds(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

/*@
requires \true;
behavior positive:
    assumes x > 0;
    requires x < 2;
*/
func check(x int) {}

func main() { check(3) }
`)
	if err == nil || !strings.Contains(output, "precondition") {
		t.Fatalf("expected active behavior precondition failure, got %v\n%s", err, output)
	}
}

func TestNamedBehaviorSkipsRequiresWhenAssumptionIsFalse(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

import "fmt"

/*@
behavior positive:
    assumes x > 0;
    requires x < 2;
*/
func check(x int) {}

func main() {
	check(-1)
	fmt.Println("inactive behavior skipped")
}
`)
	if err != nil {
		t.Fatalf("inactive behavior should not check its requires: %v\n%s", err, output)
	}
	if !strings.Contains(output, "inactive behavior skipped") {
		t.Fatalf("instrumented program did not finish: %s", output)
	}
}

func TestNamedBehaviorCombinesMultipleAssumptions(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

/*@
behavior bounded:
    assumes x > 0;
    assumes x < 10;
    requires \false;
*/
func check(x int) {}

func main() { check(20) }
`)
	if err != nil {
		t.Fatalf("all assumptions must hold before checking requires: %v\n%s", err, output)
	}
}

func TestNamedBehaviorChecksEnsuresOnlyWhenActive(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

import "fmt"

/*@
behavior positive:
    assumes x > 0;
    ensures \result > 0;
*/
func value(x int) int { return -1 }

func main() {
	fmt.Println(value(-1))
	value(1)
}
`)
	if !strings.Contains(output, "-1") {
		t.Fatalf("inactive behavior should allow the first call to return: %v\n%s", err, output)
	}
	if err == nil || !strings.Contains(output, "postcondition") {
		t.Fatalf("expected active behavior postcondition failure, got %v\n%s", err, output)
	}
}

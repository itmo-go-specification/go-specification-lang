package dynamicchecks_test

import (
	"strings"
	"testing"
)

func TestAssignsNothingRejectsGlobalChange(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

var state int

/*@ requires \true; assigns \nothing; */
func change() { state++ }

func main() { change() }
`)
	if err == nil || !strings.Contains(output, "assigns") {
		t.Fatalf("expected assigns violation, got %v\n%s", err, output)
	}
}

func TestAssignsAllowsListedGlobalChange(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

import "fmt"

var state int

/*@ assigns state; */
func change() { state++ }

func main() {
	change()
	fmt.Println(state)
}
`)
	if err != nil {
		t.Fatalf("listed global should be allowed to change: %v\n%s", err, output)
	}
	if !strings.Contains(output, "1") {
		t.Fatalf("instrumented program did not run: %s", output)
	}
}

func TestAssignsRejectsChangeToUnlistedGlobal(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

var allowed int
var protected int

/*@ assigns allowed; */
func change() { protected++ }

func main() { change() }
`)
	if err == nil || !strings.Contains(output, "assigns") {
		t.Fatalf("expected assigns violation for unlisted global, got %v\n%s", err, output)
	}
}

func TestAssignsNothingSupportsSliceReassignment(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

var values []int

/*@ assigns \nothing; */
func change() { values = []int{1} }

func main() { change() }
`)
	if err == nil || !strings.Contains(output, "assigns") {
		t.Fatalf("expected assigns violation for slice reassignment, got %v\n%s", err, output)
	}
}

func TestMissingAssignsClauseDoesNotAddFrameCondition(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

import "fmt"

var state int

/*@ requires \true; */
func change() { state++ }

func main() {
	change()
	fmt.Println(state)
}
`)
	if err != nil {
		t.Fatalf("contract without assigns should keep existing behavior: %v\n%s", err, output)
	}
}

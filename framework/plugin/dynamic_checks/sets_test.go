package dynamicchecks_test

import (
	"strings"
	"testing"
)

func TestFiniteSetMembershipPasses(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

func main() {
	/*@ assert 2 \in {1, 2, 3}; */
	return
}
`)
	if err != nil {
		t.Fatalf("expected finite-set membership to pass: %v\n%s", err, output)
	}
}

func TestFiniteSetMembershipFailureIsReported(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

func main() {
	/*@ assert 4 \in {1, 2, 3}; */
	return
}
`)
	if err == nil || !strings.Contains(output, "assertion") {
		t.Fatalf("expected membership assertion failure, got %v\n%s", err, output)
	}
}

func TestFiniteSetSubsetPasses(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

func main() {
	/*@ assert \subset({1, 2}, {1, 2, 3}); */
	return
}
`)
	if err != nil {
		t.Fatalf("expected finite subset check to pass: %v\n%s", err, output)
	}
}

func TestFiniteSetSubsetFailureIsReported(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

func main() {
	/*@ assert \subset({1, 4}, {1, 2, 3}); */
	return
}
`)
	if err == nil || !strings.Contains(output, "assertion") {
		t.Fatalf("expected subset assertion failure, got %v\n%s", err, output)
	}
}

func TestEmptySetSemantics(t *testing.T) {
	output, err := runInstrumentedProgram(t, `package main

func main() {
	/*@ assert \subset(\empty, {1, 2, 3}); */
	/*@ assert !(1 \in \empty); */
	return
}
`)
	if err != nil {
		t.Fatalf("empty-set identities should hold: %v\n%s", err, output)
	}
}

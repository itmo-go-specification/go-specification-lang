package dynamicchecks_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInclusiveQuantifierDoesNotOverflowMaxInt(t *testing.T) {
	src := `package example

/*@ requires \forall i integer; ((lo <= i) && (i <= hi)) ==> \false; */
func check(lo, hi int) {}
`
	path := filepath.Join(t.TempDir(), "input.go")
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := runDynamicChecksPlugin(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"for i := lo; i <= hi; {", "if i == hi {", "break"} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated code does not contain %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "hi + 1") {
		t.Fatalf("inclusive bound must not compute hi + 1:\n%s", got)
	}
}

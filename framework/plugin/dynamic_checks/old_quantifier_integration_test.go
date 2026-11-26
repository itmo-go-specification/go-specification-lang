package dynamicchecks_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOldInsideQuantifierSnapshotsSlice(t *testing.T) {
	src := `package example

/*@ ensures \forall i integer; ((0 <= i) && (i < len(a))) ==> (a[i] == \old(a[i])); */
func preserve(a []int) { if len(a) > 0 { a[0]++ } }
`
	path := filepath.Join(t.TempDir(), "input.go")
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := runDynamicChecksPlugin(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"__old_1 := append([]int(nil), a...)",
		"a[i] == __old_1[i]",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated code does not contain %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "__old_1 := a[i]") {
		t.Fatalf("snapshot must not reference the quantifier variable before its declaration:\n%s", got)
	}
}

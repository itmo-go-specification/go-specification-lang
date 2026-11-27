package dynamicchecks_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogicalXorGeneratesBooleanInequality(t *testing.T) {
	src := `package example

var p, q = true, false

func check() {
/*@ assert (p ^^ q) == \true; */
	return
}
`
	path := filepath.Join(t.TempDir(), "input.go")
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := runDynamicChecksPlugin(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "(p != q) == true") {
		t.Fatalf("expected logical xor to become !=:\n%s", got)
	}
	if strings.Contains(got, "p ^ q") {
		t.Fatalf("logical xor must not become bitwise xor:\n%s", got)
	}
}

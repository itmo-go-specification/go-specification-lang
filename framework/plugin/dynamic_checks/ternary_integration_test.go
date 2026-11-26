package dynamicchecks_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTernaryTermGeneratesTypedIIFE(t *testing.T) {
	src := `package example

var flag = true

func check() {
/*@ assert (flag ? 1 : 2) == 1; */
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
	for _, want := range []string{"func() int", "if flag", "return 1", "return 2"} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated code does not contain %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "flag ? 1 : 2") {
		t.Fatalf("generated Go must not retain the ternary syntax:\n%s", got)
	}
}

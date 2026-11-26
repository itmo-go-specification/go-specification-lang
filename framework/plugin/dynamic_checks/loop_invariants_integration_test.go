package dynamicchecks_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoopInvariantChecksBeforeAndAfterIteration(t *testing.T) {
	src := `package example

func check() {
	i := 0
	/*@ loop invariant (0 <= i) && (i <= 2); */
	for i < 2 { i++ }
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
	if strings.Count(got, "checks.CheckAt") != 2 {
		t.Fatalf("expected two invariant checks:\n%s", got)
	}
	firstCheck := strings.Index(got, "checks.CheckAt")
	loop := strings.Index(got, "for i < 2")
	lastCheck := strings.LastIndex(got, "checks.CheckAt")
	increment := strings.Index(got, "i++")
	if firstCheck < 0 || !(firstCheck < loop && increment < lastCheck) {
		t.Fatalf("expected checks before the loop and after its body:\n%s", got)
	}
}

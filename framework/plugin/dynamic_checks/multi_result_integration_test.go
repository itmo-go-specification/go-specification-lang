package dynamicchecks_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMultiResultReturnsPreserveGoSemantics(t *testing.T) {
	src := `package example

/*@ ensures \\result_0 == 2 && \\result_1 == 1; */
func Swap() (a, b int) {
	a = 1
	b = 2
	return b, a
}

func pair() (int, int) { return 1, 2 }

/*@ ensures \\result_0 == 1 && \\result_1 == 2; */
func Pair() (a, b int) { return pair() }
`
	path := filepath.Join(t.TempDir(), "input.go")
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := runDynamicChecksPlugin(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"a, b = b, a", "a, b = pair()"} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated code does not contain %q:\n%s", want, got)
		}
	}
}

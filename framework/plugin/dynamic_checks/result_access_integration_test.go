package dynamicchecks_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResultFieldAndIndexAccess(t *testing.T) {
	src := `package example

type Box struct { Value int }

/*@ ensures \\result.Value == 42; */
func MakeBox() (result Box) { return Box{Value: 42} }

/*@ ensures \\result[0] == 42; */
func MakeArray() (result [1]int) { return [1]int{42} }
`
	path := filepath.Join(t.TempDir(), "input.go")
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := runDynamicChecksPlugin(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"result.Value == 42", "result[0] == 42"} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated code does not contain %q:\n%s", want, got)
		}
	}
}

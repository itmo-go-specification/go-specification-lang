package dynamicchecks_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQuantifierUsesParameterIntegerType(t *testing.T) {
	tests := []struct {
		name   string
		goType string
	}{
		{name: "int32", goType: "int32"},
		{name: "int64", goType: "int64"},
		{name: "uint64", goType: "uint64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := `package example

/*@ requires \forall i integer; ((0 <= i) && (i < n)) ==> (i < n); */
func check(n ` + tt.goType + `) {}
`
			path := filepath.Join(t.TempDir(), "input.go")
			if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
				t.Fatal(err)
			}

			got, err := runDynamicChecksPlugin(path)
			if err != nil {
				t.Fatal(err)
			}
			want := "for i := " + tt.goType + "(0); i < n; i++"
			if !strings.Contains(got, want) {
				t.Fatalf("generated code does not contain %q:\n%s", want, got)
			}
		})
	}
}

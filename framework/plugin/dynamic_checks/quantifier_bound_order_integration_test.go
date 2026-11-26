package dynamicchecks_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQuantifierFindsNestedAndReorderedBounds(t *testing.T) {
	tests := []struct {
		name string
		spec string
	}{
		{
			name: "nested",
			spec: `\exists i integer; (0 <= i) && ((i < 3) && (i == 1))`,
		},
		{
			name: "reordered",
			spec: `\exists i integer; (i == 1) && ((3 > i) && (i >= 0))`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := "package example\nfunc check() {\n/*@ assert " + tt.spec + "; */\nreturn\n}\n"
			path := filepath.Join(t.TempDir(), "input.go")
			if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
				t.Fatal(err)
			}

			got, err := runDynamicChecksPlugin(path)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"for i := 0; i < 3; i++", "if i == 1"} {
				if !strings.Contains(got, want) {
					t.Fatalf("generated code does not contain %q:\n%s", want, got)
				}
			}
		})
	}
}

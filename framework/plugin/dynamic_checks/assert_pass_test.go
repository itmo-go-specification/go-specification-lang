package dynamicchecks_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssertInSwitchCase(t *testing.T) {
	src := `package main

func main() {
	/*@ assert \true; */
	_ = 0
	switch 1 {
	case 1:
		/*@ assert \false; */
		return
	}
}
`
	got := runOnSource(t, src)
	if !strings.Contains(got, `checks.CheckAt(false, `) {
		t.Errorf("assert inside switch case was dropped:\n%s", got)
	}
}

func TestAssertInSelectCase(t *testing.T) {
	src := `package main

func main() {
	ch := make(chan int, 1)
	ch <- 1
	select {
	case <-ch:
		/*@ assert \false; */
		return
	}
}
`
	got := runOnSource(t, src)
	if !strings.Contains(got, `checks.CheckAt(false, `) {
		t.Errorf("assert inside select case was dropped:\n%s", got)
	}
}

func runOnSource(t *testing.T, src string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "input.go")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := runDynamicChecksPlugin(path)
	if err != nil {
		t.Fatalf("plugin processing failed: %v", err)
	}
	return got
}

package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFiles(t *testing.T) {
	sourcesDir := "../test/sources"
	goldenDir := "../test/golden"

	files, err := os.ReadDir(sourcesDir)
	if err != nil {
		t.Fatalf("Failed to read sources directory: %v", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			sourcePath := filepath.Join(sourcesDir, file.Name())
			goldenPath := filepath.Join(goldenDir, file.Name()+".tree")

			p := NewParser()
			astData, err := p.ParseFile(sourcePath)
			if err != nil {
				t.Fatalf("Failed to parse file %s: %v", file.Name(), err)
			}

			if astData.File == nil {
				t.Errorf("Parsed file is nil for %s", file.Name())
			}

			if astData.FileSet == nil {
				t.Errorf("FileSet is nil for %s", file.Name())
			}

			if _, err := os.Stat(goldenPath); err == nil {
				goldenContent, err := os.ReadFile(goldenPath)
				if err != nil {
					t.Logf("Warning: Could not read golden file %s: %v", goldenPath, err)
				} else {
					t.Logf("Golden file exists: %s (%d bytes)", goldenPath, len(goldenContent))
				}
			}
		})
	}
}

func TestParseSource(t *testing.T) {
	source := []byte(`package test

func main() {
	println("hello")
}
`)

	p := NewParser()
	astData, err := p.ParseSource(source)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	if astData.File == nil {
		t.Error("Parsed file is nil")
	}

	if astData.File.Name.Name != "test" {
		t.Errorf("Expected package name 'test', got '%s'", astData.File.Name.Name)
	}
}
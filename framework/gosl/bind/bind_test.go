package bind

import (
	"go/ast"
	"gpma/framework/gosl/front"
	"gpma/framework/parser"
	"path/filepath"
	"testing"
)


func TestBindFuncContracts(t *testing.T) {
	sourcesDir := "../../test/notation"

	filename := "count.go"
	t.Run(filename, func (t *testing.T) {
		p := parser.NewParser()
		astData, err := p.ParseFile(filepath.Join(sourcesDir, filename))

		if err != nil {
			t.Fatalf("Failed to parse file %s: %v", filename, err)
		}

		if astData.File == nil {
			t.Errorf("Parsed file is nil for %s", filename)
		}

		if astData.FileSet == nil {
			t.Errorf("FileSet is nil for %s", filename)
		}

		blocks := front.CollectRawBlocks(astData.FileSet, astData.File, astData.FileSet.File(astData.File.Pos()).Name())
		spec, err := front.BuildSpecModule(blocks)
		astData, err = Bind(astData, spec)

		if err != nil {
			t.Fatalf("Failed to bind spec modules in file %s: %v", filename, err)
		}

		ast.Inspect(astData.File, func(n ast.Node) bool {
			if funcDecl, ok := n.(*ast.FuncDecl); ok {
				if astData.NodeInfo[funcDecl] == nil {
					t.Fatalf("Failed to bind spec for func %s", funcDecl.Name)
				}
			}
			return true
		})
	})

	filename = "count_after_var.go"
	t.Run(filename, func (t *testing.T) {
		p := parser.NewParser()
		astData, err := p.ParseFile(filepath.Join(sourcesDir, filename))

		if err != nil {
			t.Fatalf("Failed to parse file %s: %v", filename, err)
		}

		if astData.File == nil {
			t.Errorf("Parsed file is nil for %s", filename)
		}

		if astData.FileSet == nil {
			t.Errorf("FileSet is nil for %s", filename)
		}

		blocks := front.CollectRawBlocks(astData.FileSet, astData.File, astData.FileSet.File(astData.File.Pos()).Name())
		spec, err := front.BuildSpecModule(blocks)
		astData, err = Bind(astData, spec)

		if err == nil {
			t.Fatalf("Binded invalid spec modules in file %s", filename)
		}
	})
}

func TestBindLoopSpecs(t *testing.T) {
	sourcesDir := "../../test/notation"

	filename := "lower_bound.go"
	t.Run(filename, func (t *testing.T) {
		p := parser.NewParser()
		astData, err := p.ParseFile(filepath.Join(sourcesDir, filename))

		if err != nil {
			t.Fatalf("Failed to parse file %s: %v", filename, err)
		}

		if astData.File == nil {
			t.Errorf("Parsed file is nil for %s", filename)
		}

		if astData.FileSet == nil {
			t.Errorf("FileSet is nil for %s", filename)
		}

		blocks := front.CollectRawBlocks(astData.FileSet, astData.File, astData.FileSet.File(astData.File.Pos()).Name())
		spec, err := front.BuildSpecModule(blocks)
		astData, err = Bind(astData, spec)

		if err != nil {
			t.Fatalf("Failed to bind spec modules in file %s: %v", filename, err)
		}

		ast.Inspect(astData.File, func(n ast.Node) bool {
			if forStmt, ok := n.(*ast.ForStmt); ok {
				if astData.NodeInfo[forStmt] == nil {
					t.Fatal("Failed to bind spec for loop at", forStmt.For)
				}
			}
			return true
		})
	})

	filename = "lower_bound_after_var.go"
	t.Run(filename, func (t *testing.T) {
		p := parser.NewParser()
		astData, err := p.ParseFile(filepath.Join(sourcesDir, filename))

		if err != nil {
			t.Fatalf("Failed to parse file %s: %v", filename, err)
		}

		if astData.File == nil {
			t.Errorf("Parsed file is nil for %s", filename)
		}

		if astData.FileSet == nil {
			t.Errorf("FileSet is nil for %s", filename)
		}

		blocks := front.CollectRawBlocks(astData.FileSet, astData.File, astData.FileSet.File(astData.File.Pos()).Name())
		spec, err := front.BuildSpecModule(blocks)
		astData, err = Bind(astData, spec)

		if err == nil {
			t.Fatalf("Binded invalid spec modules in file %s", filename)
		}
	})
}

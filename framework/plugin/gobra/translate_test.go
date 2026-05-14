package gobra

import (
	"cmp"
	"fmt"
	"gpma/framework/data"
	"gpma/framework/gosl/bind"
	"gpma/framework/gosl/front"
	specir "gpma/framework/gosl/ir"
	"gpma/framework/parser"
	"slices"
	"testing"
)

func printSpanMap(m map[int]specir.Span) {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	for _, k := range keys {
		s := m[k]
		fmt.Printf("%d -> [%d:%d] %s\n", k, s.Start, s.End, s.File)
	}
}

func PrintAllNamedPredicates(data *data.ASTData, fileContent string) {
	type entry struct {
		line, col int
		name      string
		predicate specir.Expr
	}

	var entries []entry

	for _, info := range data.NodeInfo {
		if info == nil {
			continue
		}
		if fc := info.FuncContract; fc != nil {
			for _, np := range fc.Requires {
				line, col, _ := OffsetToLineCol(fileContent, np.Span.Start)
				entries = append(entries, entry{line, col, np.Label, np.Expr})
			}
			for _, np := range fc.Ensures {
				line, col, _ := OffsetToLineCol(fileContent, np.Span.Start)
				entries = append(entries, entry{line, col, np.Label, np.Expr})
			}
		}
		if ls := info.LoopSpec; ls != nil {
			for _, np := range ls.Invariants {
				line, col, _ := OffsetToLineCol(fileContent, np.Span.Start)
				entries = append(entries, entry{line, col, np.Label, np.Expr})
			}
		}
		if as := info.Asserts; as != nil {
			line, col, _ := OffsetToLineCol(fileContent, as.Assertion.Span.Start)
			entries = append(entries, entry{line, col, as.Assertion.Label, as.Assertion.Expr})
		}
	}

	slices.SortFunc(entries, func(a, b entry) int {
		if a.line != b.line {
			return cmp.Compare(a.line, b.line)
		}
		return cmp.Compare(a.col, b.col)
	})

	for _, e := range entries {
		fmt.Printf("%d:%d  %-20s  %s\n", e.line, e.col, e.name, PrintExpr(e.predicate))
	}
}

func TestTranslationToGobra(t *testing.T) {
	filepath := "../../samples/projects/expr/expr.go"
	parser := parser.NewParser()

	t.Run("General", func(t *testing.T) {
		astData, err := parser.ParseFile(filepath)
		if err != nil {
			t.Fatalf("failed to parse file: %v", err)
		}

		blocks := front.CollectRawBlocks(astData.FileSet, astData.File, astData.FileSet.File(astData.File.Pos()).Name())
		spec, err := front.BuildSpecModule(blocks)
		astData, err = bind.Bind(astData, spec)

		failed_predicates, err := RunGobraVerifier(astData)
		if err != nil {
			t.Fatalf("failed to run gobra: %v", err)
		}

		if len(failed_predicates) == 0 {
			t.Fatalf("failed assertion not found")
		}

		for _, predicate := range failed_predicates {
			line, col, err := OffsetToLineCol(astData.Source, predicate.Span.Start)
			if err != nil {
				t.Fatalf("failed to get error coords: %v", err)
			}

			fmt.Printf("Error: condition at %d:%d may fail\n", line, col)
			fmt.Printf("\t%s\n", PrintExpr(predicate.Expr))
		}
	})
}

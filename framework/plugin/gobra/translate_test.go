package gobra

import (
	"cmp"
	"fmt"
	"gpma/framework/data"
	gobra_runner "gpma/framework/gobra_runner"
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
				line, col, _ := gobra_runner.OffsetToLineCol([]byte(fileContent), np.Span.Start)
				entries = append(entries, entry{line, col, np.Label, np.Expr})
			}
			for _, np := range fc.Ensures {
				line, col, _ := gobra_runner.OffsetToLineCol([]byte(fileContent), np.Span.Start)
				entries = append(entries, entry{line, col, np.Label, np.Expr})
			}
		}
		if ls := info.LoopSpec; ls != nil {
			for _, np := range ls.Invariants {
				line, col, _ := gobra_runner.OffsetToLineCol([]byte(fileContent), np.Span.Start)
				entries = append(entries, entry{line, col, np.Label, np.Expr})
			}
		}
		for _, as := range info.Asserts {
			line, col, _ := gobra_runner.OffsetToLineCol([]byte(fileContent), as.Assertion.Span.Start)
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

func TestHybridMappingGeneration(t *testing.T) {
	filepath := "../../../samples/projects/expr/expr.go"
	p := parser.NewParser()

	astData, err := p.ParseFile(filepath)
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	blocks := front.CollectRawBlocks(astData.FileSet, astData.File, astData.FileSet.File(astData.File.Pos()).Name())
	spec, _ := front.BuildSpecModule(blocks)
	astData, err = bind.Bind(astData, spec)
	if err != nil {
		t.Fatalf("failed to bind spec: %v", err)
	}

	result, err := GenerateGobra(astData, filepath)
	if err != nil {
		t.Fatalf("failed to generate gobra: %v", err)
	}

	// Check that hybrid mapping is created
	if result.HybridMapping == nil {
		t.Fatal("expected hybrid mapping to be created")
	}

	// Check that source file is set in hybrid mapping
	if result.HybridMapping.SourceFile != filepath {
		t.Errorf("expected source file %s, got %s", filepath, result.HybridMapping.SourceFile)
	}

	// Check that additions are tracked
	if len(result.HybridMapping.Replacements) == 0 {
		t.Fatal("expected additions to be tracked")
	}

	// Check that intervals don't overlap
	for i := 1; i < len(result.HybridMapping.Replacements); i++ {
		prev := result.HybridMapping.Replacements[i-1]
		curr := result.HybridMapping.Replacements[i]

		if prev.GenEnd >= curr.GenStart {
			t.Errorf("intervals overlap: [%d:%d] and [%d:%d]",
				prev.GenStart, prev.GenEnd, curr.GenStart, curr.GenEnd)
		}
	}

	// Check that intervals are valid
	for _, interval := range result.HybridMapping.Replacements {
		if !interval.IsValid() {
			t.Errorf("invalid interval: %+v", interval)
		}
	}

	// Check that predicate spans are set for predicates
	predicateCount := 0
	for _, interval := range result.HybridMapping.Replacements {
		if interval.PredicateSpan != nil {
			predicateCount++
		}
	}

	if predicateCount == 0 {
		t.Fatal("expected some intervals to have predicate spans")
	}
}

func TestMultiplePredicatesSameSourcePos(t *testing.T) {
	// This test verifies that multiple predicates at the same source position
	// (e.g., requires and ensures in the same contract) are handled correctly
	filepath := "../../../samples/projects/expr/expr.go"
	p := parser.NewParser()

	astData, err := p.ParseFile(filepath)
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	blocks := front.CollectRawBlocks(astData.FileSet, astData.File, astData.FileSet.File(astData.File.Pos()).Name())
	spec, _ := front.BuildSpecModule(blocks)
	astData, err = bind.Bind(astData, spec)
	if err != nil {
		t.Fatalf("failed to bind spec: %v", err)
	}

	result, err := GenerateGobra(astData, filepath)
	if err != nil {
		t.Fatalf("failed to generate gobra: %v", err)
	}

	// Group intervals by source position
	sourcePosGroups := make(map[int][]gobra_runner.Interval)
	for _, interval := range result.HybridMapping.Replacements {
		sourcePos := interval.SourcePos.Offset
		sourcePosGroups[sourcePos] = append(sourcePosGroups[sourcePos], interval)
	}

	// Check if any source position has multiple intervals
	foundMultiple := false
	for sourcePos, intervals := range sourcePosGroups {
		if len(intervals) >= 2 {
			foundMultiple = true

			// Verify intervals don't overlap
			for i := 1; i < len(intervals); i++ {
				prev := intervals[i-1]
				curr := intervals[i]

				if prev.GenEnd >= curr.GenStart {
					t.Errorf("intervals at source pos %d overlap: [%d:%d] and [%d:%d]",
						sourcePos, prev.GenStart, prev.GenEnd, curr.GenStart, curr.GenEnd)
				}
			}

			// Verify predicate spans are different
			predicateSpans := make(map[int]bool)
			for _, interval := range intervals {
				if interval.PredicateSpan != nil {
					if predicateSpans[interval.PredicateSpan.Offset] {
						t.Errorf("duplicate predicate span at source pos %d: offset %d",
							sourcePos, interval.PredicateSpan.Offset)
					}
					predicateSpans[interval.PredicateSpan.Offset] = true
				}
			}

			// Verify all intervals have the same source position
			for _, interval := range intervals {
				if interval.SourcePos.Offset != sourcePos {
					t.Errorf("interval has different source pos: expected %d, got %d",
						sourcePos, interval.SourcePos.Offset)
				}
			}
		}
	}

	if !foundMultiple {
		t.Log("Warning: no source position with multiple intervals found - test may not be comprehensive")
	}
}

func TestTranslationToGobra(t *testing.T) {
	filepath := "../../../samples/projects/expr/expr.go"
	p := parser.NewParser()

	astData, err := p.ParseFile(filepath)
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	blocks := front.CollectRawBlocks(astData.FileSet, astData.File, astData.FileSet.File(astData.File.Pos()).Name())
	spec, _ := front.BuildSpecModule(blocks)
	astData, err = bind.Bind(astData, spec)
	if err != nil {
		t.Fatalf("failed to bind spec: %v", err)
	}

	generated, err := GenerateGobra(astData, filepath)
	if err != nil {
		t.Fatalf("failed to generate gobra: %v", err)
	}
	if generated.GobraCode == "" {
		t.Fatal("expected non-empty gobra source")
	}
	if len(generated.HybridMapping.Replacements) == 0 {
		t.Fatal("expected hybrid mapping additions")
	}
}

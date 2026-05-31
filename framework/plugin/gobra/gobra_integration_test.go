//go:build integration

package gobra

import (
	"os"
	"testing"

	"gpma/framework/gosl/bind"
	"gpma/framework/gosl/front"
	"gpma/framework/parser"
)

func TestTranslationToGobraIntegration(t *testing.T) {
	if os.Getenv("GOBRA_JAR") == "" {
		if _, err := os.Stat(defaultGobraJar); err != nil {
			t.Skip("Gobra not installed; set GOBRA_JAR or install gpma deb package")
		}
	}

	filepath := "../../../samples/projects/expr/expr.go"
	p := parser.NewParser()

	t.Run("General", func(t *testing.T) {
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

		cfg := DefaultConfig()
		failed_predicates, err := RunGobraVerifier(astData, cfg)
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

			t.Logf("Error: condition at %d:%d may fail: %s", line, col, PrintExpr(predicate.Expr))
		}
	})
}

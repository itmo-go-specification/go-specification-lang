package gobra

import (
	"fmt"
	"gpma/framework/gosl/bind"
	"gpma/framework/gosl/front"
	"gpma/framework/parser"
	"os"
	"testing"
)

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
		text, err := os.ReadFile(filepath)

		generated, err := GenerateGobra(string(text), astData)
		if err != nil {
			t.Fatalf("failed to generate gobra: %v", err)
		}

		fmt.Print(generated.GetGobraSource())
	})
}
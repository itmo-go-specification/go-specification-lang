package front

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	specir "gpma/framework/gosl/ir"

	"github.com/antlr4-go/antlr/v4"

	acsl "gpma/framework/gosl/front/generated/grammar/acsl"
)

func buildSpecModuleACSL_ANTLR(spec *specir.SpecModule, blocks []RawSpecBlock) (*specir.SpecModule, error) {
	for i, v := range blocks {
		// fmt.Println("i=", i, "block:", v.Text, "Start:", v.Start, "End:", v.End)
		fmt.Println("---------- Build SpecModule ----------")

		is := antlr.NewInputStream(v.Text)
		lexer := acsl.NewACSLLexer(is)
		tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
		parser := acsl.NewACSLParser(tokens)
		parser.BuildParseTrees = true

		tree := parser.Acsl_comment()
		jsonStr, _ := DumpTreeAsJSON(parser, tokens, tree.(antlr.Tree))
		path, _ := writeJSONForBlock("acsl_trees", v, jsonStr, i)
		fmt.Println("ACSL ast wrote:", path)

		builder := newACSLBuilder(spec, v) // v = RawSpecBlock
		if err := builder.BuildFromAcslComment(tree); err != nil {
			fmt.Println("Finshed building spec from block:", i)
		}

		b, _ := json.MarshalIndent(spec, "", "  ")
		path, _ = writeSpecModuleJSON("acsl_trees", string(b), "spec_"+strconv.Itoa(i))
		fmt.Println("Spec wrote: ", path)
	}
	b, _ := json.MarshalIndent(spec, "", "  ")
	_, _ = writeSpecModuleJSON("acsl_trees", string(b), "spec_final")
	return spec, fmt.Errorf("BuildSpecModuleACSL_ANTLR not implemented")
}

func writeSpecModuleJSON(outDir string, jsonStr string, filename string) (string, error) {

	name := filename + ".json"
	path := filepath.Join(outDir, name)

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(jsonStr), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

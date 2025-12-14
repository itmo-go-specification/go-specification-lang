package front

import (
	"errors"
	"fmt"

	specir "gpma/framework/gosl/ir"

	"github.com/antlr4-go/antlr/v4"

	acsl "gpma/framework/gosl/front/generated/grammar/acsl"
)

func buildSpecModuleACSL_ANTLR(spec *specir.SpecModule, blocks []RawSpecBlock) (*specir.SpecModule, error) {
	for _, v := range blocks {

		is := antlr.NewInputStream(v.Text)
		lexer := acsl.NewACSLLexer(is)
		tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
		parser := acsl.NewACSLParser(tokens)

		// cst building. TODO: building c+acsl cst
		parser.BuildParseTrees = true
		// tree := parser.Acsl_comment() // acsl cst
		tree := parser.TranslationUnit() // c cst

		fmt.Println(antlr.TreesStringTree(tree, parser.GetRuleNames(), parser))
	}

	return spec, errors.New("BuildSpecModuleACSL_ANTLR not implemented")
}

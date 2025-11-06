package main

import (
	"fmt"
	"log"
	"os"

	acsl "gpma/parser/generated/grammar"

	"github.com/antlr4-go/antlr/v4"
)

func main() {
	// 1. read file
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <c-file>\n", os.Args[0])
		os.Exit(1)
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// 2. lexer/parser
	is := antlr.NewInputStream(string(data))
	lexer := acsl.NewACSLLexer(is)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := acsl.NewACSLParser(tokens)

	// fmt.Println("Rules:", parser.GetRuleNames())

	// 3. cst building. TODO: building c+acsl cst
	parser.BuildParseTrees = true
	// tree := parser.Acsl_comment() // acsl cst
	tree := parser.TranslationUnit() // c cst

	fmt.Println(antlr.TreesStringTree(tree, parser.GetRuleNames(), parser))
}

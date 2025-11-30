package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"gpma/framework/data"
	"os"
)

type Parser struct {
	fset *token.FileSet
}

func NewParser() *Parser {
	return &Parser{
		fset: token.NewFileSet(),
	}
}

func (p *Parser) ParseFile(filename string) (*data.ASTData, error) {
	file, err := parser.ParseFile(p.fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return data.NewASTData(file, p.fset), nil
}

func (p *Parser) ParseSource(source []byte) (*data.ASTData, error) {
	file, err := parser.ParseFile(p.fset, "", source, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return data.NewASTData(file, p.fset), nil
}

func (p *Parser) ParseFileContent(filename string) (*data.ASTData, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	file, err := parser.ParseFile(p.fset, filename, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return data.NewASTData(file, p.fset), nil
}

func ParseFile(filename string) (*ast.File, *token.FileSet, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	return file, fset, nil
}

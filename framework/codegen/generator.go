package codegen

import (
	"bytes"
	"go/format"
	"go/printer"
	"gpma/framework/data"
)

type Generator struct {
	astData *data.ASTData
}

func NewGenerator(astData *data.ASTData) *Generator {
	return &Generator{
		astData: astData,
	}
}

func (g *Generator) Generate() ([]byte, error) {
	var buf bytes.Buffer

	err := printer.Fprint(&buf, g.astData.FileSet, g.astData.File)
	if err != nil {
		return nil, err
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes(), nil
	}

	return formatted, nil
}

func (g *Generator) GenerateToString() (string, error) {
	code, err := g.Generate()
	if err != nil {
		return "", err
	}
	return string(code), nil
}

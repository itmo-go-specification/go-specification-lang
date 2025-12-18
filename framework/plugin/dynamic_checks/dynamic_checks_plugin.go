package dynamicchecks

import (
	"go/ast"
	"go/token"
	"gpma/framework/data"
	"gpma/framework/plugin"
)

func init() {
	plugin.Register("dynamic_checks", func() plugin.Plugin {
		return NewDynamicChecksPlugin()
	})
}

type DynamicChecksPlugin struct{}

func NewDynamicChecksPlugin() *DynamicChecksPlugin {
	return &DynamicChecksPlugin{}
}

func (p *DynamicChecksPlugin) Name() string {
	return "DynamicChecksPlugin"
}

func (p *DynamicChecksPlugin) Process(astData *data.ASTData) *plugin.PluginResult {
	result := plugin.NewPluginResult()

	addImport(astData.File)

	processFunctions(astData, result)
	processLoops(astData, result)

	return result
}

func addImport(file *ast.File) {
	importPath := "gpma/framework/plugin/dynamic_checks/check"

	for _, imp := range file.Imports {
		if imp.Path.Value == "\""+importPath+"\"" {
			return
		}
	}

	file.Comments = nil

	newImport := &ast.ImportSpec{
		Name: &ast.Ident{
			Name: "checks",
		},
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: "\"" + importPath + "\"",
		},
	}

	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			genDecl.Specs = append(genDecl.Specs, newImport)
			file.Imports = append(file.Imports, newImport)
			return
		}
	}

	importDecl := &ast.GenDecl{
		Tok: token.IMPORT,
		Specs: []ast.Spec{
			newImport,
		},
	}

	file.Decls = append([]ast.Decl{importDecl}, file.Decls...)
	file.Imports = append(file.Imports, newImport)
}

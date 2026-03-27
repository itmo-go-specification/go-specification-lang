package dynamicchecks

import (
	"errors"
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

type DynamicChecksPlugin struct {
	strictMode bool
}

func NewDynamicChecksPlugin() *DynamicChecksPlugin {
	return &DynamicChecksPlugin{
		strictMode: true,
	}
}

func (p *DynamicChecksPlugin) Name() string {
	return "DynamicChecksPlugin"
}

func (p *DynamicChecksPlugin) Configure(params map[string]interface{}) error {
	if val, ok := params["strict_mode"]; ok {
		if strict, ok := val.(bool); ok {
			p.strictMode = strict
		} else if strVal, ok := val.(string); ok {
			switch strVal {
			case "true", "strict", "yes":
				p.strictMode = true
			case "false", "lenient", "no":
				p.strictMode = false
			default:
				return errors.New("strict_mode: must be boolean or one of: true, false, strict, lenient, yes, no")
			}
		} else {
			return errors.New("strict_mode: must be boolean or string")
		}
	}
	return nil
}

func (p *DynamicChecksPlugin) Process(astData *data.ASTData) *plugin.PluginResult {
	result := plugin.NewPluginResult()

	addImport(astData.File)

	if applyAssertPass(astData) {
		result.WithASTModified().WithMapModified()
	}
	if applyContractPass(astData, p.strictMode) {
		result.WithASTModified().WithMapModified()
	}
	if applyLoopPass(astData) {
		result.WithASTModified().WithMapModified()
	}

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
		Tok:   token.IMPORT,
		Specs: []ast.Spec{newImport},
	}

	file.Decls = append([]ast.Decl{importDecl}, file.Decls...)
	file.Imports = append(file.Imports, newImport)
}

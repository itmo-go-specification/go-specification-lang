package examples

import (
	"go/ast"
	"go/token"

	"gpma/framework/data"
	"gpma/framework/plugin"
)

func init() {
	plugin.Register("add_trace", func() plugin.Plugin {
		return NewAddTracePlugin()
	})
}

type AddTracePlugin struct{}

func NewAddTracePlugin() *AddTracePlugin {
	return &AddTracePlugin{}
}

func (p *AddTracePlugin) Name() string {
	return "AddTracePlugin"
}

func (p *AddTracePlugin) Process(astData *data.ASTData) *plugin.PluginResult {
	result := plugin.NewPluginResult()
	modified := false

	needsFmtImport := false

	ast.Inspect(astData.File, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if funcDecl.Body == nil || len(funcDecl.Body.List) == 0 {
			return true
		}

		funcName := funcDecl.Name.Name
		receiverName := ""
		if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
			if len(funcDecl.Recv.List[0].Names) > 0 {
				receiverName = funcDecl.Recv.List[0].Names[0].Name
			}
			typeName := getTypeName(funcDecl.Recv.List[0].Type)
			funcName = typeName + "." + funcName
		}

		traceStmt := &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "fmt"},
					Sel: &ast.Ident{Name: "Printf"},
				},
				Args: []ast.Expr{
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: `"[TRACE] Entering function: ` + funcName + `\n"`,
					},
				},
			},
		}

		newBody := make([]ast.Stmt, 0, len(funcDecl.Body.List)+1)
		newBody = append(newBody, traceStmt)
		newBody = append(newBody, funcDecl.Body.List...)
		funcDecl.Body.List = newBody

		needsFmtImport = true
		modified = true

		info := astData.GetNodeInfo(funcDecl)
		info.Data["has_trace"] = true
		info.Data["function_name"] = funcName
		if receiverName != "" {
			info.Data["receiver_name"] = receiverName
		}

		return true
	})

	if needsFmtImport {
		ensureFmtImport(astData.File)
	}

	if modified {
		result.WithASTModified().WithMapModified()
	}

	return result
}

func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getTypeName(t.X)
	case *ast.SelectorExpr:
		return getTypeName(t.X) + "." + t.Sel.Name
	default:
		return "Unknown"
	}
}

func ensureFmtImport(file *ast.File) {
	for _, imp := range file.Imports {
		if imp.Path.Value == `"fmt"` {
			return
		}
	}

	fmtImport := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: `"fmt"`,
		},
	}

	if len(file.Decls) > 0 {
		if genDecl, ok := file.Decls[0].(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			genDecl.Specs = append(genDecl.Specs, fmtImport)
			return
		}
	}

	importDecl := &ast.GenDecl{
		Tok: token.IMPORT,
		Specs: []ast.Spec{
			fmtImport,
		},
	}

	newDecls := make([]ast.Decl, 0, len(file.Decls)+1)
	newDecls = append(newDecls, importDecl)
	newDecls = append(newDecls, file.Decls...)
	file.Decls = newDecls
}
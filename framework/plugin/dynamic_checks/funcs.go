package dynamicchecks

import (
	"fmt"
	"go/ast"
	"go/token"

	"gpma/framework/data"
	specir "gpma/framework/gosl/ir"
	"gpma/framework/plugin"
)

func processFunctions(astData *data.ASTData, result *plugin.PluginResult) {
	ast.Inspect(astData.File, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			info := astData.TryGetNodeInfo(node)
			if info != nil && info.FuncContract != nil {
				if processFunc(node, info.FuncContract) {
					result.WithASTModified()
				}
				astData.GetNodeInfo(node).Data["processed"] = true
				result.WithMapModified()
			}
		}
		return true
	})
}

func processFunc(funcDecl *ast.FuncDecl, contract *specir.FuncContract) bool {
	if funcDecl.Body == nil {
		return false
	}

	converter := NewConverter()

	// Собираем \old выражения из постусловий
	for _, ens := range contract.Ensures {
		converter.CollectOld(ens.Expr)
	}

	var newStmts []ast.Stmt

	// Добавляем проверки requires
	for _, req := range contract.Requires {
		condition := converter.Convert(req.Expr)
		checkCall := createCheckCall(condition, "requires", req.Label)
		newStmts = append(newStmts, &ast.ExprStmt{X: checkCall})
	}

	// Объявление переменных для \old
	for _, oldInfo := range converter.GetOldVars() {
		value := converter.Convert(oldInfo.expr)
		assign := &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: oldInfo.name}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{value},
		}
		newStmts = append(newStmts, assign)
		oldInfo.isDeclared = true
	}

	// Добавляем проверки ensures (defer)
	for _, ens := range contract.Ensures {
		condition := converter.Convert(ens.Expr)
		checkCall := createCheckCall(condition, "ensures", ens.Label)
		deferCheck := &ast.DeferStmt{Call: checkCall}
		newStmts = append(newStmts, deferCheck)
	}

	if len(newStmts) > 0 {
		funcDecl.Body.List = append(newStmts, funcDecl.Body.List...)
		return true
	}
	return false
}

func createCheckCall(condition ast.Expr, prefix, label string) *ast.CallExpr {
	var message string
	if label != "" {
		message = fmt.Sprintf(`"%s: %s"`, prefix, label)
	} else {
		message = fmt.Sprintf(`"%s"`, prefix)
	}

	return &ast.CallExpr{
		Fun: &ast.Ident{Name: "check"},
		Args: []ast.Expr{
			condition,
			&ast.BasicLit{
				Kind:  token.STRING,
				Value: message,
			},
		},
	}
}

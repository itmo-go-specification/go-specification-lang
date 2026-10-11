package dynamicchecks

import (
	"go/ast"
	"go/token"

	"gpma/framework/data"
	specir "gpma/framework/gosl/ir"
)

func applyContractPass(astData *data.ASTData, strictMode bool) bool {
	modified := false

	ast.Inspect(astData.File, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			return true
		}

		info := astData.TryGetNodeInfo(funcDecl)
		if info == nil || info.FuncContract == nil {
			return true
		}

		if applyContractToFunc(funcDecl, info.FuncContract, strictMode) {
			modified = true
		}

		return true
	})

	return modified
}

func applyContractToFunc(funcDecl *ast.FuncDecl, contract *specir.FuncContract, strictMode bool) bool {
	conv := newContractConverter()

	for _, ens := range contract.Ensures {
		conv.CollectOld(ens.Expr)
	}

	var newStmts []ast.Stmt

	hasRequires := len(contract.Requires) > 0

	var preconditionsPassedVar *ast.Ident
	if !strictMode && hasRequires {
		preconditionsPassedVar = &ast.Ident{Name: "__preconditions_passed"}
		newStmts = append(newStmts, &ast.AssignStmt{
			Lhs: []ast.Expr{preconditionsPassedVar},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.Ident{Name: "true"}},
		})
	}

	for _, req := range contract.Requires {
		condition := conv.Convert(req.Expr)

		if strictMode {
			newStmts = append(newStmts, &ast.ExprStmt{
				X: createCheckCall(condition, "precondition", req.Label, req.StartPos),
			})
		} else {
			// if !condition { __preconditions_passed = false }
			newStmts = append(newStmts, &ast.IfStmt{
				Cond: &ast.UnaryExpr{Op: token.NOT, X: condition},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.AssignStmt{
							Lhs: []ast.Expr{preconditionsPassedVar},
							Tok: token.ASSIGN,
							Rhs: []ast.Expr{&ast.Ident{Name: "false"}},
						},
					},
				},
			})
		}
	}

	for _, oldInfo := range conv.GetOldVars() {
		newStmts = append(newStmts, &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: oldInfo.name}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{conv.Convert(oldInfo.expr)},
		})
		oldInfo.isDeclared = true
	}

	// Determine return value metadata.
	// We inject result variables only when there are ensures clauses.
	hasEnsures := len(contract.Ensures) > 0
	var resultInfos []resultInfo
	if hasEnsures && funcDecl.Type.Results != nil {
		resultInfos = collectResultInfos(funcDecl)
	}
	hasResultVars := len(resultInfos) > 0

	// Let the converter resolve \result / \result_N to the correct variable names
	// (named return values use their source names; unnamed use __result_N).
	if hasResultVars {
		conv.SetResultInfos(resultInfos)
	}

	if hasResultVars {
		newStmts = append(newStmts, createResultDeclarations(resultInfos)...)
	}

	for _, ens := range contract.Ensures {
		condition := conv.Convert(ens.Expr)
		checkCall := createCheckCall(condition, "postcondition", ens.Label, ens.StartPos)

		// Always wrap in anonymous function so condition is evaluated at defer execution time
		deferFunc := &ast.FuncLit{
			Type: &ast.FuncType{Params: &ast.FieldList{}},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{X: checkCall},
				},
			},
		}

		if strictMode || !hasRequires {
			// defer func() { checks.CheckAt(...) }()
			newStmts = append(newStmts, &ast.DeferStmt{
				Call: &ast.CallExpr{Fun: deferFunc},
			})
		} else {
			// defer func() { if __preconditions_passed { checks.CheckAt(...) } }()
			deferFunc.Body.List = []ast.Stmt{
				&ast.IfStmt{
					Cond: preconditionsPassedVar,
					Body: &ast.BlockStmt{
						List: []ast.Stmt{&ast.ExprStmt{X: checkCall}},
					},
				},
			}
			newStmts = append(newStmts, &ast.DeferStmt{
				Call: &ast.CallExpr{Fun: deferFunc},
			})
		}
	}

	if len(newStmts) == 0 {
		return false
	}

	funcDecl.Body.List = append(newStmts, funcDecl.Body.List...)

	if hasResultVars {
		replaceReturnsInBlock(funcDecl.Body, resultInfos)
	}

	return true
}

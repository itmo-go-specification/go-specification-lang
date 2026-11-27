package dynamicchecks

import (
	"fmt"
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
	for _, behavior := range contract.Behaviors {
		for _, ens := range behavior.Ensures {
			conv.CollectOld(ens.Expr)
		}
	}

	var newStmts []ast.Stmt

	hasRequires := len(contract.Requires) > 0
	hasEnsures := len(contract.Ensures) > 0
	for _, behavior := range contract.Behaviors {
		hasRequires = hasRequires || len(behavior.Requires) > 0
		hasEnsures = hasEnsures || len(behavior.Ensures) > 0
	}

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

	behaviorConditions := make([]ast.Expr, len(contract.Behaviors))
	for i, behavior := range contract.Behaviors {
		condition := behaviorCondition(conv, behavior.Assumes)
		if len(behavior.Ensures) > 0 {
			active := &ast.Ident{Name: fmt.Sprintf("__behavior_%s_%d_active", behavior.Name, i)}
			newStmts = append(newStmts, &ast.AssignStmt{
				Lhs: []ast.Expr{active},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{condition},
			})
			condition = active
		}
		behaviorConditions[i] = condition

		if len(behavior.Requires) > 0 {
			newStmts = append(newStmts, &ast.IfStmt{
				Cond: condition,
				Body: &ast.BlockStmt{List: preconditionStatements(conv, behavior.Requires, strictMode, preconditionsPassedVar)},
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

	for i, behavior := range contract.Behaviors {
		for _, ens := range behavior.Ensures {
			condition := conv.Convert(ens.Expr)
			checkCall := createCheckCall(condition, "postcondition", ens.Label, ens.StartPos)
			guard := behaviorConditions[i]
			if !strictMode && hasRequires {
				guard = &ast.BinaryExpr{X: guard, Op: token.LAND, Y: preconditionsPassedVar}
			}
			newStmts = append(newStmts, guardedDefer(checkCall, guard))
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

func behaviorCondition(conv exprConverter, assumes []specir.NamedPredicate) ast.Expr {
	if len(assumes) == 0 {
		return &ast.Ident{Name: "true"}
	}

	condition := conv.Convert(assumes[0].Expr)
	for _, assume := range assumes[1:] {
		condition = &ast.BinaryExpr{
			X:  condition,
			Op: token.LAND,
			Y:  conv.Convert(assume.Expr),
		}
	}
	return condition
}

func preconditionStatements(conv exprConverter, requires []specir.NamedPredicate, strictMode bool, passed *ast.Ident) []ast.Stmt {
	stmts := make([]ast.Stmt, 0, len(requires))
	for _, req := range requires {
		condition := conv.Convert(req.Expr)
		if strictMode {
			stmts = append(stmts, &ast.ExprStmt{
				X: createCheckCall(condition, "precondition", req.Label, req.StartPos),
			})
			continue
		}
		stmts = append(stmts, &ast.IfStmt{
			Cond: &ast.UnaryExpr{Op: token.NOT, X: condition},
			Body: &ast.BlockStmt{List: []ast.Stmt{&ast.AssignStmt{
				Lhs: []ast.Expr{passed},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{&ast.Ident{Name: "false"}},
			}}},
		})
	}
	return stmts
}

func guardedDefer(checkCall *ast.CallExpr, guard ast.Expr) *ast.DeferStmt {
	return &ast.DeferStmt{Call: &ast.CallExpr{Fun: &ast.FuncLit{
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{List: []ast.Stmt{&ast.IfStmt{
			Cond: guard,
			Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{X: checkCall}}},
		}}},
	}}}
}

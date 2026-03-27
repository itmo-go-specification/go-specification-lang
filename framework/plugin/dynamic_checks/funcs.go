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
				X: createCheckCall(condition, "precondition", req.Label),
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

	hasEnsures := len(contract.Ensures) > 0
	hasResultVar := hasEnsures &&
		funcDecl.Type.Results != nil &&
		len(funcDecl.Type.Results.List) == 1

	if hasResultVar {
		newStmts = append(newStmts, createResultDeclaration(funcDecl))
	}

	for _, ens := range contract.Ensures {
		condition := conv.Convert(ens.Expr)
		checkCall := createCheckCall(condition, "postcondition", ens.Label)

		if strictMode || !hasRequires {
			newStmts = append(newStmts, &ast.DeferStmt{Call: checkCall})
		} else {
			// defer func() { if __preconditions_passed { checks.Check(...) } }()
			deferFunc := &ast.FuncLit{
				Type: &ast.FuncType{Params: &ast.FieldList{}},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.IfStmt{
							Cond: preconditionsPassedVar,
							Body: &ast.BlockStmt{
								List: []ast.Stmt{&ast.ExprStmt{X: checkCall}},
							},
						},
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

	if hasResultVar {
		replaceReturnsInBlock(funcDecl.Body)
	}

	return true
}

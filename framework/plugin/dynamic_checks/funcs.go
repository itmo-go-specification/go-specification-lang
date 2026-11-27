package dynamicchecks

import (
	"go/ast"
	"go/token"

	"gpma/framework/data"
	specir "gpma/framework/gosl/ir"
)

func applyContractPass(astData *data.ASTData, strictMode bool) bool {
	modified := false
	packageVars := collectPackageVariables(astData.File)

	ast.Inspect(astData.File, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			return true
		}

		info := astData.TryGetNodeInfo(funcDecl)
		if info == nil || info.FuncContract == nil {
			return true
		}

		if applyContractToFunc(funcDecl, info.FuncContract, packageVars, strictMode) {
			modified = true
		}

		return true
	})

	return modified
}

func applyContractToFunc(funcDecl *ast.FuncDecl, contract *specir.FuncContract, packageVars []string, strictMode bool) bool {
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

	newStmts = append(newStmts, createAssignsChecks(contract, packageVars, preconditionsPassedVar)...)

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

func collectPackageVariables(file *ast.File) []string {
	var names []string
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range valueSpec.Names {
				if name.Name != "_" {
					names = append(names, name.Name)
				}
			}
		}
	}
	return names
}

func createAssignsChecks(contract *specir.FuncContract, packageVars []string, guard *ast.Ident) []ast.Stmt {
	if !contract.Assigns.Specified {
		return nil
	}

	allowed := make(map[string]struct{}, len(contract.Assigns.Vars))
	if contract.Assigns.Kind == specir.LocSetVars {
		for _, name := range contract.Assigns.Vars {
			allowed[name] = struct{}{}
		}
	}

	var stmts []ast.Stmt
	for _, name := range packageVars {
		if _, ok := allowed[name]; ok {
			continue
		}

		snapshotName := "__frame_" + name
		stmts = append(stmts, &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: snapshotName}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.Ident{Name: name}},
		})

		condition := &ast.CallExpr{
			Fun: &ast.Ident{Name: "checks.Equal"},
			Args: []ast.Expr{
				&ast.Ident{Name: name},
				&ast.Ident{Name: snapshotName},
			},
		}
		check := ast.Stmt(&ast.ExprStmt{
			X: createCheckCall(condition, "assigns", "", contract.StartPos),
		})
		if guard != nil {
			check = &ast.IfStmt{
				Cond: guard,
				Body: &ast.BlockStmt{List: []ast.Stmt{check}},
			}
		}

		stmts = append(stmts, &ast.DeferStmt{Call: &ast.CallExpr{
			Fun: &ast.FuncLit{
				Type: &ast.FuncType{Params: &ast.FieldList{}},
				Body: &ast.BlockStmt{List: []ast.Stmt{check}},
			},
		}})
	}

	return stmts
}

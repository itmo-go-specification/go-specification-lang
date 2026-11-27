package dynamicchecks

import (
	"go/ast"
	"go/token"

	"gpma/framework/data"
)

func applyAssertPass(astData *data.ASTData) bool {
	modified := false

	ast.Inspect(astData.File, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			return true
		}
		if applyAssertsToFunc(astData, funcDecl) {
			modified = true
		}
		return false
	})

	return modified
}

func applyAssertsToFunc(astData *data.ASTData, funcDecl *ast.FuncDecl) bool {
	conv := newContractConverter()
	hasAsserts := false

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		if info := astData.TryGetNodeInfo(n); info != nil {
			for _, assert := range info.Asserts {
				hasAsserts = true
				conv.CollectOld(assert.Assertion.Expr)
			}
		}
		return true
	})
	if !hasAsserts {
		return false
	}

	var snapshots []ast.Stmt
	for _, oldInfo := range conv.GetOldVars() {
		snapshots = append(snapshots, &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: oldInfo.name}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{conv.Convert(oldInfo.expr)},
		})
		oldInfo.isDeclared = true
	}
	funcDecl.Body.List = append(snapshots, funcDecl.Body.List...)

	insertAsserts := func(list []ast.Stmt) []ast.Stmt {
		newList := make([]ast.Stmt, 0, len(list))
		for _, stmt := range list {
			if info := astData.TryGetNodeInfo(stmt); info != nil {
				for _, assert := range info.Asserts {
					condition := conv.Convert(assert.Assertion.Expr)
					checkStmt := &ast.ExprStmt{X: createCheckCall(condition, "assertion", assert.Assertion.Label, assert.StartPos)}
					newList = append(newList, checkStmt)
				}
			}
			newList = append(newList, stmt)
		}
		return newList
	}

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.BlockStmt:
			stmt.List = insertAsserts(stmt.List)
		case *ast.CaseClause:
			stmt.Body = insertAsserts(stmt.Body)
		case *ast.CommClause:
			stmt.Body = insertAsserts(stmt.Body)
		}
		return true
	})

	return true
}

package dynamicchecks

import (
	"go/ast"

	"gpma/framework/data"
)

func applyAssertPass(astData *data.ASTData) bool {
	modified := false
	conv := &baseConverter{}

	ast.Inspect(astData.File, func(n ast.Node) bool {
		changed := false
		switch s := n.(type) {
		case *ast.BlockStmt:
			s.List, changed = insertAsserts(astData, conv, s.List)
		case *ast.CaseClause:
			s.Body, changed = insertAsserts(astData, conv, s.Body)
		case *ast.CommClause:
			s.Body, changed = insertAsserts(astData, conv, s.Body)
		}

		if changed {
			modified = true
		}

		return true
	})

	return modified
}

func insertAsserts(astData *data.ASTData, conv exprConverter, stmts []ast.Stmt) ([]ast.Stmt, bool) {
	modified := false
	newList := make([]ast.Stmt, 0, len(stmts))
	for _, stmt := range stmts {
		if info := astData.TryGetNodeInfo(stmt); info != nil {
			for _, assert := range info.Asserts {
				condition := conv.Convert(assert.Assertion.Expr)
				checkStmt := &ast.ExprStmt{X: createCheckCall(condition, "assertion", assert.Assertion.Label, assert.StartPos)}
				newList = append(newList, checkStmt)
				modified = true
			}
		}
		newList = append(newList, stmt)
	}
	return newList, modified
}

package dynamicchecks

import (
	"go/ast"

	"gpma/framework/data"
)

func applyAssertPass(astData *data.ASTData) bool {
	modified := false
	conv := &baseConverter{}

	insertAsserts := func(list []ast.Stmt) []ast.Stmt {
		newList := make([]ast.Stmt, 0, len(list))
		for _, stmt := range list {
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
		return newList
	}

	
	ast.Inspect(astData.File, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.BlockStmt:
			s.List = insertAsserts(s.List)
		case *ast.CaseClause:
			s.Body = insertAsserts(s.Body)
		case *ast.CommClause:
			s.Body = insertAsserts(s.Body)
		}
		return true
	})

	return modified
}

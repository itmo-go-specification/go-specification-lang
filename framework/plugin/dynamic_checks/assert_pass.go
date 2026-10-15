package dynamicchecks

import (
	"go/ast"

	"gpma/framework/data"
)

func applyAssertPass(astData *data.ASTData) bool {
	modified := false
	conv := &baseConverter{}

	ast.Inspect(astData.File, func(n ast.Node) bool {
		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		newList := make([]ast.Stmt, 0, len(block.List))
		for _, stmt := range block.List {
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
		block.List = newList

		return true
	})

	return modified
}

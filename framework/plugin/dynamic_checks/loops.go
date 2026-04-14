package dynamicchecks

import (
	"go/ast"
	"go/token"

	"gpma/framework/data"
	specir "gpma/framework/gosl/ir"
)

func applyLoopPass(astData *data.ASTData) bool {
	modified := false

	ast.Inspect(astData.File, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.ForStmt:
			info := astData.TryGetNodeInfo(node)
			if info != nil && info.LoopSpec != nil {
				if applyLoopSpecToFor(node, info.LoopSpec) {
					modified = true
				}
			}
		case *ast.RangeStmt:
			info := astData.TryGetNodeInfo(node)
			if info != nil && info.LoopSpec != nil {
				if applyLoopSpecToRange(node, info.LoopSpec) {
					modified = true
				}
			}
		}
		return true
	})

	return modified
}

func applyLoopSpecToFor(forStmt *ast.ForStmt, spec *specir.LoopSpec) bool {
	if forStmt.Body == nil {
		return false
	}

	conv := newContractConverter()
	checks := generateLoopChecks(spec, conv)

	newBody, modified := insertLoopChecks(forStmt.Body, checks)
	forStmt.Body = &ast.BlockStmt{List: newBody}

	return modified
}

func applyLoopSpecToRange(rangeStmt *ast.RangeStmt, spec *specir.LoopSpec) bool {
	if rangeStmt.Body == nil {
		return false
	}

	conv := newContractConverter()
	checks := generateLoopChecks(spec, conv)

	newBody, modified := insertLoopChecks(rangeStmt.Body, checks)
	rangeStmt.Body = &ast.BlockStmt{List: newBody}

	return modified
}

type loopChecks struct {
	beginStmts []ast.Stmt
	endStmts   []ast.Stmt
}

func generateLoopChecks(spec *specir.LoopSpec, conv *contractConverter) *loopChecks {
	result := &loopChecks{}

	for _, inv := range spec.Invariants {
		condition := conv.Convert(inv.Expr)
		checkCall := createCheckCall(condition, "loop invariant", inv.Label, spec.StartPos)
		result.beginStmts = append(result.beginStmts, &ast.ExprStmt{X: checkCall})
	}

	for _, inv := range spec.Invariants {
		condition := conv.Convert(inv.Expr)
		checkCall := createCheckCall(condition, "loop invariant", inv.Label, spec.StartPos)
		result.endStmts = append(result.endStmts, &ast.ExprStmt{X: checkCall})
	}

	if spec.Variant != nil {
		variantExpr := conv.Convert(spec.Variant)

		beforeVarName := "__variant_before_" + conv.newVar()
		result.beginStmts = append(result.beginStmts, &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: beforeVarName}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{variantExpr},
		})

		afterVarName := "__variant_after_" + conv.newVar()
		result.endStmts = append(result.endStmts,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: afterVarName}},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{conv.Convert(spec.Variant)},
			},
			&ast.ExprStmt{
				X: createCheckCall(
					&ast.BinaryExpr{
						X:  &ast.Ident{Name: afterVarName},
						Op: token.LSS,
						Y:  &ast.Ident{Name: beforeVarName},
					},
					"loop variant",
					"",
					spec.StartPos,
				),
			},
		)
	}

	return result
}

func insertLoopChecks(body *ast.BlockStmt, checks *loopChecks) (newList []ast.Stmt, modified bool) {
	panic("NOT IMPLEMENTED")
}

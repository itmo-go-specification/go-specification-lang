package dynamicchecks

import (
	"go/ast"
	"go/token"

	"gpma/framework/data"
	specir "gpma/framework/gosl/ir"
	"gpma/framework/plugin"
)

type LoopChecks struct {
	BeginStmts []ast.Stmt
	EndStmts   []ast.Stmt
}

func processLoops(astData *data.ASTData, result *plugin.PluginResult) {
	ast.Inspect(astData.File, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.ForStmt: // while-style или C-style
			info := astData.TryGetNodeInfo(node)
			if info != nil && info.LoopSpec != nil {
				if processForLoop(node, info.LoopSpec) {
					result.WithASTModified()
				}
				astData.GetNodeInfo(node).Data["processed"] = true
				result.WithMapModified()
			}

		case *ast.RangeStmt:
			info := astData.TryGetNodeInfo(node)
			if info != nil && info.LoopSpec != nil {
				if processRangeLoop(node, info.LoopSpec) {
					result.WithASTModified()
				}
				astData.GetNodeInfo(node).Data["processed"] = true
				result.WithMapModified()
			}
		}
		return true
	})
}

func processForLoop(forStmt *ast.ForStmt, spec *specir.LoopSpec) bool {
	if forStmt.Body == nil {
		return false
	}

	converter := NewConverter()
	checks := generateLoopChecks(spec, converter)

	newBody, modified := insertLoopChecks(forStmt.Body, checks)
	forStmt.Body = &ast.BlockStmt{List: newBody}

	return modified
}

func processRangeLoop(rangeStmt *ast.RangeStmt, spec *specir.LoopSpec) bool {
	if rangeStmt.Body == nil {
		return false
	}

	converter := NewConverter()
	checks := generateLoopChecks(spec, converter)

	newBody, modified := insertLoopChecks(rangeStmt.Body, checks)
	rangeStmt.Body = &ast.BlockStmt{List: newBody}

	return modified
}

func generateLoopChecks(spec *specir.LoopSpec, converter *Converter) *LoopChecks {
	checks := &LoopChecks{}

	for _, inv := range spec.Invariants {
		condition := converter.Convert(inv.Expr)
		checkCall := createCheckCall(condition, "loop invariant", inv.Label)
		checks.BeginStmts = append(checks.BeginStmts, &ast.ExprStmt{X: checkCall})
	}

	for _, inv := range spec.Invariants {
		condition := converter.Convert(inv.Expr)
		checkCall := createCheckCall(condition, "loop invariant", inv.Label)
		checks.EndStmts = append(checks.EndStmts, &ast.ExprStmt{X: checkCall})
	}

	if spec.Variant != nil {
		variantExpr := converter.Convert(spec.Variant)

		beforeVarName := "__variant_before_" + converter.NewVar()
		checks.BeginStmts = append(checks.BeginStmts, &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: beforeVarName}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{variantExpr},
		})

		afterVarName := "__variant_after_" + converter.NewVar()
		checks.EndStmts = append(checks.EndStmts,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: afterVarName}},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{converter.Convert(spec.Variant)},
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
				),
			},
		)
	}

	return checks
}

func insertLoopChecks(body *ast.BlockStmt, checks *LoopChecks) (newList []ast.Stmt, modified bool) {
	panic("NOT IMPLEMENTED")
	return newList, modified
}

package dynamicchecks

import (
	"fmt"
	"go/ast"
	"go/token"

	specir "gpma/framework/gosl/ir"
)

// convertQuantifier dispatches to convertForall or convertExists.
func convertQuantifier(conv exprConverter, q *specir.ExprQuantifier) ast.Expr {
	if len(q.Vars) != 1 {
		panic(fmt.Sprintf(
			"quantifier: only a single bound variable is supported, got %d", len(q.Vars),
		))
	}

	v := q.Vars[0]
	if v.Type.Kind != specir.TypeInteger {
		panic(fmt.Sprintf(
			"quantifier: only 'integer' bound variable type is supported, got %v", v.Type,
		))
	}

	switch q.Kind {
	case specir.QuantifierForall:
		return convertForall(conv, q, v.Name)
	case specir.QuantifierExists:
		return convertExists(conv, q, v.Name)
	default:
		panic(fmt.Sprintf("quantifier: unknown kind %v", q.Kind))
	}
}

// rangeInfo holds the extracted loop bounds and the inner body predicate.
type rangeInfo struct {
	lo        ast.Expr
	hi        ast.Expr
	inclusive bool
	body      ast.Expr
}

// extractForallRange expects Body of the form:
//
//	(guard) ==> body
//
// where guard encodes the range of the bound variable.
// Returns nil if the pattern is not recognised.
func extractForallRange(conv exprConverter, body specir.Expr, varName string) *rangeInfo {
	impl, ok := body.(*specir.ExprBinary)
	if !ok || impl.Op != specir.OpImplies {
		return nil
	}
	lo, hi, inclusive := extractBounds(impl.Left, varName)
	if lo == nil || hi == nil {
		return nil
	}
	return &rangeInfo{
		lo:        conv.Convert(lo),
		hi:        conv.Convert(hi),
		inclusive: inclusive,
		body:      conv.Convert(impl.Right),
	}
}

// extractExistsRange expects Body of the form:
//
//	(guard) && body
//
// where guard encodes the range of the bound variable.
// Returns nil if the pattern is not recognised.
func extractExistsRange(conv exprConverter, body specir.Expr, varName string) *rangeInfo {
	and, ok := body.(*specir.ExprBinary)
	if !ok || and.Op != specir.OpAnd {
		return nil
	}
	lo, hi, inclusive := extractBounds(and.Left, varName)
	if lo == nil || hi == nil {
		return nil
	}
	return &rangeInfo{
		lo:        conv.Convert(lo),
		hi:        conv.Convert(hi),
		inclusive: inclusive,
		body:      conv.Convert(and.Right),
	}
}

// extractBounds tries to parse a guard expression of the form:
//
//	lo <= i && i < hi
//	lo <= i && i <= hi
//	i < hi               (lo defaults to 0)
//	i <= hi              (lo defaults to 0)
//
// Returns (lo, hi, inclusive) as IR Expr nodes, or nil bounds if the pattern is not matched.
func extractBounds(guard specir.Expr, varName string) (lo, hi specir.Expr, inclusive bool) {
	// Case: lo <= i && i < hi  OR  lo <= i && i <= hi
	if and, ok := guard.(*specir.ExprBinary); ok && and.Op == specir.OpAnd {
		loExpr := extractLowerBound(and.Left, varName)
		hiExpr, inclusive := extractUpperBound(and.Right, varName)
		if loExpr != nil && hiExpr != nil {
			return loExpr, hiExpr, inclusive
		}
		// Try reversed order: i < hi && lo <= i
		loExpr = extractLowerBound(and.Right, varName)
		hiExpr, inclusive = extractUpperBound(and.Left, varName)
		if loExpr != nil && hiExpr != nil {
			return loExpr, hiExpr, inclusive
		}
	}

	// Case: i < hi  (no lower bound — default to 0)
	if hiExpr, inclusive := extractUpperBound(guard, varName); hiExpr != nil {
		zero := &specir.ExprLiteral{Kind: specir.LitInteger, Int: 0}
		return zero, hiExpr, inclusive
	}

	return nil, nil, false
}

// extractLowerBound matches  lo <= i  (i.e. ExprBinary{OpLe, lo, ExprVar{varName}}).
func extractLowerBound(expr specir.Expr, varName string) specir.Expr {
	bin, ok := expr.(*specir.ExprBinary)
	if !ok {
		return nil
	}
	if bin.Op == specir.OpLe {
		if isVar(bin.Right, varName) {
			return bin.Left
		}
	}
	return nil
}

// extractUpperBound matches i < hi or i <= hi and reports whether the bound is inclusive.
func extractUpperBound(expr specir.Expr, varName string) (specir.Expr, bool) {
	bin, ok := expr.(*specir.ExprBinary)
	if !ok {
		return nil, false
	}
	switch bin.Op {
	case specir.OpLt:
		// i < hi
		if isVar(bin.Left, varName) {
			return bin.Right, false
		}
	case specir.OpLe:
		if isVar(bin.Left, varName) {
			return bin.Right, true
		}
	}
	return nil, false
}

func isVar(expr specir.Expr, name string) bool {
	v, ok := expr.(*specir.ExprVar)
	return ok && v.Name == name
}

// convertForall generates:
//
//	func() bool {
//	    for i := lo; i < hi; i++ {
//	        if !(body) { return false }
//	    }
//	    return true
//	}()
func convertForall(conv exprConverter, q *specir.ExprQuantifier, varName string) ast.Expr {
	ri := extractForallRange(conv, q.Body, varName)
	if ri == nil {
		panic(fmt.Sprintf(
			`\forall: cannot infer loop range for variable %q. `+
				`Body must have the form "(lo <= %s && %s < hi) ==> P(%s)"`,
			varName, varName, varName, varName,
		))
	}
	return buildQuantifierIIFE(varName, ri.lo, ri.hi, ri.inclusive, ri.body, true)
}

// convertExists generates:
//
//	func() bool {
//	    for i := lo; i < hi; i++ {
//	        if body { return true }
//	    }
//	    return false
//	}()
func convertExists(conv exprConverter, q *specir.ExprQuantifier, varName string) ast.Expr {
	ri := extractExistsRange(conv, q.Body, varName)
	if ri == nil {
		panic(fmt.Sprintf(
			`\exists: cannot infer loop range for variable %q. `+
				`Body must have the form "(lo <= %s && %s < hi) && P(%s)"`,
			varName, varName, varName, varName,
		))
	}
	return buildQuantifierIIFE(varName, ri.lo, ri.hi, ri.inclusive, ri.body, false)
}

// buildQuantifierIIFE builds the IIFE for both forall and exists.
//
// forall=true:
//
//	func() bool { for i:=lo; i<hi; i++ { if !(cond) { return false } }; return true }()
//
// forall=false (exists):
//
//	func() bool { for i:=lo; i<hi; i++ { if cond { return true } }; return false }()
func buildQuantifierIIFE(varName string, lo, hi ast.Expr, inclusive bool, cond ast.Expr, forall bool) ast.Expr {
	iVar := &ast.Ident{Name: varName}

	comparison := token.LSS
	if inclusive {
		comparison = token.LEQ
	}
	loopCond := &ast.BinaryExpr{X: iVar, Op: comparison, Y: hi}

	// if-condition and early-return value
	var ifCond ast.Expr
	var earlyReturn ast.Expr
	var finalReturn ast.Expr

	if forall {
		// if !(cond) { return false }
		ifCond = &ast.UnaryExpr{Op: token.NOT, X: &ast.ParenExpr{X: cond}}
		earlyReturn = &ast.Ident{Name: "false"}
		finalReturn = &ast.Ident{Name: "true"}
	} else {
		// if cond { return true }
		ifCond = cond
		earlyReturn = &ast.Ident{Name: "true"}
		finalReturn = &ast.Ident{Name: "false"}
	}

	forBody := &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.IfStmt{
				Cond: ifCond,
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{Results: []ast.Expr{earlyReturn}},
					},
				},
			},
		},
	}
	var loopPost ast.Stmt = &ast.IncDecStmt{X: iVar, Tok: token.INC}
	if inclusive {
		// Avoid incrementing the maximum integer value. The explicit break keeps
		// an inclusive range finite without evaluating hi + 1.
		forBody.List = append(forBody.List,
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{X: iVar, Op: token.EQL, Y: hi},
				Body: &ast.BlockStmt{List: []ast.Stmt{&ast.BranchStmt{Tok: token.BREAK}}},
			},
			&ast.IncDecStmt{X: iVar, Tok: token.INC},
		)
		loopPost = nil
	}

	forStmt := &ast.ForStmt{
		Init: &ast.AssignStmt{
			Lhs: []ast.Expr{iVar},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{lo},
		},
		Cond: loopCond,
		Post: loopPost,
		Body: forBody,
	}

	funcLit := &ast.FuncLit{
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{Type: &ast.Ident{Name: "bool"}},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				forStmt,
				&ast.ReturnStmt{Results: []ast.Expr{finalReturn}},
			},
		},
	}

	return &ast.CallExpr{Fun: funcLit}
}

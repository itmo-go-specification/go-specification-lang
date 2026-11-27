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
	lo   ast.Expr
	hi   ast.Expr
	body ast.Expr
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
	lo, hi, remaining := splitBounds(impl.Left, varName)
	if lo == nil || hi == nil {
		return nil
	}
	predicate := impl.Right
	if len(remaining) > 0 {
		predicate = &specir.ExprBinary{
			Op:    specir.OpImplies,
			Left:  joinAnd(remaining),
			Right: predicate,
		}
	}
	return &rangeInfo{
		lo:   conv.Convert(lo),
		hi:   conv.Convert(hi),
		body: conv.Convert(predicate),
	}
}

// extractExistsRange expects Body of the form:
//
//	(guard) && body
//
// where guard encodes the range of the bound variable.
// Returns nil if the pattern is not recognised.
func extractExistsRange(conv exprConverter, body specir.Expr, varName string) *rangeInfo {
	lo, hi, remaining := splitBounds(body, varName)
	if lo == nil || hi == nil {
		return nil
	}
	predicate := joinAnd(remaining)
	if predicate == nil {
		predicate = &specir.ExprLiteral{Kind: specir.LitBool, Bool: true}
	}
	return &rangeInfo{
		lo:   conv.Convert(lo),
		hi:   conv.Convert(hi),
		body: conv.Convert(predicate),
	}
}

// splitBounds flattens conjunctions, consumes one lower and one upper bound,
// and returns every other condition in its original order.
func splitBounds(expr specir.Expr, varName string) (lo, hi specir.Expr, remaining []specir.Expr) {
	for _, term := range flattenAnd(expr) {
		if lo == nil {
			if bound := extractLowerBound(term, varName); bound != nil {
				lo = bound
				continue
			}
		}
		if hi == nil {
			if bound := extractUpperBound(term, varName); bound != nil {
				hi = bound
				continue
			}
		}
		remaining = append(remaining, term)
	}
	return lo, hi, remaining
}

func flattenAnd(expr specir.Expr) []specir.Expr {
	if and, ok := expr.(*specir.ExprBinary); ok && and.Op == specir.OpAnd {
		return append(flattenAnd(and.Left), flattenAnd(and.Right)...)
	}
	return []specir.Expr{expr}
}

func joinAnd(exprs []specir.Expr) specir.Expr {
	if len(exprs) == 0 {
		return nil
	}
	result := exprs[0]
	for _, expr := range exprs[1:] {
		result = &specir.ExprBinary{Op: specir.OpAnd, Left: result, Right: expr}
	}
	return result
}

// extractBounds tries to parse a guard expression of the form:
//
//	lo <= i && i < hi
//	lo <= i && i <= hi   (hi becomes hi+1 conceptually — we use OpLe → OpLt adjustment)
//	i < hi               (lo defaults to 0)
//	i <= hi              (lo defaults to 0, hi+1 conceptually)
//
// Returns (lo, hi) as IR Expr nodes, or (nil, nil) if the pattern is not matched.
func extractBounds(guard specir.Expr, varName string) (lo, hi specir.Expr) {
	loExpr, hiExpr, _ := splitBounds(guard, varName)
	if loExpr != nil && hiExpr != nil {
		return loExpr, hiExpr
	}

	// Case: i < hi  (no lower bound — default to 0)
	if hiExpr := extractUpperBound(guard, varName); hiExpr != nil {
		zero := &specir.ExprLiteral{Kind: specir.LitInteger, Int: 0}
		return zero, hiExpr
	}

	return nil, nil
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
	if bin.Op == specir.OpGe && isVar(bin.Left, varName) {
		return bin.Right
	}
	return nil
}

// extractUpperBound matches  i < hi  or  i <= hi.
// For  i <= hi  we wrap hi in  hi+1  so the generated for-loop uses strict <.
func extractUpperBound(expr specir.Expr, varName string) specir.Expr {
	bin, ok := expr.(*specir.ExprBinary)
	if !ok {
		return nil
	}
	switch bin.Op {
	case specir.OpLt:
		// i < hi
		if isVar(bin.Left, varName) {
			return bin.Right
		}
	case specir.OpLe:
		// i <= hi  →  treat upper bound as hi+1
		if isVar(bin.Left, varName) {
			return &specir.ExprBinary{
				Op:    specir.OpAdd,
				Left:  bin.Right,
				Right: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 1},
			}
		}
	case specir.OpGt:
		// hi > i
		if isVar(bin.Right, varName) {
			return bin.Left
		}
	case specir.OpGe:
		// hi >= i  →  treat upper bound as hi+1
		if isVar(bin.Right, varName) {
			return &specir.ExprBinary{
				Op:    specir.OpAdd,
				Left:  bin.Left,
				Right: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 1},
			}
		}
	}
	return nil
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
	return buildQuantifierIIFE(varName, ri.lo, ri.hi, ri.body, true)
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
	return buildQuantifierIIFE(varName, ri.lo, ri.hi, ri.body, false)
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
func buildQuantifierIIFE(varName string, lo, hi, cond ast.Expr, forall bool) ast.Expr {
	iVar := &ast.Ident{Name: varName}

	// Loop condition: i < hi
	loopCond := &ast.BinaryExpr{X: iVar, Op: token.LSS, Y: hi}

	// Loop post: i++
	loopPost := &ast.IncDecStmt{X: iVar, Tok: token.INC}

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

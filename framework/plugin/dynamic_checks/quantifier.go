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
	lo, hi := extractBounds(impl.Left, varName)
	if lo == nil || hi == nil {
		return nil
	}
	return &rangeInfo{
		lo:   conv.Convert(lo),
		hi:   conv.Convert(hi),
		body: conv.Convert(impl.Right),
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
	lo, hi := extractBounds(and.Left, varName)
	if lo == nil || hi == nil {
		return nil
	}
	return &rangeInfo{
		lo:   conv.Convert(lo),
		hi:   conv.Convert(hi),
		body: conv.Convert(and.Right),
	}
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
	parts := flattenAnd(guard)

	var loExpr, hiExpr specir.Expr

	for _, part := range parts {
		lower := extractLowerBound(part, varName)
		upper := extractUpperBound(part, varName)

		if lower != nil {
			if loExpr != nil {
				// Несколько нижних границ: шаблон неоднозначен.
				return nil, nil
			}
			loExpr = lower
		}

		if upper != nil {
			if hiExpr != nil {
				// Несколько верхних границ: шаблон неоднозначен.
				return nil, nil
			}
			hiExpr = upper
		}
	}

	// Без верхней границы нельзя построить конечный диапазон.
	if hiExpr == nil {
		return nil, nil
	}

	// Если нижняя граница не задана, используем 0.
	if loExpr == nil {
		loExpr = &specir.ExprLiteral{
			Kind: specir.LitInteger,
			Int:  0,
		}
	}

	return loExpr, hiExpr
}

func flattenAnd(expr specir.Expr) []specir.Expr {
	bin, ok := expr.(*specir.ExprBinary)
	if !ok || bin.Op != specir.OpAnd {
		return []specir.Expr{expr}
	}

	left := flattenAnd(bin.Left)
	right := flattenAnd(bin.Right)

	return append(left, right...)
}

// extractLowerBound matches lower-bound conditions of the form:
//
//	lo <= i
//	i >= lo
//	lo < i
//	i > lo
//
// Strict inequalities are adjusted by adding 1 to the bound,
// assuming integer values.
func extractLowerBound(expr specir.Expr, varName string) specir.Expr {
	bin, ok := expr.(*specir.ExprBinary)
	if !ok {
		return nil
	}

	switch bin.Op {
	case specir.OpLe:
		if isVar(bin.Right, varName) {
			return bin.Left
		}

	case specir.OpGe:
		if isVar(bin.Left, varName) {
			return bin.Right
		}

	case specir.OpLt:
		if isVar(bin.Right, varName) {
			return addOne(bin.Left)
		}

	case specir.OpGt:
		if isVar(bin.Left, varName) {
			return addOne(bin.Right)
		}
	}

	return nil
}

func addOne(expr specir.Expr) specir.Expr {
	return &specir.ExprBinary{
		Left: expr,
		Op:   specir.OpAdd,
		Right: &specir.ExprLiteral{
			Kind: specir.LitInteger,
			Int:  1,
		},
	}
}

// extractUpperBound matches upper-bound conditions of the form:
//
//	i < hi
//	i <= hi
//	hi > i
//	hi >= i
//
// Inclusive upper bounds are adjusted by adding 1, so the generated
// for-loop can use a strict less-than comparison.
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
		// i <= hi => hi + 1
		if isVar(bin.Left, varName) {
			return addOne(bin.Right)
		}

	case specir.OpGt:
		// hi > i
		if isVar(bin.Right, varName) {
			return bin.Left
		}

	case specir.OpGe:
		// hi >= i => hi + 1
		if isVar(bin.Right, varName) {
			return addOne(bin.Left)
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

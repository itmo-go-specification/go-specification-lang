package dynamicchecks

import (
	"go/ast"
	"go/token"
	specir "gpma/framework/gosl/ir"
	"testing"
)

func assertIdent(t *testing.T, expr ast.Expr, name string) {
	t.Helper()
	ident, ok := expr.(*ast.Ident)
	if !ok {
		t.Fatalf("expected *ast.Ident, got %T", expr)
	}
	if ident.Name != name {
		t.Errorf("expected ident %q, got %q", name, ident.Name)
	}
}

func assertBasicLit(t *testing.T, expr ast.Expr, kind token.Token, value string) {
	t.Helper()
	lit, ok := expr.(*ast.BasicLit)
	if !ok {
		t.Fatalf("expected *ast.BasicLit, got %T", expr)
	}
	if lit.Kind != kind {
		t.Errorf("expected token kind %v, got %v", kind, lit.Kind)
	}
	if lit.Value != value {
		t.Errorf("expected value %q, got %q", value, lit.Value)
	}
}

func TestConvertLiteral_BoolTrue(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprLiteral{Kind: specir.LitBool, Bool: true}
	result := c.Convert(expr)
	assertIdent(t, result, "true")
}

func TestConvertLiteral_BoolFalse(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprLiteral{Kind: specir.LitBool, Bool: false}
	result := c.Convert(expr)
	assertIdent(t, result, "false")
}

func TestConvertLiteral_Integer(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprLiteral{Kind: specir.LitInteger, Int: 42}
	result := c.Convert(expr)
	assertBasicLit(t, result, token.INT, "42")
}

func TestConvertLiteral_IntegerNegativeValue(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprLiteral{Kind: specir.LitInteger, Int: -7}
	result := c.Convert(expr)
	assertBasicLit(t, result, token.INT, "-7")
}

func TestConvertLiteral_IntegerZero(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprLiteral{Kind: specir.LitInteger, Int: 0}
	result := c.Convert(expr)
	assertBasicLit(t, result, token.INT, "0")
}

func TestConvertLiteral_Real(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprLiteral{Kind: specir.LitReal, Real: 3.14}
	result := c.Convert(expr)
	lit, ok := result.(*ast.BasicLit)
	if !ok {
		t.Fatalf("expected *ast.BasicLit, got %T", result)
	}
	if lit.Kind != token.FLOAT {
		t.Errorf("expected FLOAT token, got %v", lit.Kind)
	}
}

func TestConvertLiteral_Float(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprLiteral{Kind: specir.LitFloat, Float: 2.718}
	result := c.Convert(expr)
	lit, ok := result.(*ast.BasicLit)
	if !ok {
		t.Fatalf("expected *ast.BasicLit, got %T", result)
	}
	if lit.Kind != token.FLOAT {
		t.Errorf("expected FLOAT token, got %v", lit.Kind)
	}
}

func TestConvertLiteral_String(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprLiteral{Kind: specir.LitString, Str: "hello"}
	result := c.Convert(expr)
	assertBasicLit(t, result, token.STRING, `"hello"`)
}

func TestConvertVar(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprVar{Name: "x"}
	result := c.Convert(expr)
	assertIdent(t, result, "x")
}

func TestConvertUnary_Plus(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprUnary{
		Op:   specir.UnaryPlus,
		Expr: &specir.ExprVar{Name: "x"},
	}
	result := c.Convert(expr)
	unary, ok := result.(*ast.UnaryExpr)
	if !ok {
		t.Fatalf("expected *ast.UnaryExpr, got %T", result)
	}
	if unary.Op != token.ADD {
		t.Errorf("expected ADD, got %v", unary.Op)
	}
	assertIdent(t, unary.X, "x")
}

func TestConvertUnary_Minus(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprUnary{
		Op:   specir.UnaryMinus,
		Expr: &specir.ExprVar{Name: "y"},
	}
	result := c.Convert(expr)
	unary, ok := result.(*ast.UnaryExpr)
	if !ok {
		t.Fatalf("expected *ast.UnaryExpr, got %T", result)
	}
	if unary.Op != token.SUB {
		t.Errorf("expected SUB, got %v", unary.Op)
	}
	assertIdent(t, unary.X, "y")
}

func TestConvertUnary_Not(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprUnary{
		Op:   specir.UnaryNot,
		Expr: &specir.ExprLiteral{Kind: specir.LitBool, Bool: true},
	}
	result := c.Convert(expr)
	unary, ok := result.(*ast.UnaryExpr)
	if !ok {
		t.Fatalf("expected *ast.UnaryExpr, got %T", result)
	}
	if unary.Op != token.NOT {
		t.Errorf("expected NOT, got %v", unary.Op)
	}
	assertIdent(t, unary.X, "true")
}

func TestConvertBinary_AllOps(t *testing.T) {
	tests := []struct {
		op   specir.BinaryOp
		want token.Token
	}{
		{specir.OpAdd, token.ADD},
		{specir.OpSub, token.SUB},
		{specir.OpMul, token.MUL},
		{specir.OpDiv, token.QUO},
		{specir.OpMod, token.REM},
		{specir.OpEq, token.EQL},
		{specir.OpNeq, token.NEQ},
		{specir.OpLe, token.LEQ},
		{specir.OpGe, token.GEQ},
		{specir.OpLt, token.LSS},
		{specir.OpGt, token.GTR},
		{specir.OpAnd, token.LAND},
		{specir.OpOr, token.LOR},
		{specir.OpXor, token.NEQ},
	}

	for _, tt := range tests {
		t.Run(tt.want.String(), func(t *testing.T) {
			c := NewConverter()
			expr := &specir.ExprBinary{
				Op:    tt.op,
				Left:  &specir.ExprVar{Name: "a"},
				Right: &specir.ExprVar{Name: "b"},
			}
			result := c.Convert(expr)
			bin, ok := result.(*ast.BinaryExpr)
			if !ok {
				t.Fatalf("expected *ast.BinaryExpr, got %T", result)
			}
			if bin.Op != tt.want {
				t.Errorf("expected %v, got %v", tt.want, bin.Op)
			}
			assertIdent(t, bin.X, "a")
			assertIdent(t, bin.Y, "b")
		})
	}
}

func TestConvertBinary_ImpliesFallback(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprBinary{
		Op:    specir.OpImplies,
		Left:  &specir.ExprVar{Name: "p"},
		Right: &specir.ExprVar{Name: "q"},
	}
	result := c.Convert(expr)
	bin, ok := result.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr, got %T", result)
	}
	if bin.Op != token.EQL {
		t.Errorf("OpImplies should fall through to default EQL, got %v", bin.Op)
	}
}

func TestConvertBinary_NestedExpressions(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprBinary{
		Op: specir.OpAnd,
		Left: &specir.ExprBinary{
			Op:    specir.OpGe,
			Left:  &specir.ExprVar{Name: "x"},
			Right: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 0},
		},
		Right: &specir.ExprBinary{
			Op:    specir.OpLt,
			Left:  &specir.ExprVar{Name: "x"},
			Right: &specir.ExprVar{Name: "n"},
		},
	}
	result := c.Convert(expr)
	bin, ok := result.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr, got %T", result)
	}
	if bin.Op != token.LAND {
		t.Errorf("expected LAND, got %v", bin.Op)
	}

	left, ok := bin.X.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected left *ast.BinaryExpr, got %T", bin.X)
	}
	if left.Op != token.GEQ {
		t.Errorf("expected GEQ, got %v", left.Op)
	}

	right, ok := bin.Y.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected right *ast.BinaryExpr, got %T", bin.Y)
	}
	if right.Op != token.LSS {
		t.Errorf("expected LSS, got %v", right.Op)
	}
}

func TestConvertCall_NoArgs(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprCall{Name: "foo", Args: nil}
	result := c.Convert(expr)
	call, ok := result.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected *ast.CallExpr, got %T", result)
	}
	assertIdent(t, call.Fun, "foo")
	if len(call.Args) != 0 {
		t.Errorf("expected 0 args, got %d", len(call.Args))
	}
}

func TestConvertCall_SingleArg(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprCall{
		Name: "abs",
		Args: []specir.Expr{&specir.ExprVar{Name: "x"}},
	}
	result := c.Convert(expr)
	call, ok := result.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected *ast.CallExpr, got %T", result)
	}
	assertIdent(t, call.Fun, "abs")
	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(call.Args))
	}
	assertIdent(t, call.Args[0], "x")
}

func TestConvertCall_MultipleArgs(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprCall{
		Name: "Count",
		Args: []specir.Expr{
			&specir.ExprVar{Name: "a"},
			&specir.ExprVar{Name: "n"},
			&specir.ExprVar{Name: "v"},
		},
	}
	result := c.Convert(expr)
	call, ok := result.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected *ast.CallExpr, got %T", result)
	}
	assertIdent(t, call.Fun, "Count")
	if len(call.Args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(call.Args))
	}
	assertIdent(t, call.Args[0], "a")
	assertIdent(t, call.Args[1], "n")
	assertIdent(t, call.Args[2], "v")
}

func TestConvertConditional(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprConditional{
		Cond: &specir.ExprVar{Name: "flag"},
		Then: &specir.ExprVar{Name: "x"},
		Else: &specir.ExprVar{Name: "y"},
	}
	result := c.Convert(expr)

	call, ok := result.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected *ast.CallExpr (IIFE), got %T", result)
	}

	funcLit, ok := call.Fun.(*ast.FuncLit)
	if !ok {
		t.Fatalf("expected *ast.FuncLit, got %T", call.Fun)
	}

	if len(funcLit.Body.List) != 1 {
		t.Fatalf("expected 1 statement in IIFE body, got %d", len(funcLit.Body.List))
	}

	ifStmt, ok := funcLit.Body.List[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected *ast.IfStmt, got %T", funcLit.Body.List[0])
	}

	assertIdent(t, ifStmt.Cond, "flag")

	thenReturn := ifStmt.Body.List[0].(*ast.ReturnStmt)
	assertIdent(t, thenReturn.Results[0], "x")

	elseBlock := ifStmt.Else.(*ast.BlockStmt)
	elseReturn := elseBlock.List[0].(*ast.ReturnStmt)
	assertIdent(t, elseReturn.Results[0], "y")
}

func TestConvertResult(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprResult{}
	result := c.Convert(expr)
	assertIdent(t, result, "__result")
}

func TestConvertOld_BeforeCollect(t *testing.T) {
	c := NewConverter()
	inner := &specir.ExprVar{Name: "x"}
	expr := &specir.ExprOld{Inner: inner}

	result := c.Convert(expr)
	assertIdent(t, result, "x")
}

func TestConvertOld_AfterCollectNotDeclared(t *testing.T) {
	c := NewConverter()
	inner := &specir.ExprVar{Name: "x"}
	oldExpr := &specir.ExprOld{Inner: inner}

	c.CollectOld(oldExpr)

	result := c.Convert(oldExpr)
	assertIdent(t, result, "x")
}

func TestConvertOld_AfterCollectAndDeclared(t *testing.T) {
	c := NewConverter()
	inner := &specir.ExprVar{Name: "x"}
	oldExpr := &specir.ExprOld{Inner: inner}

	c.CollectOld(oldExpr)

	vars := c.GetOldVars()
	if len(vars) != 1 {
		t.Fatalf("expected 1 old var, got %d", len(vars))
	}
	vars[0].isDeclared = true

	result := c.Convert(oldExpr)
	assertIdent(t, result, "__old_1")
}

func TestCollectOld_Deduplication(t *testing.T) {
	c := NewConverter()
	inner := &specir.ExprVar{Name: "x"}

	old1 := &specir.ExprOld{Inner: inner}
	old2 := &specir.ExprOld{Inner: inner}

	c.CollectOld(old1)
	c.CollectOld(old2)

	vars := c.GetOldVars()
	if len(vars) != 1 {
		t.Errorf("expected 1 old var (deduplicated), got %d", len(vars))
	}
}

func TestCollectOld_DifferentExprs(t *testing.T) {
	c := NewConverter()
	innerX := &specir.ExprVar{Name: "x"}
	innerY := &specir.ExprVar{Name: "y"}

	c.CollectOld(&specir.ExprOld{Inner: innerX})
	c.CollectOld(&specir.ExprOld{Inner: innerY})

	vars := c.GetOldVars()
	if len(vars) != 2 {
		t.Fatalf("expected 2 old vars, got %d", len(vars))
	}
	if vars[0].name != "__old_1" {
		t.Errorf("expected __old_1, got %s", vars[0].name)
	}
	if vars[1].name != "__old_2" {
		t.Errorf("expected __old_2, got %s", vars[1].name)
	}
}

func TestCollectOld_PreservesOrder(t *testing.T) {
	c := NewConverter()
	a := &specir.ExprVar{Name: "a"}
	b := &specir.ExprVar{Name: "b"}
	cc := &specir.ExprVar{Name: "c"}

	c.CollectOld(&specir.ExprOld{Inner: cc})
	c.CollectOld(&specir.ExprOld{Inner: a})
	c.CollectOld(&specir.ExprOld{Inner: b})

	vars := c.GetOldVars()
	if len(vars) != 3 {
		t.Fatalf("expected 3 old vars, got %d", len(vars))
	}
	if vars[0].expr != cc {
		t.Error("first old var should be c")
	}
	if vars[1].expr != a {
		t.Error("second old var should be a")
	}
	if vars[2].expr != b {
		t.Error("third old var should be b")
	}
}

func TestCollectOld_NestedInBinary(t *testing.T) {
	c := NewConverter()
	innerP := &specir.ExprVar{Name: "p"}
	innerQ := &specir.ExprVar{Name: "q"}

	expr := &specir.ExprBinary{
		Op:    specir.OpAnd,
		Left:  &specir.ExprOld{Inner: innerP},
		Right: &specir.ExprOld{Inner: innerQ},
	}

	c.CollectOld(expr)

	vars := c.GetOldVars()
	if len(vars) != 2 {
		t.Fatalf("expected 2 old vars from nested binary, got %d", len(vars))
	}
}

func TestCollectOld_NoOldExpressions(t *testing.T) {
	c := NewConverter()
	expr := &specir.ExprBinary{
		Op:    specir.OpAdd,
		Left:  &specir.ExprVar{Name: "x"},
		Right: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 1},
	}

	c.CollectOld(expr)

	vars := c.GetOldVars()
	if len(vars) != 0 {
		t.Errorf("expected 0 old vars, got %d", len(vars))
	}
}

func TestNewVar_Increments(t *testing.T) {
	c := NewConverter()
	v1 := c.NewVar()
	v2 := c.NewVar()
	v3 := c.NewVar()

	if v1 != "1" {
		t.Errorf("expected '1', got %q", v1)
	}
	if v2 != "2" {
		t.Errorf("expected '2', got %q", v2)
	}
	if v3 != "3" {
		t.Errorf("expected '3', got %q", v3)
	}
}

func TestNewVar_SharedCounterWithCollectOld(t *testing.T) {
	c := NewConverter()
	inner := &specir.ExprVar{Name: "x"}
	c.CollectOld(&specir.ExprOld{Inner: inner})

	v := c.NewVar()
	if v != "2" {
		t.Errorf("expected '2' (counter shared with CollectOld), got %q", v)
	}
}

func TestConvert_ComplexExpression(t *testing.T) {
	c := NewConverter()

	// (\result >= 0) && ((\result == x) || (\result == -x))
	expr := &specir.ExprBinary{
		Op: specir.OpAnd,
		Left: &specir.ExprBinary{
			Op:    specir.OpGe,
			Left:  &specir.ExprResult{},
			Right: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 0},
		},
		Right: &specir.ExprBinary{
			Op: specir.OpOr,
			Left: &specir.ExprBinary{
				Op:    specir.OpEq,
				Left:  &specir.ExprResult{},
				Right: &specir.ExprVar{Name: "x"},
			},
			Right: &specir.ExprBinary{
				Op:   specir.OpEq,
				Left: &specir.ExprResult{},
				Right: &specir.ExprUnary{
					Op:   specir.UnaryMinus,
					Expr: &specir.ExprVar{Name: "x"},
				},
			},
		},
	}

	result := c.Convert(expr)

	top, ok := result.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr at top, got %T", result)
	}
	if top.Op != token.LAND {
		t.Errorf("expected LAND at top, got %v", top.Op)
	}

	left, ok := top.X.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr on left, got %T", top.X)
	}
	if left.Op != token.GEQ {
		t.Errorf("expected GEQ, got %v", left.Op)
	}
	assertIdent(t, left.X, "__result")

	right, ok := top.Y.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr on right, got %T", top.Y)
	}
	if right.Op != token.LOR {
		t.Errorf("expected LOR, got %v", right.Op)
	}
}

func TestConvert_OldWithComplexInner(t *testing.T) {
	c := NewConverter()

	// \old(*p) where *p is represented as ExprVar for simplicity
	innerDerefP := &specir.ExprVar{Name: "*p"}
	innerDerefQ := &specir.ExprVar{Name: "*q"}

	// ensures (*p == \old(*q)) && (*q == \old(*p))
	expr := &specir.ExprBinary{
		Op: specir.OpAnd,
		Left: &specir.ExprBinary{
			Op:    specir.OpEq,
			Left:  &specir.ExprVar{Name: "*p"},
			Right: &specir.ExprOld{Inner: innerDerefQ},
		},
		Right: &specir.ExprBinary{
			Op:    specir.OpEq,
			Left:  &specir.ExprVar{Name: "*q"},
			Right: &specir.ExprOld{Inner: innerDerefP},
		},
	}

	c.CollectOld(expr)
	vars := c.GetOldVars()
	if len(vars) != 2 {
		t.Fatalf("expected 2 old vars, got %d", len(vars))
	}

	for _, v := range vars {
		v.isDeclared = true
	}

	result := c.Convert(expr)
	top, ok := result.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr, got %T", result)
	}

	leftBin := top.X.(*ast.BinaryExpr)
	assertIdent(t, leftBin.Y, "__old_1")

	rightBin := top.Y.(*ast.BinaryExpr)
	assertIdent(t, rightBin.Y, "__old_2")
}

func TestGetOldVars_EmptyConverter(t *testing.T) {
	c := NewConverter()
	vars := c.GetOldVars()
	if len(vars) != 0 {
		t.Errorf("expected 0 old vars from fresh converter, got %d", len(vars))
	}
}

func TestConvert_PanicsOnNilExprType(t *testing.T) {
	c := NewConverter()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on unknown expr type")
		}
	}()

	c.Convert(&specir.ExprQuantifier{
		Kind: specir.QuantifierForall,
		Body: &specir.ExprLiteral{Kind: specir.LitBool, Bool: true},
	})
}

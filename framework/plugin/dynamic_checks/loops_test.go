package dynamicchecks

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"gpma/framework/data"
	specir "gpma/framework/gosl/ir"
	"gpma/framework/plugin"
)

func TestGenerateLoopChecks_InvariantsOnly(t *testing.T) {
	spec := &specir.LoopSpec{
		Invariants: []specir.NamedPredicate{
			{Label: "bound", Expr: &specir.ExprBinary{
				Op: specir.OpLe, Left: &specir.ExprVar{Name: "i"}, Right: &specir.ExprVar{Name: "n"},
			}},
		},
	}
	c := NewConverter()
	checks := generateLoopChecks(spec, c)

	if len(checks.BeginStmts) != 1 {
		t.Fatalf("expected 1 begin stmt, got %d", len(checks.BeginStmts))
	}
	if len(checks.EndStmts) != 1 {
		t.Fatalf("expected 1 end stmt, got %d", len(checks.EndStmts))
	}

	beginES := checks.BeginStmts[0].(*ast.ExprStmt)
	beginCall := beginES.X.(*ast.CallExpr)
	assertIdent(t, beginCall.Fun, "checks.Check")
	beginMsg := beginCall.Args[1].(*ast.BasicLit)
	if beginMsg.Value != `"loop invariant: bound"` {
		t.Errorf("begin msg = %q", beginMsg.Value)
	}

	endES := checks.EndStmts[0].(*ast.ExprStmt)
	endCall := endES.X.(*ast.CallExpr)
	assertIdent(t, endCall.Fun, "checks.Check")
}

func TestGenerateLoopChecks_MultipleInvariants(t *testing.T) {
	spec := &specir.LoopSpec{
		Invariants: []specir.NamedPredicate{
			{Label: "bound", Expr: &specir.ExprVar{Name: "a"}},
			{Label: "count", Expr: &specir.ExprVar{Name: "b"}},
		},
	}
	c := NewConverter()
	checks := generateLoopChecks(spec, c)

	if len(checks.BeginStmts) != 2 {
		t.Errorf("expected 2 begin stmts, got %d", len(checks.BeginStmts))
	}
	if len(checks.EndStmts) != 2 {
		t.Errorf("expected 2 end stmts, got %d", len(checks.EndStmts))
	}
}

func TestGenerateLoopChecks_VariantOnly(t *testing.T) {
	spec := &specir.LoopSpec{
		Variant: &specir.ExprBinary{
			Op: specir.OpSub, Left: &specir.ExprVar{Name: "n"}, Right: &specir.ExprVar{Name: "i"},
		},
	}
	c := NewConverter()
	checks := generateLoopChecks(spec, c)

	if len(checks.BeginStmts) != 1 {
		t.Fatalf("expected 1 begin stmt (variant save), got %d", len(checks.BeginStmts))
	}

	assign, ok := checks.BeginStmts[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected *ast.AssignStmt, got %T", checks.BeginStmts[0])
	}
	if assign.Tok != token.DEFINE {
		t.Errorf("expected DEFINE, got %v", assign.Tok)
	}
	beforeName := assign.Lhs[0].(*ast.Ident).Name
	if beforeName != "__variant_before_1" {
		t.Errorf("expected __variant_before_1, got %s", beforeName)
	}

	if len(checks.EndStmts) != 2 {
		t.Fatalf("expected 2 end stmts (assign + check), got %d", len(checks.EndStmts))
	}

	afterAssign, ok := checks.EndStmts[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected *ast.AssignStmt, got %T", checks.EndStmts[0])
	}
	afterName := afterAssign.Lhs[0].(*ast.Ident).Name
	if afterName != "__variant_after_2" {
		t.Errorf("expected __variant_after_2, got %s", afterName)
	}

	checkES := checks.EndStmts[1].(*ast.ExprStmt)
	checkCall := checkES.X.(*ast.CallExpr)
	assertIdent(t, checkCall.Fun, "checks.Check")

	cond := checkCall.Args[0].(*ast.BinaryExpr)
	if cond.Op != token.LSS {
		t.Errorf("expected LSS (after < before), got %v", cond.Op)
	}
	assertIdent(t, cond.X, "__variant_after_2")
	assertIdent(t, cond.Y, "__variant_before_1")

	checkMsg := checkCall.Args[1].(*ast.BasicLit)
	if checkMsg.Value != `"loop variant"` {
		t.Errorf("msg = %q", checkMsg.Value)
	}
}

func TestGenerateLoopChecks_InvariantsAndVariant(t *testing.T) {
	spec := &specir.LoopSpec{
		Invariants: []specir.NamedPredicate{
			{Label: "bound", Expr: &specir.ExprVar{Name: "x"}},
		},
		Variant: &specir.ExprVar{Name: "n"},
	}
	c := NewConverter()
	checks := generateLoopChecks(spec, c)

	if len(checks.BeginStmts) != 2 {
		t.Errorf("expected 2 begin stmts (1 invariant + 1 variant save), got %d", len(checks.BeginStmts))
	}

	if len(checks.EndStmts) != 3 {
		t.Errorf("expected 3 end stmts (1 invariant + 1 variant assign + 1 variant check), got %d", len(checks.EndStmts))
	}
}

func TestGenerateLoopChecks_EmptySpec(t *testing.T) {
	spec := &specir.LoopSpec{}
	c := NewConverter()
	checks := generateLoopChecks(spec, c)

	if len(checks.BeginStmts) != 0 {
		t.Errorf("expected 0 begin stmts, got %d", len(checks.BeginStmts))
	}
	if len(checks.EndStmts) != 0 {
		t.Errorf("expected 0 end stmts, got %d", len(checks.EndStmts))
	}
}

func TestGenerateLoopChecks_InvariantWithoutLabel(t *testing.T) {
	spec := &specir.LoopSpec{
		Invariants: []specir.NamedPredicate{
			{Expr: &specir.ExprVar{Name: "cond"}},
		},
	}
	c := NewConverter()
	checks := generateLoopChecks(spec, c)

	es := checks.BeginStmts[0].(*ast.ExprStmt)
	call := es.X.(*ast.CallExpr)
	msg := call.Args[1].(*ast.BasicLit)
	if msg.Value != `"loop invariant"` {
		t.Errorf("expected no label in message, got %q", msg.Value)
	}
}

func TestProcessForLoop_NilBody(t *testing.T) {
	forStmt := &ast.ForStmt{Body: nil}
	spec := &specir.LoopSpec{
		Invariants: []specir.NamedPredicate{{Expr: &specir.ExprVar{Name: "x"}}},
	}
	result := processForLoop(forStmt, spec)
	if result {
		t.Error("should return false for nil body")
	}
}

func TestProcessRangeLoop_NilBody(t *testing.T) {
	rangeStmt := &ast.RangeStmt{Body: nil}
	spec := &specir.LoopSpec{
		Invariants: []specir.NamedPredicate{{Expr: &specir.ExprVar{Name: "x"}}},
	}
	result := processRangeLoop(rangeStmt, spec)
	if result {
		t.Error("should return false for nil body")
	}
}

func TestInsertLoopChecks_PanicsNotImplemented(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic from NOT IMPLEMENTED")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if msg != "NOT IMPLEMENTED" {
			t.Errorf("expected 'NOT IMPLEMENTED', got %q", msg)
		}
	}()

	body := &ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{X: &ast.Ident{Name: "work"}}}}
	checks := &LoopChecks{
		BeginStmts: []ast.Stmt{&ast.ExprStmt{X: &ast.Ident{Name: "begin"}}},
		EndStmts:   []ast.Stmt{&ast.ExprStmt{X: &ast.Ident{Name: "end"}}},
	}
	insertLoopChecks(body, checks)
}

func TestProcessForLoop_PanicsOnInsert(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic from NOT IMPLEMENTED")
		}
	}()

	forStmt := &ast.ForStmt{
		Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{X: &ast.Ident{Name: "work"}}}},
	}
	spec := &specir.LoopSpec{
		Invariants: []specir.NamedPredicate{{Expr: &specir.ExprVar{Name: "x"}}},
	}
	processForLoop(forStmt, spec)
}

func TestProcessRangeLoop_PanicsOnInsert(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic from NOT IMPLEMENTED")
		}
	}()

	rangeStmt := &ast.RangeStmt{
		Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{X: &ast.Ident{Name: "work"}}}},
	}
	spec := &specir.LoopSpec{
		Invariants: []specir.NamedPredicate{{Expr: &specir.ExprVar{Name: "x"}}},
	}
	processRangeLoop(rangeStmt, spec)
}

func TestProcessLoops_NoSpecs(t *testing.T) {
	src := `package test
func foo() {
	for i := 0; i < 10; i++ {}
}`
	fset := token.NewFileSet()
	file, _ := parseSource(t, fset, src)
	ad := data.NewASTData(file, fset)
	result := plugin.NewPluginResult()
	processLoops(ad, result)
	if result.ASTModified {
		t.Error("should not modify AST without specs")
	}
}

func TestLoopChecks_StructFields(t *testing.T) {
	lc := &LoopChecks{}
	if lc.BeginStmts != nil {
		t.Error("BeginStmts should be nil initially")
	}
	if lc.EndStmts != nil {
		t.Error("EndStmts should be nil initially")
	}

	lc.BeginStmts = append(lc.BeginStmts, &ast.ExprStmt{X: &ast.Ident{Name: "a"}})
	lc.EndStmts = append(lc.EndStmts, &ast.ExprStmt{X: &ast.Ident{Name: "b"}})

	if len(lc.BeginStmts) != 1 {
		t.Errorf("expected 1 begin, got %d", len(lc.BeginStmts))
	}
	if len(lc.EndStmts) != 1 {
		t.Errorf("expected 1 end, got %d", len(lc.EndStmts))
	}
}

func TestGenerateLoopChecks_VariantCounterNames(t *testing.T) {
	spec := &specir.LoopSpec{
		Invariants: []specir.NamedPredicate{
			{Expr: &specir.ExprVar{Name: "a"}},
			{Expr: &specir.ExprVar{Name: "b"}},
		},
		Variant: &specir.ExprVar{Name: "v"},
	}
	c := NewConverter()
	checks := generateLoopChecks(spec, c)

	variantSave := checks.BeginStmts[2].(*ast.AssignStmt)
	name := variantSave.Lhs[0].(*ast.Ident).Name
	if name != "__variant_before_1" {
		t.Errorf("expected __variant_before_1, got %s", name)
	}

	variantAfter := checks.EndStmts[2].(*ast.AssignStmt)
	afterName := variantAfter.Lhs[0].(*ast.Ident).Name
	if afterName != "__variant_after_2" {
		t.Errorf("expected __variant_after_2, got %s", afterName)
	}
}

func parseSource(t *testing.T, fset *token.FileSet, src string) (*ast.File, error) {
	t.Helper()
	return parser.ParseFile(fset, "test.go", src, parser.ParseComments)
}

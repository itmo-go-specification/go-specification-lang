package dynamicchecks

import (
	"go/ast"
	"go/token"
	"testing"

	"gpma/framework/data"
	specir "gpma/framework/gosl/ir"
)

func TestInsertLoopChecksBeforeContinueAndAtEnd(t *testing.T) {
	check := &ast.ExprStmt{X: &ast.Ident{Name: "check"}}
	body := &ast.BlockStmt{List: []ast.Stmt{
		&ast.IfStmt{
			Cond: &ast.Ident{Name: "skip"},
			Body: &ast.BlockStmt{List: []ast.Stmt{&ast.BranchStmt{Tok: token.CONTINUE}}},
		},
		&ast.ExprStmt{X: &ast.Ident{Name: "work"}},
	}}

	got, modified := insertLoopChecks(body, &loopChecks{endStmts: []ast.Stmt{check}})
	if !modified || len(got) != 3 || got[2] != check {
		t.Fatalf("expected final check, got %#v", got)
	}
	ifBody := got[0].(*ast.IfStmt).Body.List
	if len(ifBody) != 2 || ifBody[0] != check {
		t.Fatalf("expected check before continue, got %#v", ifBody)
	}
}

func TestApplyLoopPassSupportsRangeStmt(t *testing.T) {
	rangeStmt := &ast.RangeStmt{
		Body: &ast.BlockStmt{},
		X:    &ast.Ident{Name: "values"},
	}
	fn := &ast.FuncDecl{
		Name: &ast.Ident{Name: "check"},
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{List: []ast.Stmt{rangeStmt}},
	}
	file := &ast.File{
		Name:  &ast.Ident{Name: "example"},
		Decls: []ast.Decl{fn},
	}
	astData := data.NewASTData(file, token.NewFileSet(), "")
	info := data.NewAdditionalInfo()
	info.LoopSpec = &specir.LoopSpec{Invariants: []specir.NamedPredicate{{
		Expr: &specir.ExprLiteral{Kind: specir.LitBool, Bool: true},
	}}}
	astData.SetNodeInfo(rangeStmt, info)

	if !applyLoopPass(astData) {
		t.Fatal("expected range loop to be instrumented")
	}
	if len(fn.Body.List) != 2 {
		t.Fatalf("expected pre-loop check and range statement, got %d statements", len(fn.Body.List))
	}
	if len(rangeStmt.Body.List) != 1 {
		t.Fatalf("expected invariant check in range body, got %d statements", len(rangeStmt.Body.List))
	}
}

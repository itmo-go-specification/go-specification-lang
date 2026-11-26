package dynamicchecks

import (
	"go/ast"
	"go/token"
	"testing"
)

func TestCreateResultReturnBlockUsesSimultaneousAssignment(t *testing.T) {
	ret := &ast.ReturnStmt{Results: []ast.Expr{
		&ast.Ident{Name: "b"},
		&ast.Ident{Name: "a"},
	}}
	infos := []resultInfo{
		{name: "a", isNamed: true},
		{name: "b", isNamed: true},
	}

	block := createResultReturnBlock(ret, infos)
	if len(block.List) != 2 {
		t.Fatalf("expected assignment and return, got %d statements", len(block.List))
	}
	assign, ok := block.List[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected first statement to be an assignment, got %T", block.List[0])
	}
	if assign.Tok != token.ASSIGN || len(assign.Lhs) != 2 || len(assign.Rhs) != 2 {
		t.Fatalf("expected one 2-to-2 assignment, got %d lhs and %d rhs", len(assign.Lhs), len(assign.Rhs))
	}
	if assign.Lhs[0].(*ast.Ident).Name != "a" || assign.Lhs[1].(*ast.Ident).Name != "b" {
		t.Fatalf("unexpected assignment targets: %#v", assign.Lhs)
	}
	if assign.Rhs[0].(*ast.Ident).Name != "b" || assign.Rhs[1].(*ast.Ident).Name != "a" {
		t.Fatalf("unexpected assignment values: %#v", assign.Rhs)
	}
	if ret := block.List[1].(*ast.ReturnStmt); len(ret.Results) != 0 {
		t.Fatalf("named results must use a bare return, got %d expressions", len(ret.Results))
	}
}

func TestCreateResultReturnBlockAcceptsMultiValueCall(t *testing.T) {
	call := &ast.CallExpr{Fun: &ast.Ident{Name: "pair"}}
	ret := &ast.ReturnStmt{Results: []ast.Expr{call}}
	infos := []resultInfo{
		{name: "a", isNamed: true},
		{name: "b", isNamed: true},
	}

	block := createResultReturnBlock(ret, infos)
	assign := block.List[0].(*ast.AssignStmt)
	if len(assign.Lhs) != 2 || len(assign.Rhs) != 1 || assign.Rhs[0] != call {
		t.Fatalf("expected a, b = pair(), got %d lhs and %d rhs", len(assign.Lhs), len(assign.Rhs))
	}
}

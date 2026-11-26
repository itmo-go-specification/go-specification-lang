package dynamicchecks

import (
	"go/ast"
	"testing"

	specir "gpma/framework/gosl/ir"
)

func TestConvertResultFieldAccess(t *testing.T) {
	conv := newContractConverter()
	conv.SetResultInfos([]resultInfo{{name: "result", isNamed: true}})
	expr := &specir.ExprSelector{
		Expr:  &specir.ExprResult{},
		Field: "Value",
	}

	got, ok := conv.Convert(expr).(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("expected selector expression, got %T", conv.Convert(expr))
	}
	if got.X.(*ast.Ident).Name != "result" || got.Sel.Name != "Value" {
		t.Fatalf("expected result.Value, got %#v", got)
	}
}

func TestConvertResultIndexAccess(t *testing.T) {
	conv := newContractConverter()
	conv.SetResultInfos([]resultInfo{{name: "result", isNamed: true}})
	expr := &specir.ExprIndex{
		Expr:  &specir.ExprResult{},
		Index: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 0},
	}

	got, ok := conv.Convert(expr).(*ast.IndexExpr)
	if !ok {
		t.Fatalf("expected index expression, got %T", conv.Convert(expr))
	}
	if got.X.(*ast.Ident).Name != "result" || got.Index.(*ast.BasicLit).Value != "0" {
		t.Fatalf("expected result[0], got %#v", got)
	}
}

package dynamicchecks

import (
	"go/ast"
	"testing"

	specir "gpma/framework/gosl/ir"
)

func TestConvertConditionalInfersIntegerResult(t *testing.T) {
	got := convertConditional(&baseConverter{}, &specir.ExprConditional{
		Cond: &specir.ExprVar{Name: "flag"},
		Then: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 1},
		Else: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 2},
	}).(*ast.CallExpr)
	fn := got.Fun.(*ast.FuncLit)
	result, ok := fn.Type.Results.List[0].Type.(*ast.Ident)
	if !ok || result.Name != "int" {
		t.Fatalf("expected int result type, got %#v", fn.Type.Results.List[0].Type)
	}
}

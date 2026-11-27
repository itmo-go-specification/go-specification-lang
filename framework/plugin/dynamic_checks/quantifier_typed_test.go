package dynamicchecks

import (
	"go/ast"
	"testing"

	specir "gpma/framework/gosl/ir"
)

func TestConvertRangeBoundsCastsUntypedLowerBound(t *testing.T) {
	conv := newContractConverter()
	conv.varTypes["n"] = &ast.Ident{Name: "int64"}
	lo, _ := convertRangeBounds(conv,
		&specir.ExprLiteral{Kind: specir.LitInteger, Int: 0},
		&specir.ExprVar{Name: "n"},
	)

	call, ok := lo.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected conversion call, got %T", lo)
	}
	if call.Fun.(*ast.Ident).Name != "int64" {
		t.Fatalf("expected int64 conversion, got %#v", call.Fun)
	}
}

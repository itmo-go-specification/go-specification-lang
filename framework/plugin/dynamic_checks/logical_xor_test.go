package dynamicchecks

import (
	"go/ast"
	"go/token"
	"testing"

	specir "gpma/framework/gosl/ir"
)

func TestConvertLogicalXor(t *testing.T) {
	got := (&baseConverter{}).Convert(&specir.ExprBinary{
		Op:    specir.OpXor,
		Left:  &specir.ExprVar{Name: "p"},
		Right: &specir.ExprVar{Name: "q"},
	})
	paren, ok := got.(*ast.ParenExpr)
	if !ok {
		t.Fatalf("expected parenthesized expression, got %#v", got)
	}
	binary, ok := paren.X.(*ast.BinaryExpr)
	if !ok || binary.Op != token.NEQ {
		t.Fatalf("expected != binary expression, got %#v", got)
	}
}

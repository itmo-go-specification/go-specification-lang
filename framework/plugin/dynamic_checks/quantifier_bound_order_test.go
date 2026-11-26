package dynamicchecks

import (
	"testing"

	specir "gpma/framework/gosl/ir"
)

func TestSplitBoundsPreservesPredicate(t *testing.T) {
	loTerm := &specir.ExprBinary{
		Op: specir.OpLe, Left: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 0}, Right: &specir.ExprVar{Name: "i"},
	}
	hiTerm := &specir.ExprBinary{
		Op: specir.OpLt, Left: &specir.ExprVar{Name: "i"}, Right: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 3},
	}
	predicate := &specir.ExprBinary{
		Op: specir.OpEq, Left: &specir.ExprVar{Name: "i"}, Right: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 1},
	}
	body := &specir.ExprBinary{
		Op: specir.OpAnd, Left: predicate,
		Right: &specir.ExprBinary{Op: specir.OpAnd, Left: hiTerm, Right: loTerm},
	}

	lo, hi, remaining := splitBounds(body, "i")
	if lo == nil || hi == nil {
		t.Fatalf("expected both bounds, got lo=%#v hi=%#v", lo, hi)
	}
	if len(remaining) != 1 || remaining[0] != predicate {
		t.Fatalf("expected predicate to remain, got %#v", remaining)
	}
}

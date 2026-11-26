package dynamicchecks

import (
	"fmt"
	"strings"
	"testing"

	specir "gpma/framework/gosl/ir"
)

func TestExtractBoundsRejectsMissingLowerBound(t *testing.T) {
	guard := &specir.ExprBinary{
		Op:    specir.OpLt,
		Left:  &specir.ExprVar{Name: "i"},
		Right: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 10},
	}
	lo, hi := extractBounds(guard, "i")
	if lo != nil || hi != nil {
		t.Fatalf("expected missing range, got lo=%#v hi=%#v", lo, hi)
	}
}

func TestForallReportsMissingExplicitBound(t *testing.T) {
	q := &specir.ExprQuantifier{
		Kind: specir.QuantifierForall,
		Vars: []specir.BoundVar{{Name: "i", Type: specir.TypeExpr{Kind: specir.TypeInteger}}},
		Body: &specir.ExprBinary{
			Op: specir.OpImplies,
			Left: &specir.ExprBinary{
				Op:    specir.OpLt,
				Left:  &specir.ExprVar{Name: "i"},
				Right: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 0},
			},
			Right: &specir.ExprLiteral{Kind: specir.LitBool, Bool: false},
		},
	}

	defer func() {
		got := recover()
		if got == nil || !strings.Contains(fmt.Sprint(got), "explicit lower and upper bounds are required") {
			t.Fatalf("expected clear missing-bound error, got %v", got)
		}
	}()
	convertQuantifier(&baseConverter{}, q)
}

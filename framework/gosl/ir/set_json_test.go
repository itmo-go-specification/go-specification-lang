package specir

import (
	"encoding/json"
	"testing"
)

func TestFiniteSetJSONRoundTrip(t *testing.T) {
	original := NamedPredicate{Expr: &ExprSetSubset{
		Left: &ExprSet{Elements: []Expr{
			&ExprLiteral{Kind: LitInteger, Int: 1},
		}},
		Right: &ExprSet{Elements: []Expr{
			&ExprLiteral{Kind: LitInteger, Int: 1},
			&ExprLiteral{Kind: LitInteger, Int: 2},
		}},
	}}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded NamedPredicate
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	subset, ok := decoded.Expr.(*ExprSetSubset)
	if !ok {
		t.Fatalf("decoded expression has type %T", decoded.Expr)
	}
	right, ok := subset.Right.(*ExprSet)
	if !ok || len(right.Elements) != 2 {
		t.Fatalf("decoded right set is %#v", subset.Right)
	}
}

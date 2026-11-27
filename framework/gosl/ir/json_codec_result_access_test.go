package specir

import (
	"encoding/json"
	"testing"
)

func TestResultAccessJSONRoundTrip(t *testing.T) {
	want := NamedPredicate{Expr: &ExprIndex{
		Expr: &ExprSelector{
			Expr:  &ExprResult{Index: 1},
			Field: "Values",
		},
		Index: &ExprLiteral{Kind: LitInteger, Int: 0},
	}}

	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got NamedPredicate
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	index, ok := got.Expr.(*ExprIndex)
	if !ok {
		t.Fatalf("expected ExprIndex, got %T", got.Expr)
	}
	selector, ok := index.Expr.(*ExprSelector)
	if !ok || selector.Field != "Values" {
		t.Fatalf("expected Values selector, got %#v", index.Expr)
	}
	result, ok := selector.Expr.(*ExprResult)
	if !ok || result.Index != 1 {
		t.Fatalf("expected result index 1, got %#v", selector.Expr)
	}
}

package specir

import (
	"encoding/json"
	"testing"
)

func TestFuncContractJSONRoundTripBehaviors(t *testing.T) {
	original := FuncContract{Behaviors: []Behavior{{
		Name: "positive",
		Assumes: []NamedPredicate{{
			Expr: &ExprBinary{
				Op:    OpGt,
				Left:  &ExprVar{Name: "x"},
				Right: &ExprLiteral{Kind: LitInteger, Int: 0},
			},
		}},
	}}}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	var decoded FuncContract
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Behaviors) != 1 || decoded.Behaviors[0].Name != "positive" {
		t.Fatalf("behavior was not preserved: %#v", decoded.Behaviors)
	}
	if _, ok := decoded.Behaviors[0].Assumes[0].Expr.(*ExprBinary); !ok {
		t.Fatalf("behavior assumption has unexpected type %T", decoded.Behaviors[0].Assumes[0].Expr)
	}
}

package dynamicchecks

import (
	"testing"

	specir "gpma/framework/gosl/ir"
)

func TestCollectOldLiftsQuantifiedIndexToContainer(t *testing.T) {
	inner := &specir.ExprVar{Name: "a[i]"}
	conv := newContractConverter()
	conv.CollectOld(&specir.ExprQuantifier{
		Vars: []specir.BoundVar{{Name: "i"}},
		Body: &specir.ExprOld{Inner: inner},
	})

	infos := conv.GetOldVars()
	if len(infos) != 1 {
		t.Fatalf("expected one snapshot, got %d", len(infos))
	}
	target, ok := infos[0].expr.(*specir.ExprVar)
	if !ok || target.Name != "a" || infos[0].access != "[i]" {
		t.Fatalf("expected snapshot of a with [i] access, got %#v", infos[0])
	}
}

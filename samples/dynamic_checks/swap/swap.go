package notation

import checks "gpma/framework/plugin/dynamic_checks/check"

/*
@
assigns a, b;
ensures (a == \old(b)) && (b == \old(a));
*/
func swap(a, b int) (int, int) {
	__old_1 := b
	__old_2 := a
	defer checks.Check(a == __old_1 && b == __old_2, "ensures")
	return b, a
}


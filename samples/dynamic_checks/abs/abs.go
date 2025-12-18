package notation

import checks "gpma/framework/plugin/dynamic_checks/check"

/*
@
assigns \nothing;
ensures (\result == x) || (\result == -x);
ensures \result >= 0;
*/
func Abs(x int) int {
	var __result int
	defer checks.Check(__result == x || __result == -x, "ensures")
	defer checks.Check(__result >= 0, "ensures")
	if x < 0 {
		{
			__result = -x
			return __result
		}
	}
	{
		__result = x
		return __result
	}
}


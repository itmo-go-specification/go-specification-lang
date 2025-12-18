package notation

import checks "gpma/framework/plugin/dynamic_checks/check"

/*
@
assigns \nothing;
ensures (\result == a) || (\result == b);
ensures (\result <= a) && (\result <= b);
*/
func min(a, b int) int {
	var __result int
	defer checks.Check(__result == a || __result == b, "ensures")
	defer checks.Check(__result <= a && __result <= b, "ensures")
	if a < b {
		{
			__result = a
			return __result
		}
	}
	{
		__result = b
		return __result
	}
}


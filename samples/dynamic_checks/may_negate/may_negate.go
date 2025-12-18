package notation

import checks "gpma/framework/plugin/dynamic_checks/check"

/*
@
requires (mode == \true) || (mode == \false);
assigns \nothing;
ensures ((!(mode == false)) || (\result == x)) && ((!(mode == true)) || (\result == -x));
*/
func mayNegate(x int, mode bool) int {
	checks.Check(mode == true || mode == false, "requires")
	var __result int
	defer checks.Check((!(mode == false) || __result == x) && (!(mode == true) || __result == -x), "ensures")
	if mode {
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


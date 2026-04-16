package notation

/*@
requires (mode == \true) || (mode == \false);
assigns \nothing;
ensures ((!(mode == false)) || (\result == x)) &&
		((!(mode == true)) || (\result == -x));
*/
func mayNegate(x int, mode bool) int {
	if mode {
		return -x
	}
	return x
}

// Note: because for now gosl doesn't support implications '=>'
// p -> q == !p || q
// (mode == 0 ==> \result == x) && (mode == 1 ==> \result == -x);
// (!(mode == 0) || (\result == x)) && (!(mode == 1) || (\result == -x))

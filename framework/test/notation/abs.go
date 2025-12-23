package notation

/*
@
assigns \nothing;
ensures (\result == x) || (\result == -x);
ensures \result >= 0;
*/
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

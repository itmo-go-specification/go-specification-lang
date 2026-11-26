package notation

/*@
assigns \nothing;
ensures (\result == x) || (\result == -x);
ensures \result >= 0;
*/
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

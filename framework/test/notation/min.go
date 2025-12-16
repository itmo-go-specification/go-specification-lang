package notation

/*@
assigns \nothing;
ensures (\result == a) || (\result == b);
ensures \result <= a && \result <= b;
*/
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

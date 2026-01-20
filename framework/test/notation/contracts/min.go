//go:build exclude

package contracts

/*
@
assigns \nothing;
ensures (\result == a) || (\result == b);
ensures (\result <= a) && (\result <= b);
*/
func min(a, b int) int {
	/*@
	assigns \nothing;
	ensures (\result == a) || (\result == b);
	ensures (\result <= a) && (\result <= b);
	*/
	if a < b {
		return a
	}
	return b
}

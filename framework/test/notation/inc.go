package notation

/*@
assigns x;
ensures x == \old(x) + 1;
*/
func inc(x int) int {
	x = x + 1
	return x
}

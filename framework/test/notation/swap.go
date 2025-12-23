package notation

/*
@
assigns a, b;
ensures (a == \old(b)) && (b == \old(a));
*/
func swap(a, b int) (int, int) {
	return b, a
}

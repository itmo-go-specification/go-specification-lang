package math

/*@
requires (a >= 0) && (b >= 0);
ensures res >= 0;
assigns \nothing;
*/
func GCD(a, b int) (res int) {
	if b == 0 {
		return a
	}
	return GCD(b, a%b)
}

/*@
requires (a >= 0) && (b >= 0);
ensures res >= 0;
assigns \nothing;
*/
func GCDIterative(a, b int) (res int) {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

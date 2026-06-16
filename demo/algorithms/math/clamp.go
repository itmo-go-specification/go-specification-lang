package math

/*@
requires lo <= hi;
ensures res >= lo;
ensures res <= hi;
assigns \nothing;
*/
func Clamp(x, lo, hi int) (res int) {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

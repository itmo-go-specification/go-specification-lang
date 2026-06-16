package math

/*@
requires \true;
ensures res >= 0;
assigns \nothing;
*/
func Abs(x int) (res int) {
	if x < 0 {
		return -x
	}
	return x
}

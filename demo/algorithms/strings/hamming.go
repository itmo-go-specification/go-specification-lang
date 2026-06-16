package strings

/*@
requires len(a) == len(b);
ensures res >= 0;
assigns \nothing;
*/
func Hamming(a, b string) (res int) {
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			res++
		}
	}
	return res
}

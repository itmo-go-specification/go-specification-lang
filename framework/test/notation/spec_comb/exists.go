package spec_comb

/*
@
ensures \exists a int ; a == true;
*/
func HasZero(a []int, n int) bool {
	for i := 0; i < n; i++ {
		if a[i] == 0 {
			return true
		}
	}
	return false
}

package spec_comb

/*
@
requires \forall i int. 0 <= i < n-1 => a[i] <= a[i+1]
*/
func IsSorted(a []int, n int) bool {
	for i := 0; i < n-1; i++ {
		if a[i] > a[i+1] {
			return false
		}
	}
	return true
}

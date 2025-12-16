package notation

/*@
assigns \nothing;
ensures bound: 0 <= \result <= n;
*/
func count(a []int, n int, v int) int {
	var counted = 0
	
	for i := 0; i < n; i++ {
		if a[i] == v {
			counted += 1
		}
	}
	return counted
}

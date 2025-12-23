package notation

/*
@
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

/*
@
requires 0 <= n;
assigns \nothing;
ensures bound: (0 <= \result) && (\result <= n);
*/
func countPrefix(a []int, n int, v int) int {
	c := 0
	for i := 0; i < n; i++ {
		if a[i] == v {
			c++
		}
	}
	return c
}

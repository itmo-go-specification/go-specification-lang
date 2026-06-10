package acsl

/*
@
axiomatic Count

	{
		logic integer
		Count(value_type* a, integer n, value_type v) =
		n <= 0 ? 0 : Count(a, n-1, v) + (a[n-1] == v ? 1 : 0);
	}

requires valid: \valid_read(a + (0..n-1));
assigns \nothing;
ensures bound: 0 <= \result <= n;
ensures count: \result == Count(a, n, v);
*/
func count(a []int, n int, v int) int {
	var counted = 0
	/*@
	loop invariant bound: 0 <= i <= n;
	loop invariant bound: 0 <= counted <= i;
	loop invariant count: counted == Count(a, i, v);
	loop assigns i, counted;
	loop variant n-i;
	*/
	for i := 0; i < n; i++ {
		if a[i] == v {
			counted += 1
		}
	}
	return counted
}

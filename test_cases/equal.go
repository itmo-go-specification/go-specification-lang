package acsl

/*@
axiomatic Equal
{
	predicate
	Equal{K,L}(value_type* a, integer n, value_type* b) =
		\forall integer i; 0 <= i < n ==> \at(a[i],K) == \at(b[i],L);
}

requires valid: \valid_read(a + (0..n-1));
requires valid: \valid_read(b + (0..n-1));
assigns \nothing;
ensures result: \result <==> Equal{Here,Here}(a, n, b);
*/
func equal(a []int, b []int, n int) bool {
	for i := 0; i < n; i++ {
		if a[i] != b[i] { return false }
	}
	return true
}

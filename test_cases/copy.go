package acsl

/*
@
axiomatic Equal

	{
		predicate
		Equal{K,L}(value_type* a, integer n, value_type* b) =
			\forall integer i; 0 <= i < n ==> \at(a[i],K) == \at(b[i],L);
	}

axiomatic Unchanged

	{
		predicate
		Unchanged{K,L}(value_type* a, integer m, integer n) =
			\forall integer i; m <= i < n ==> \at(a[i],K) == \at(a[i],L);
	}

requires valid: \valid_read(a + (0..n-1));
requires valid: \valid(b + (0..n-1));
requires sep: \separated(a + (0..n-1), b);
assigns b[0..n-1];
ensures equal: Equal{Old,Here}(a, n, b);
*/
func copy(a []int, b []int, n int) {
	/*@
	loop invariant bound: 0 <= i <= n;
	loop invariant equal: Equal{Pre,Here}(a, i, b);
	loop invariant unchanged: Unchanged{Pre,Here}(a, i, n);
	loop assigns i, b[0..n-1];
	loop variant n-i;
	*/
	for i := 0; i < n; i++ {
		b[i] = a[i]
	}
}

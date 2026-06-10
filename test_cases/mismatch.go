package acsl

/*
@
axiomatic Equal

	{
		predicate
		Equal{K,L}(value_type* a, integer n, value_type* b) =
			\forall integer i; 0 <= i < n ==> \at(a[i],K) == \at(b[i],L);
	}

requires valid: \valid_read(a + (0..n-1));
requires valid: \valid_read(b + (0..n-1));
assigns \nothing;
ensures result: 0 <= \result <= n;

behavior all_equal:

	assumes Equal{Here,Here}(a, n, b);
	assigns \nothing;
	ensures result: \result == n;

behavior some_not_equal:

	assumes !Equal{Here,Here}(a, n, b);
	assigns \nothing;
	ensures bound: 0 <= \result < n;
	ensures result: a[\result] != b[\result];
	ensures first: Equal{Here,Here}(a, \result, b);

complete behaviors;
disjoint behaviors;
*/
func mismatch(a []int, b []int, n int) int {
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

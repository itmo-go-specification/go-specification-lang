package acsl

/*@
axiomatic Replace
{
	predicate
	Replace{K,L}(value_type* a, integer n, value_type* b, value_type v, value_type w) =
		\forall integer i; 0 <= i < n ==>
		\let ai = \at(a[i],K);
		\let bi = \at(b[i],L);
		(ai == v ==> bi == w) && (ai != v ==> bi == ai);

	predicate
	Replace{K,L}(value_type* a, integer n, value_type v, value_type w) =
		Replace{K,L}(a, n, a, v, w);
}

axiomatic Unchanged
{
	predicate
	Unchanged{K,L}(value_type* a, integer m, integer n) =
		\forall integer i; m <= i < n ==> \at(a[i],K) == \at(a[i],L);
}

requires valid: \valid(a + (0..n-1));
assigns a[0..n-1];
ensures replace: Replace{Old,Here}(a, n, v, w);
*/
func replace(a []int, n int, v int, w int) {
	/*@
	loop invariant bounds: 0 <= i <= n;
	loop invariant replace: Replace{Pre,Here}(a, i, v, w);
	loop invariant unchanged: Unchanged{Pre,Here}(a, i, n);
	loop assigns i, a[0..n-1];
	loop variant n-i;
	*/
	for i := range n {
		if a[i] == v {
			a[i] = w;
		}
	}
}

package acsl

/*@
axiomatic AllSomeNot
{
	predicate
	AllEqual(value_type* a, integer m, integer n, value_type v) =
		\forall integer i; m <= i < n ==> a[i] == v;
}

requires valid: \valid(a + (0..n-1));
assigns a[0..n-1];
ensures constant: AllEqual(a, 0, n, v);
*/
func fill(a []int, n int, v int) {
	/*@
	loop invariant bound: 0 <= i <= n;
	loop invariant constant: AllEqual(a, i, v);
	loop assigns i, a[0..n-1];
	loop variant n-i;
	*/
	for i := 0; i < n; i++ {
		a[i] = v;
	}
}

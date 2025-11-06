package acsl

/*@
axiomatic ArrayBounds
{
	predicate
	LowerBound{L}(value_type* a, integer m, integer n, value_type v) =
		\forall integer i; m <= i < n ==> v <= a[i];

	predicate
	StrictUpperBound{L}(value_type* a, integer m, integer n, value_type v) =
		\forall integer i; m <= i < n ==> a[i] < v;
}

requires valid: \valid_read(a + (0..n-1));
requires increasing: Increasing(a, n);
assigns \nothing;
ensures result: 0 <= \result <= n;
ensures left: StrictUpperBound(a, 0, \result, v);
ensures right: LowerBound(a, \result, n, v);
*/
func lower_bound(a []int, n int, v int) int {
	var left = 0;
	var right = n;

	/*@
	loop invariant bound: 0 <= left <= right <= n;
	loop invariant left: StrictUpperBound(a, 0, left, v);
	loop invariant right: LowerBound(a, right, n, v);
	loop assigns left, right;
	loop variant right - left;
	*/
	for ;left < right; {
		var middle = left + (right - left) / 2
		if (a[middle] < v) {
			left = middle + 1;
		} else {
			right = middle;
		}
	}
	return left;
}

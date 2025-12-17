package notation

/*@
requires valid: \valid_read(a + (0..n-1));
requires increasing: Increasing(a, n);
assigns \nothing;
ensures result: 0 <= \result <= n;
ensures left: StrictUpperBound(a, 0, \result, v);
ensures right: LowerBound(a, \result, n, v);
*/
func lower_bound(a []int, n int, v int) int {
	
	/*@
	loop invariant bound: 0 <= left <= right <= n;
	loop invariant left: StrictUpperBound(a, 0, left, v);
	loop invariant right: LowerBound(a, right, n, v);
	loop assigns left, right;
	loop variant right - left;
	*/
	var left = 0;
	var right = n;
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
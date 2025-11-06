package acsl

/*@
requires valid: \valid_read(a + (0..n-1));
assigns \nothing;
ensures result: 0 <= \result <= n;

behavior empty:
	assumes n == 0;
	assigns \nothing;
	ensures result: \result == 0;
behavior not_empty:
	assumes 0 < n;
	assigns \nothing;
	ensures result: 0 <= \result < n;
	ensures upper: \forall integer i; 0 <= i < n ==> a[i] <= a[\result];
	ensures first: \forall integer i; 0 <= i < \result ==> a[i] < a[\result];
complete behaviors;
disjoint behaviors;
*/
func max_element(a []int, n int) int {
	if n == 0 {
		return 0
	}
	max := 0

	/*@
	loop invariant bound: 0 <= i <= n;
	loop invariant max: 0 <= max < n;
	loop invariant upper: \forall integer k; 0 <= k < i ==> a[k] <= a[max];
	loop invariant first: \forall integer k; 0 <= k < max ==> a[k] < a[max];
	loop assigns max, i;
	loop variant n-i;
	*/
	for i := 1; i < n; i++ {
		if a[i] > a[max] {
			max = i
		}
	}
	return max
}

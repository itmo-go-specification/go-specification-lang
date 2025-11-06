package acsl

/*@
requires \valid_read(a + (0..n-1));
assigns \nothing;
ensures 0 <= \result <= n;

behavior some:

	assumes \exists integer i; 0 <= i < n && a[i] == v;
	assigns \nothing;
	ensures 0 <= \result < n;
	ensures a[\result] == v;
	ensures \forall integer i; 0 <= i < \result ==> a[i] != v;

behavior none:

	assumes \forall integer i; 0 <= i < n ==> a[i] != v;
	assigns \nothing;
	ensures \result == n;

complete behaviors;
disjoint behaviors;
*/
func find(a []int, n int, v int) int {
	for i := 0; i < n; i++ {
		if a[i] == v {
			return i
		}
	}
	return n
}
